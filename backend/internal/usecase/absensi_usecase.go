// file: internal/usecase/absensi_usecase.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"daarulilmi-presence/internal/domain" // Ganti dengan nama modul Anda

	"github.com/skip2/go-qrcode"
)

// Definisikan secret key untuk payload QR, agar bisa divalidasi nanti
const qrSecretKey = "daarulilmi-presence"

// --- BUAT KONTRAK UNTUK REPOSITORY ABSENSI ---
type AbsensiRepository interface {
	RecordAttendance(ctx context.Context, username, status string) error
	GetAttendanceForUser(ctx context.Context, username string) ([]domain.LogAbsensi, error)
	GetAllLeaveRequests(ctx context.Context) ([]domain.PengajuanIzinLengkap, error)
	GetTotalSiswa(ctx context.Context) (int, error)
	GetTodaysAttendanceAndLeave(ctx context.Context) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error)
	GetTodaysAttendance(ctx context.Context) ([]domain.LogAbsensi, error)
	GetTodaysLeave(ctx context.Context) ([]domain.PengajuanIzinLengkap, error)
	CreateManualAttendance(ctx context.Context, data *domain.KehadiranManual) error
	CreateBatchManualAttendance(ctx context.Context, data []domain.KehadiranManual) error
	DeleteAttendance(ctx context.Context, rowNumber int) error
	UpdateAttendance(ctx context.Context, rowNumber int, data *domain.KehadiranManual) error
	GetAttendanceByRow(ctx context.Context, rowNumber int) (*domain.LogAbsensi, error)
	GetAttendanceByDate(ctx context.Context, date string) ([]domain.LogAbsensi, error)
	GetLeaveByDate(ctx context.Context, date string) ([]domain.PengajuanIzinLengkap, error)
	GetHolidays(ctx context.Context) (map[string]bool, error)
	GetAllLogsInMonth(ctx context.Context, year, month int) ([]domain.LogAbsensi, error)
	GetLogsByDateRange(ctx context.Context, startDate, endDate string) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error)
	FindTodaysAttendanceLog(ctx context.Context, nisn string) (*domain.LogAbsensi, error)
	UpdateClockOut(ctx context.Context, rowNumber int, clockOutTime string) error
	GetAllLogsForUserInMonth(ctx context.Context, nisn string, year, month int) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error)
}

type SiswaRepository interface {
	FindAll(ctx context.Context) ([]domain.Siswa, error)
	FindByNISN(ctx context.Context, nisn string) (*domain.Siswa, error)
	Save(ctx context.Context, siswa *domain.Siswa) error
	Update(ctx context.Context, nisn string, siswa *domain.Siswa) error
	Delete(ctx context.Context, nisn string) error
}

type absensiUsecase struct {
	absensiRepo AbsensiRepository
	siswaRepo   SiswaRepository
	userRepo    UserRepository
}

// NewAbsensiUsecase adalah "pabrik" untuk usecase absensi
func NewAbsensiUsecase(absensiRepo AbsensiRepository, siswaRepo SiswaRepository, userRepo UserRepository) domain.AbsensiUsecase {
	return &absensiUsecase{
		absensiRepo: absensiRepo,
		siswaRepo:   siswaRepo,
		userRepo:    userRepo,
	}
}

func (uc *absensiUsecase) DeleteAttendance(ctx context.Context, rowNumber int) error {
	return uc.absensiRepo.DeleteAttendance(ctx, rowNumber)
}

func (uc *absensiUsecase) UpdateAttendance(ctx context.Context, rowNumber int, data *domain.KehadiranManual) error {
	// Di sini bisa ditambahkan validasi data sebelum dikirim ke repository
	return uc.absensiRepo.UpdateAttendance(ctx, rowNumber, data)
}

func (uc *absensiUsecase) GetAttendanceByRow(ctx context.Context, rowNumber int) (*domain.LogAbsensi, error) {
	return uc.absensiRepo.GetAttendanceByRow(ctx, rowNumber)
}

func (uc *absensiUsecase) GetSmartDashboardData(ctx context.Context, username string, dateStr string) (*domain.SmartDashboardData, error) {
	// Ambil Nama Lengkap user yang login terlebih dahulu
	loggedInUser, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		log.Printf("Peringatan: Gagal mencari user '%s' untuk nama lengkap: %v", username, err)
	}
	var namaLengkapUser string
	if loggedInUser != nil {
		namaLengkapUser = loggedInUser.NamaLengkap
	}
	if namaLengkapUser == "" {
		namaLengkapUser = username // Fallback jika nama lengkap tidak ada
	}

	// Jika tanggal kosong, gunakan hari ini
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("format tanggal salah: %v", err)
	}

	// Cek hari libur
	holidays, err := uc.absensiRepo.GetHolidays(ctx)
	if err != nil {
		return nil, err
	}
	if holidays[dateStr] {
		return &domain.SmartDashboardData{NamaLengkapUser: namaLengkapUser, IsHoliday: true, HolidayDescription: "Tanggal Merah"}, nil
	}
	if targetDate.Weekday() == time.Saturday || targetDate.Weekday() == time.Sunday {
		return &domain.SmartDashboardData{NamaLengkapUser: namaLengkapUser, IsHoliday: true, HolidayDescription: "Akhir Pekan"}, nil
	}

	// Ambil semua data mentah
	allSiswa, err := uc.siswaRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	hadirList, err := uc.absensiRepo.GetAttendanceByDate(ctx, dateStr)
	if err != nil {
		return nil, err
	}

	izinList, err := uc.absensiRepo.GetLeaveByDate(ctx, dateStr)
	if err != nil {
		return nil, err
	}

	// --- LOGIKA BARU: Proses data manual & izin terlebih dahulu ---
	manualStatusMap := make(map[string]domain.LogAbsensi)
	izinMap := make(map[string]domain.PengajuanIzinLengkap)

	// Proses data izin dari GForm
	for _, log := range izinList {
		izinMap[log.SiswaNISN] = log
	}

	// Proses SEMUA log absensi (dari QR dan manual), cari yang paling baru untuk setiap siswa
	for _, log := range hadirList {
		nisn := log.Username
		if existingLog, ok := manualStatusMap[nisn]; !ok || log.Timestamp > existingLog.Timestamp {
			manualStatusMap[nisn] = log
		}
	}

	var daftarStatusSiswa []domain.SiswaStatus
	totalHadir := 0
	totalIzinSakit := 0
	isAfter6PM := time.Now().Hour() >= 18 && dateStr == time.Now().Format("2006-01-02")

	for _, siswa := range allSiswa {
		statusSiswa := domain.SiswaStatus{
			NISN:        siswa.NISN,
			NamaLengkap: siswa.NamaLengkap,
			Kelas:       siswa.Kelas,
		}

		// --- PRIORITAS BARU ---
		// 1. Cek apakah ada status manual (termasuk Hadir dari QR)
		if dataManual, found := manualStatusMap[siswa.NISN]; found {
			statusSiswa.Status = dataManual.Status
			statusSiswa.RowNumber = dataManual.RowNumber
			parts := strings.Split(dataManual.Timestamp, " ")
			if len(parts) == 2 {
				statusSiswa.Keterangan = fmt.Sprintf("Dicatat pukul %s", parts[1])
			} else {
				statusSiswa.Keterangan = "Tercatat"
			}

			// 2. Jika tidak ada, baru cek dari GForm Izin
		} else if dataIzin, found := izinMap[siswa.NISN]; found {
			statusSiswa.Status = dataIzin.JenisIzin
			statusSiswa.RowNumber = dataIzin.RowNumber
			statusSiswa.Keterangan = "Surat/Form diterima"
			// 3. Jika tidak ada sama sekali, tentukan statusnya
		} else {
			if isAfter6PM {
				statusSiswa.Status = "Alpa"
				statusSiswa.Keterangan = "Tidak ada konfirmasi"
			} else {
				statusSiswa.Status = "Belum Ada Kabar"
				statusSiswa.Keterangan = "-"
			}
		}
		daftarStatusSiswa = append(daftarStatusSiswa, statusSiswa)
	}

	// Hitung ulang total statistik berdasarkan status final
	for _, s := range daftarStatusSiswa {
		switch strings.ToLower(s.Status) {
		case "hadir":
			totalHadir++
		case "izin", "sakit":
			totalIzinSakit++
		}
	}

	return &domain.SmartDashboardData{
		NamaLengkapUser:    namaLengkapUser,
		IsHoliday:          false,
		TotalSiswa:         len(allSiswa),
		TotalHadir:         totalHadir,
		TotalIzin:          totalIzinSakit,
		TotalBelumAdaKabar: len(allSiswa) - totalHadir - totalIzinSakit,
		DaftarStatusSiswa:  daftarStatusSiswa,
	}, nil
}

// GenerateQR adalah implementasi logika pembuatan QR code
func (uc *absensiUsecase) GenerateQR(ctx context.Context, qrType string) ([]byte, error) {
	// Buat payload: berisi secret key dan timestamp saat ini.
	// Ini untuk memastikan QR code valid dan tidak bisa digunakan berulang kali di lain hari.
	payload := fmt.Sprintf("%s:%d:%s", qrSecretKey, time.Now().Unix(), qrType)

	// Generate QR code dari payload menjadi gambar PNG dengan ukuran 256x256 pixel.
	// qrcode.Encode akan mengembalikan byte slice dari gambar PNG.
	var png []byte
	png, err := qrcode.Encode(payload, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return png, nil
}

func (uc *absensiUsecase) VerifyAndRecordScan(ctx context.Context, qrData string, username string) (string, error) {
	// 1. Pisahkan payload: "secret-key:timestamp"
	parts := strings.Split(qrData, ":")
	if len(parts) != 3 {
		return "", errors.New("QR code tidak valid: format salah")
	}

	// 2. Validasi secret key
	if parts[0] != qrSecretKey {
		return "", errors.New("QR code tidak valid: kunci tidak cocok")
	}

	// 3. Validasi timestamp (misal: QR hanya valid selama 60 detik)
	qrTimestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", errors.New("QR code tidak valid: timestamp rusak")
	}

	if time.Now().Unix()-qrTimestamp > 60 {
		return "", errors.New("QR code sudah kedaluwarsa")
	}

	qrType := parts[2]
	siswa, err := uc.siswaRepo.FindByNISN(ctx, username)
	if err != nil || siswa == nil {
		return "", errors.New("data siswa tidak ditemukan")
	}

	switch qrType {
	case "masuk":
		// Cek apakah siswa sudah absen masuk hari ini
		existingLog, _ := uc.absensiRepo.FindTodaysAttendanceLog(ctx, siswa.NISN)
		if existingLog != nil {
			return "", errors.New("Anda sudah melakukan absensi masuk hari ini")
		}

		// Buat data absensi baru
		data := &domain.KehadiranManual{
			NISN:        siswa.NISN,
			NamaSiswa:   siswa.NamaLengkap,
			Status:      "Hadir",
			Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
			DicatatOleh: "Sistem QR",
		}
		err = uc.absensiRepo.CreateManualAttendance(ctx, data)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Absensi Masuk untuk %s berhasil!", siswa.NamaLengkap), nil

	case "pulang":
		existingLog, err := uc.absensiRepo.FindTodaysAttendanceLog(ctx, siswa.NISN)
		if err != nil {
			return "", err
		}
		if existingLog == nil {
			return "", errors.New("Anda belum melakukan absensi masuk hari ini")
		}
		if existingLog.TimestampPulang != "" {
			return "", errors.New("Anda sudah melakukan absensi pulang hari ini")
		}

		// Update data absensi yang sudah ada
		clockOutTime := time.Now().Format("15:04:05")
		err = uc.absensiRepo.UpdateClockOut(ctx, existingLog.RowNumber, clockOutTime)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Absensi Pulang untuk %s berhasil!", siswa.NamaLengkap), nil
	}

	return "", errors.New("tipe QR code tidak dikenal")

}

func (uc *absensiUsecase) GetAttendanceForUser(ctx context.Context, username string) ([]domain.LogAbsensi, error) {
	// Logika bisnisnya sederhana, hanya meneruskan permintaan ke repository
	return uc.absensiRepo.GetAttendanceForUser(ctx, username)
}

func (uc *absensiUsecase) GetAllLeaveRequests(ctx context.Context) ([]domain.PengajuanIzinLengkap, error) {
	return uc.absensiRepo.GetAllLeaveRequests(ctx)
}

func (uc *absensiUsecase) GetDashboardData(ctx context.Context, username string) (*domain.DashboardData, error) {
	// Ambil data user untuk dapat nama lengkap
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Ambil data absensi
	totalSiswa, err := uc.absensiRepo.GetTotalSiswa(ctx)
	if err != nil {
		return nil, err
	}

	logHadir, logIzin, err := uc.absensiRepo.GetTodaysAttendanceAndLeave(ctx)
	if err != nil {
		return nil, err
	}

	totalHadir := len(logHadir)
	totalIzin := len(logIzin)

	data := &domain.DashboardData{
		NamaLengkap:         user.NamaLengkap,
		Username:            username,
		TotalSiswa:          totalSiswa,
		TotalHadirHariIni:   totalHadir,
		TotalIzinHariIni:    totalIzin,
		TotalBelumAdaKabar:  totalSiswa - totalHadir - totalIzin,
		LogKehadiranHariIni: logHadir,
		SiswaIzinHariIni:    logIzin,
	}

	return data, nil
}

func (uc *absensiUsecase) GetTodaysAttendanceAndLeave(ctx context.Context) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error) {
	// 1. Ambil data mentah (repository sekarang mengembalikan slice, bukan map)
	hadirList, err := uc.absensiRepo.GetTodaysAttendance(ctx)
	if err != nil {
		return nil, nil, err
	}
	izinList, err := uc.absensiRepo.GetTodaysLeave(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 2. Ambil daftar semua siswa untuk mendapatkan nama lengkap
	allSiswa, err := uc.siswaRepo.FindAll(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Buat "kamus" pencocokan NISN -> Nama
	nisnToNamaMap := make(map[string]string)
	for _, siswa := range allSiswa {
		nisnToNamaMap[siswa.NISN] = siswa.NamaLengkap
	}

	// 3. Proses dan gabungkan data, sekarang dengan nama lengkap
	for i := range hadirList {
		hadirList[i].NamaLengkap = nisnToNamaMap[hadirList[i].Username]
	}

	for i := range izinList {
		izinList[i].NamaLengkap = nisnToNamaMap[izinList[i].SiswaNISN]
	}

	return hadirList, izinList, nil
}

func (uc *absensiUsecase) CreateManualAttendance(ctx context.Context, data *domain.KehadiranManual) error {
	// Ambil NamaSiswa berdasarkan NISN terlebih dahulu
	siswa, err := uc.siswaRepo.FindByNISN(ctx, data.NISN)
	if err != nil || siswa == nil {
		return errors.New("NISN siswa tidak ditemukan")
	}
	data.NamaSiswa = siswa.NamaLengkap
	data.DicatatOleh = "Manual Wali Kelas"

	// --- LOGIKA BARU DIMULAI DI SINI ---
	// Cek apakah sudah ada log untuk siswa ini hari ini
	existingLog, err := uc.absensiRepo.FindTodaysAttendanceLog(ctx, data.NISN)
	if err != nil {
		// Jika error bukan karena tidak ditemukan, kembalikan error
		log.Printf("Error mencari log yang ada: %v", err)
		return err
	}

	// Jika log sudah ada, lakukan UPDATE
	if existingLog != nil {
		log.Printf("INFO: Log sudah ada di baris %d. Melakukan UPDATE.", existingLog.RowNumber)
		// Gunakan timestamp yang sudah ada, kecuali jika mau diubah
		if data.Timestamp == "" {
			data.Timestamp = existingLog.Timestamp
		}
		return uc.absensiRepo.UpdateAttendance(ctx, existingLog.RowNumber, data)
	}

	// Jika log belum ada, lakukan CREATE (buat baris baru)
	log.Printf("INFO: Log belum ada. Melakukan CREATE.")
	if data.Timestamp == "" {
		data.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}
	return uc.absensiRepo.CreateManualAttendance(ctx, data)
}

func (uc *absensiUsecase) CreateBatchManualAttendance(ctx context.Context, data []domain.KehadiranManual) error {
	// Lakukan perulangan untuk setiap data siswa yang dikirim dari frontend
	for i := range data {
		// Ambil NamaSiswa berdasarkan NISN
		siswa, err := uc.siswaRepo.FindByNISN(ctx, data[i].NISN)
		if err != nil || siswa == nil {
			// Jika satu siswa tidak ditemukan, kita bisa memilih untuk melanjutkan atau mengembalikan error
			// Di sini kita pilih untuk melanjutkan saja, tapi beri log
			log.Printf("WARNING: NISN siswa %s tidak ditemukan, data tidak dicatat.", data[i].NISN)
			continue // Lanjutkan ke siswa berikutnya
		}
		// Sisipkan data yang hilang
		data[i].NamaSiswa = siswa.NamaLengkap
		data[i].DicatatOleh = "Manual Wali Kelas (Massal)"
	}

	// Kirim data yang sudah diperkaya ke repository
	return uc.absensiRepo.CreateBatchManualAttendance(ctx, data)
}

func (uc *absensiUsecase) GetMonthlyStats(ctx context.Context, year, month int) (*domain.StatistikData, error) {
	allLogs, err := uc.absensiRepo.GetAllLogsInMonth(ctx, year, month)
	if err != nil {
		return nil, err
	}

	stats := &domain.StatistikData{}
	for _, log := range allLogs {
		switch strings.ToLower(log.Status) {
		case "hadir":
			stats.TotalHadir++
		case "izin":
			stats.TotalIzin++
		case "sakit":
			stats.TotalSakit++
		case "alpa":
			stats.TotalAlpa++
		}
	}
	return stats, nil
}

func (uc *absensiUsecase) GetRekapByDateRange(ctx context.Context, startDate, endDate string) ([]domain.RekapSiswa, error) {
	allSiswa, err := uc.siswaRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	hadirLogs, izinLogs, err := uc.absensiRepo.GetLogsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	holidays, err := uc.absensiRepo.GetHolidays(ctx)
	if err != nil {
		return nil, err
	}

	// --- LOGIKA BARU: HITUNG HARI KERJA ---
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	totalHariKerja := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		isWeekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
		isHoliday := holidays[dateStr]
		if !isWeekend && !isHoliday {
			totalHariKerja++
		}
	}
	// --- AKHIR LOGIKA BARU ---

	rekapMap := make(map[string]*domain.RekapSiswa)
	for _, siswa := range allSiswa {
		rekapMap[siswa.NISN] = &domain.RekapSiswa{
			NISN:        siswa.NISN,
			NamaLengkap: siswa.NamaLengkap,
		}
	}

	for _, log := range hadirLogs {
		if rekap, ok := rekapMap[log.Username]; ok {
			rekap.Hadir++
		}
	}
	for _, log := range izinLogs {
		if rekap, ok := rekapMap[log.SiswaNISN]; ok {
			status := strings.ToLower(log.JenisIzin)
			if status == "sakit" {
				rekap.Sakit++
			} else {
				rekap.Izin++
			}
		}
	}

	var rekapList []domain.RekapSiswa
	for _, rekap := range rekapMap {
		// Hitung Alpa
		totalAbsenTercatat := rekap.Hadir + rekap.Izin + rekap.Sakit
		rekap.Alpa = totalHariKerja - totalAbsenTercatat
		if rekap.Alpa < 0 {
			rekap.Alpa = 0
		} // Pastikan tidak negatif

		rekapList = append(rekapList, *rekap)
	}

	return rekapList, nil
}

// --- TAMBAHKAN FUNGSI BARU INI ---
func (uc *absensiUsecase) GetPortalDashboardData(ctx context.Context, username string, year, month int) (*domain.PortalDashboardData, error) {
	log.Println("--- [USECASE START] GetPortalDashboardData ---")
	log.Printf("Mencari data untuk user: %s, Tahun: %d, Bulan: %d", username, year, month)

	// Langkah 1: Dapatkan data user (wali murid) dan NISN anaknya
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil || user == nil {
		log.Printf("ERROR saat mencari user: %v", err)
		return nil, errors.New("pengguna tidak ditemukan")
	}
	log.Printf("User ditemukan. Nama: %s, Role: %s, SiswaNISN: %s", user.NamaLengkap, user.Role, user.SiswaNISN)

	if user.SiswaNISN == "" {
		log.Println("WARNING: Akun ini tidak terhubung dengan data siswa (SiswaNISN kosong).")
		return nil, errors.New("akun ini tidak terhubung dengan data siswa")
	}

	// Langkah 2: Dapatkan data siswa terkait
	siswa, err := uc.siswaRepo.FindByNISN(ctx, user.SiswaNISN)
	if err != nil || siswa == nil {
		log.Printf("ERROR saat mencari siswa dengan NISN %s: %v", user.SiswaNISN, err)
		return nil, errors.New("data siswa terkait tidak ditemukan")
	}
	log.Printf("Siswa terkait ditemukan: %s", siswa.NamaLengkap)

	hadirLogs, izinLogs, err := uc.absensiRepo.GetAllLogsForUserInMonth(ctx, siswa.NISN, year, month)
	if err != nil {
		log.Printf("ERROR saat mengambil log bulan ini: %v", err)
		return nil, err
	}
	log.Printf("Data ditemukan: %d log kehadiran, %d log izin.", len(hadirLogs), len(izinLogs))

	holidays, _ := uc.absensiRepo.GetHolidays(ctx)

	events := []domain.CalendarEvent{}
	stats := &domain.StatistikData{}

	for _, log := range hadirLogs {
		// title := log.Status
		// color := "#198754" // Warna hijau untuk hadir

		// 1. Selalu buat event untuk "Hadir"
		events = append(events, domain.CalendarEvent{
			Title: "Hadir",
			Start: strings.Split(log.Timestamp, " ")[0],
			Color: "#198754", // Warna hijau
		})

		// Cek apakah ada data jam pulang
		if log.TimestampPulang != "" {
			title := "Pulang"
			// Coba format jamnya menjadi HH:MM
			t, err := time.Parse("15:04:05", log.TimestampPulang)
			if err == nil {
				title = fmt.Sprintf("Pulang (%s)", t.Format("15:04"))
			}
			events = append(events, domain.CalendarEvent{
				Title: title,
				Start: strings.Split(log.Timestamp, " ")[0],
				Color: "#0d6efd",
			})
		}
		stats.TotalHadir++
	}
	for _, log := range izinLogs {
		events = append(events, domain.CalendarEvent{
			Title: log.JenisIzin, Start: log.TanggalMulai, Color: "#ffc107",
		})
		if strings.ToLower(log.JenisIzin) == "sakit" {
			stats.TotalSakit++
		} else {
			stats.TotalIzin++
		}
	}

	// Tambahkan hari libur ke event kalender
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

	for d := firstDayOfMonth; !d.After(lastDayOfMonth); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if holidays[dateStr] {
			events = append(events, domain.CalendarEvent{Title: "Libur Nasional", Start: dateStr, Color: "#6c757d"})
		} else if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			events = append(events, domain.CalendarEvent{Title: "Akhir Pekan", Start: dateStr, Color: "#adb5bd"})
		}
	}

	log.Println("--- [USECASE END] Berhasil mengumpulkan data. ---")

	return &domain.PortalDashboardData{
		Siswa:          siswa,
		Events:         events,
		StatistikBulan: stats,
	}, nil
}
