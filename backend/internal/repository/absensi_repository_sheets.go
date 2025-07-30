// file: internal/repository/absensi_repository_sheets.go
package repository

import (
	"context" // Tambahkan import fmt
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"daarulilmi-presence/internal/domain" // Pastikan nama modul sudah benar
	"daarulilmi-presence/internal/usecase"

	"google.golang.org/api/sheets/v4"
)

type absensiRepository struct {
	db            *sheets.Service
	spreadsheetId string
}

func NewAbsensiRepository(db *sheets.Service, spreadsheetId string) usecase.AbsensiRepository {
	return &absensiRepository{db, spreadsheetId}
}

func parseFlexibleDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02", // YYYY-MM-DD
		"1/2/2006",   // M/D/YYYY (Format umum GForm)
		"01/02/2006", // MM/DD/YYYY
		"2/1/2006",   // D/M/YYYY
		"02/01/2006", // DD/MM/YYYY
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("tidak dapat mem-parsing tanggal: %s", dateStr)
}

func (r *absensiRepository) GetAttendanceByDate(ctx context.Context, date string) ([]domain.LogAbsensi, error) {
	var hadirList []domain.LogAbsensi
	readRange := "LogAbsensi!A2:E"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	for i, row := range resp.Values {
		timestamp := getStringFromCellByIndex(row, 1) // Kolom B
		if strings.HasPrefix(timestamp, date) {
			hadirList = append(hadirList, domain.LogAbsensi{
				RowNumber: i + 2,
				Timestamp: timestamp,
				Username:  getStringFromCellByIndex(row, 2),
				Status:    getStringFromCellByIndex(row, 4),
			})
		}
	}
	return hadirList, nil
}

// --- FUNGSI INI DIPERBAIKI (untuk menulis data baru dengan benar) ---
func (r *absensiRepository) RecordAttendance(ctx context.Context, username, status string) error {
	writeRange := "LogAbsensi"
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var values [][]interface{}
	// Sesuaikan urutan agar cocok dengan sheet: A(kosong), B(Timestamp), C(NIS), D(NamaSiswa-kosong), E(Status)
	row := []interface{}{"", timestamp, username, "", status}
	values = append(values, row)

	valueRange := &sheets.ValueRange{Values: values}

	_, err := r.db.Spreadsheets.Values.Append(r.spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("Gagal menyimpan log absensi ke sheet: %v", err)
		return err
	}
	log.Printf("Absensi untuk user [%s] berhasil dicatat.", username)
	return nil
}

// --- FUNGSI INI DIPERBAIKI (untuk membaca data dengan benar) ---
func (r *absensiRepository) GetAttendanceForUser(ctx context.Context, username string) ([]domain.LogAbsensi, error) {
	// 1. Dapatkan NISN Siswa (logika ini tetap sama)
	userRange := "DataPengguna!A2:D"
	userResp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, userRange).Do()
	if err != nil {
		return nil, err
	}

	var siswaNISN string
	for _, row := range userResp.Values {
		if len(row) > 3 && row[0].(string) == username {
			siswaNISN = row[3].(string)
			break
		}
	}
	if siswaNISN == "" {
		siswaNISN = username
	}

	// --- INI BAGIAN BARUNYA ---
	var results []domain.LogAbsensi

	// 2. Baca dari LogAbsensi
	logRange := "LogAbsensi!A2:E"
	logResp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, logRange).Do()
	if err != nil {
		return nil, err
	}

	for _, row := range logResp.Values {
		if len(row) >= 5 && row[2].(string) == siswaNISN {
			results = append(results, domain.LogAbsensi{
				Timestamp: row[1].(string), // Kolom B
				Username:  row[2].(string), // Kolom C (NIS)
				Status:    row[4].(string), // Kolom E
			})
		}
	}

	// 3. Baca dari PengajuanIzin (dari Google Form)
	// Asumsi: Kolom C adalah NISN, Kolom E adalah Jenis Izin, Kolom F adalah Tanggal Mulai
	izinRange := "PengajuanIzin!A2:F"
	izinResp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, izinRange).Do()
	if err != nil {
		return nil, err
	}

	for _, row := range izinResp.Values {
		// Pastikan NISN di Kolom C cocok
		if len(row) >= 6 && row[2].(string) == siswaNISN {
			results = append(results, domain.LogAbsensi{
				Timestamp: row[5].(string), // Kolom F (Tanggal Mulai)
				Username:  row[2].(string), // Kolom C (NIS)
				Status:    row[4].(string), // Kolom E (Jenis Izin, cth: "Sakit" atau "Izin")
			})
		}
	}

	return results, nil
}

// file: internal/repository/absensi_repository_sheets.go
func (r *absensiRepository) GetAllLeaveRequests(ctx context.Context) ([]domain.PengajuanIzinLengkap, error) {
	// === LANGKAH 1: Buat Kamus Pencocokan Nama ke NISN ===
	// Baca seluruh data dari sheet DataSiswa
	siswaRange := "DataSiswa!A2:B" // Asumsi Kolom A=NISN, B=NamaLengkap
	siswaResp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, siswaRange).Do()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca DataSiswa: %v", err)
	}

	// Buat map untuk pencocokan cepat: map[NamaLengkap] -> NISN
	namaToNisnMap := make(map[string]string)
	for _, row := range siswaResp.Values {
		if len(row) >= 2 {
			nisn := row[0].(string)
			namaLengkap := row[1].(string)
			namaToNisnMap[namaLengkap] = nisn
		}
	}

	// === LANGKAH 2: Baca Data dari Sheet Respons Form Izin ===
	// Kita baca semua kolom yang relevan
	izinRange := "PengajuanIzin!A:AF" // Baca dari kolom A sampai AF (Tindak Lanjut Wali Kelas)
	izinResp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, izinRange).Do()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca PengajuanIzin: %v", err)
	}

	var requests []domain.PengajuanIzinLengkap
	// Lewati baris pertama (header) dengan memulai loop dari indeks 1
	for _, row := range izinResp.Values[1:] {
		// Pastikan baris memiliki cukup kolom untuk dibaca
		if len(row) < 32 {
			continue
		}

		// Ambil data dari kolom yang sesuai
		timestamp := row[0].(string)      // Kolom A (indeks 0): Timestamp
		namaSiswa := row[1].(string)      // Kolom B (indeks 1): Nama Siswa/i
		jenisPengaduan := row[5].(string) // Kolom F (indeks 5): Izin Tidak Masuk karena
		tglMulai := row[13].(string)      // Kolom N (indeks 13): Hari dan tanggal

		// Cek status dari kolom Tindak Lanjut Wali Kelas
		status := "Menunggu" // Default status
		if len(row) > 31 && row[31].(string) != "" {
			status = row[31].(string) // Kolom AF (indeks 31)
		}

		// Cocokkan nama siswa dengan NISN dari kamus yang kita buat
		nisn, found := namaToNisnMap[namaSiswa]
		if !found {
			nisn = "N/A - Nama tidak ditemukan di DataSiswa" // Penanda jika nama tidak cocok
		}

		requests = append(requests, domain.PengajuanIzinLengkap{
			Timestamp:      timestamp,
			SiswaNISN:      nisn,
			JenisIzin:      jenisPengaduan,
			TanggalMulai:   tglMulai,
			TanggalSelesai: tglMulai, // Asumsi izin hanya 1 hari, karena form hanya ada 1 kolom tanggal
			Status:         status,
		})
	}

	return requests, nil
}

// --- FUNGSI BARU UNTUK MENGHITUNG TOTAL SISWA ---
func (r *absensiRepository) GetTotalSiswa(ctx context.Context) (int, error) {
	readRange := "DataSiswa!A2:A"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return 0, err
	}
	return len(resp.Values), nil
}

func (r *absensiRepository) GetAttendanceByRow(ctx context.Context, rowNumber int) (*domain.LogAbsensi, error) {
	readRange := fmt.Sprintf("LogAbsensi!A%d:E%d", rowNumber, rowNumber)
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Values) == 0 {
		return nil, errors.New("data absensi tidak ditemukan")
	}
	row := resp.Values[0]
	logEntry := &domain.LogAbsensi{
		RowNumber: rowNumber,
		Timestamp: getStringFromCellByIndex(row, 1),
		Username:  getStringFromCellByIndex(row, 2),
		Status:    getStringFromCellByIndex(row, 4),
	}
	return logEntry, nil
}

// --- FUNGSI BARU UNTUK MENGAMBIL ABSENSI & IZIN HARI INI ---
func (r *absensiRepository) GetTodaysAttendanceAndLeave(ctx context.Context) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error) {
	today := time.Now().Format("2006-01-02")
	var todaysAttendance []domain.LogAbsensi
	var todaysLeave []domain.PengajuanIzinLengkap

	// 1. Ambil data dari LogAbsensi
	logRange := "LogAbsensi!A2:E"
	logResp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, logRange).Do()
	if err != nil {
		return nil, nil, err
	}

	for _, row := range logResp.Values {
		if len(row) >= 5 && strings.HasPrefix(row[1].(string), today) {
			todaysAttendance = append(todaysAttendance, domain.LogAbsensi{
				Timestamp: row[1].(string),
				Username:  row[2].(string),
				Status:    row[4].(string),
			})
		}
	}

	// 2. Ambil data dari PengajuanIzin
	allLeaveRequests, err := r.GetAllLeaveRequests(ctx) // Kita panggil ulang fungsi yang sudah ada
	if err != nil {
		return nil, nil, err
	}

	for _, req := range allLeaveRequests {
		// Cek apakah hari ini berada di antara tanggal mulai dan selesai
		// (Ini adalah penyederhanaan, logika tanggal yang lebih kompleks mungkin diperlukan)
		if strings.HasPrefix(req.TanggalMulai, today) {
			todaysLeave = append(todaysLeave, req)
		}
	}

	return todaysAttendance, todaysLeave, nil
}

func (r *absensiRepository) GetTodaysAttendance(ctx context.Context) ([]domain.LogAbsensi, error) {
	today := time.Now().Format("2006-01-02")
	var hadirList []domain.LogAbsensi

	readRange := "LogAbsensi!A2:E"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	for i, row := range resp.Values {
		actualRowNumber := i + 2
		timestamp := getStringFromCellByIndex(row, 1) // PASTIKAN INI INDEKS 1 (KOLOM B)

		if strings.HasPrefix(timestamp, today) {
			hadirList = append(hadirList, domain.LogAbsensi{
				RowNumber: actualRowNumber,
				Timestamp: timestamp,
				Username:  getStringFromCellByIndex(row, 2), // Kolom C
				Status:    getStringFromCellByIndex(row, 4), // Kolom E
			})
		}
	}
	return hadirList, nil
}

func (r *absensiRepository) GetTodaysLeave(ctx context.Context) ([]domain.PengajuanIzinLengkap, error) {
	today := time.Now().Format("2006-01-02")
	var izinList []domain.PengajuanIzinLengkap

	readRange := "PengajuanIzin!A2:J"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Values) == 0 {
		return izinList, nil
	}

	for i, row := range resp.Values {
		actualRowNumber := i + 2
		tglMulai := getStringFromCellByIndex(row, 5) // Kolom F

		if tglMulai == today {
			izinList = append(izinList, domain.PengajuanIzinLengkap{
				RowNumber:      actualRowNumber,
				Timestamp:      getStringFromCellByIndex(row, 1),
				SiswaNISN:      getStringFromCellByIndex(row, 2),
				JenisIzin:      getStringFromCellByIndex(row, 4),
				TanggalMulai:   tglMulai,
				TanggalSelesai: getStringFromCellByIndex(row, 6),
				Status:         getStringFromCellByIndex(row, 9),
			})
		}
	}
	return izinList, nil
}

func (r *absensiRepository) CreateManualAttendance(ctx context.Context, data *domain.KehadiranManual) error {
	writeRange := "LogAbsensi"

	// Buat ID unik sederhana berbasis waktu
	logID := fmt.Sprintf("LOG-%d", time.Now().UnixNano())

	var values [][]interface{}
	// Sesuaikan urutan kolom: A(kosong), B(Timestamp), C(NISN), D(NamaSiswa-kosong), E(Status)
	row := []interface{}{
		logID,            // A: LogID
		data.Timestamp,   // B: Timestamp
		data.NISN,        // C: NISN
		data.NamaSiswa,   // D: NamaSiswa
		data.Status,      // E: Status
		"",               // F: Keterangan (bisa diisi jika perlu)
		"",               // G: URLBuktiFoto
		data.DicatatOleh, // H: DicatatOleh
	}
	values = append(values, row)

	valueRange := &sheets.ValueRange{Values: values}

	_, err := r.db.Spreadsheets.Values.Append(r.spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("Gagal menyimpan kehadiran manual ke sheet: %v", err)
		return err
	}
	log.Printf("Kehadiran manual untuk NISN [%s] dengan status [%s] berhasil dicatat.", data.NISN, data.Status)
	return nil
}

func (r *absensiRepository) CreateBatchManualAttendance(ctx context.Context, data []domain.KehadiranManual) error {
	writeRange := "LogAbsensi"

	var values [][]interface{}
	for _, item := range data {
		// Lewati item jika datanya tidak lengkap setelah proses di usecase
		if item.NamaSiswa == "" {
			continue
		}

		logID := fmt.Sprintf("LOG-%d", time.Now().UnixNano())
		row := []interface{}{
			logID,            // A: LogID
			item.Timestamp,   // B: Timestamp
			item.NISN,        // C: NISN
			item.NamaSiswa,   // D: NamaSiswa
			item.Status,      // E: Status
			"",               // F: Keterangan
			"",               // G: URLBuktiFoto
			item.DicatatOleh, // H: DicatatOleh
			"",               // I: TimestampPulang
			"",               // J: KeteranganPulang
		}
		values = append(values, row)
	}

	if len(values) == 0 {
		return nil // Tidak ada data valid untuk ditulis
	}

	valueRange := &sheets.ValueRange{Values: values}
	_, err := r.db.Spreadsheets.Values.Append(r.spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("Gagal menyimpan absensi massal ke sheet: %v", err)
		return err
	}

	log.Printf("%d data kehadiran manual berhasil dicatat.", len(values))
	return nil
}

func (r *absensiRepository) DeleteAttendance(ctx context.Context, rowNumber int) error {
	clearRange := fmt.Sprintf("LogAbsensi!A%d:E%d", rowNumber, rowNumber)
	_, err := r.db.Spreadsheets.Values.Clear(r.spreadsheetId, clearRange, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		log.Printf("Gagal menghapus baris %d di LogAbsensi: %v", rowNumber, err)
	}
	return err
}

func (r *absensiRepository) UpdateAttendance(ctx context.Context, rowNumber int, data *domain.KehadiranManual) error {
	updateRange := fmt.Sprintf("LogAbsensi!B%d:E%d", rowNumber, rowNumber) // Update kolom B sampai E
	var values [][]interface{}
	row := []interface{}{data.Timestamp, data.NISN, "", data.Status} // NamaSiswa (kolom D) dikosongkan
	values = append(values, row)

	valueRange := &sheets.ValueRange{Values: values}
	_, err := r.db.Spreadsheets.Values.Update(r.spreadsheetId, updateRange, valueRange).ValueInputOption("RAW").Do()
	return err
}

func (r *absensiRepository) GetLeaveByDate(ctx context.Context, date string) ([]domain.PengajuanIzinLengkap, error) {
	var izinList []domain.PengajuanIzinLengkap
	readRange := "PengajuanIzin!A2:J"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Values) == 0 {
		return izinList, nil
	}

	for i, row := range resp.Values {
		tglMulai := getStringFromCellByIndex(row, 5) // Kolom F
		if tglMulai == date {
			izinList = append(izinList, domain.PengajuanIzinLengkap{
				RowNumber: i + 2,
				// ... (sisa field lain)
			})
		}
	}
	return izinList, nil
}

func (r *absensiRepository) GetHolidays(ctx context.Context) (map[string]bool, error) {
	holidayMap := make(map[string]bool)
	readRange := "TanggalLibur!A2:A"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		if strings.Contains(err.Error(), "Unable to parse range") {
			return holidayMap, nil
		}
		return nil, err
	}
	for _, row := range resp.Values {
		holidayMap[getStringFromCellByIndex(row, 0)] = true
	}
	return holidayMap, nil
}

func (r *absensiRepository) GetAllLogsInMonth(ctx context.Context, year, month int) ([]domain.LogAbsensi, error) {
	var allLogs []domain.LogAbsensi
	monthPrefix := fmt.Sprintf("%d-%02d", year, month) // Format: YYYY-MM

	readRange := "LogAbsensi!A2:E"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	for i, row := range resp.Values {
		timestamp := getStringFromCellByIndex(row, 1) // Kolom B
		if strings.HasPrefix(timestamp, monthPrefix) {
			allLogs = append(allLogs, domain.LogAbsensi{
				RowNumber: i + 2,
				Timestamp: timestamp,
				Username:  getStringFromCellByIndex(row, 2),
				Status:    getStringFromCellByIndex(row, 4),
			})
		}
	}
	return allLogs, nil
}

// file: backend/internal/repository/absensi_repository_sheets.go

func (r *absensiRepository) GetLogsByDateRange(ctx context.Context, startDate, endDate string) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error) {
	var hadirLogs []domain.LogAbsensi
	var izinLogs []domain.PengajuanIzinLengkap

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("format tanggal mulai salah: %v", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, nil, fmt.Errorf("format tanggal selesai salah: %v", err)
	}
	end = end.Add(23*time.Hour + 59*time.Minute)

	// 2. Baca LogAbsensi
	logRange := "LogAbsensi!A2:E"
	logResp, _ := r.db.Spreadsheets.Values.Get(r.spreadsheetId, logRange).Do()
	if logResp != nil {
		for i, row := range logResp.Values {
			timestampStr := getStringFromCellByIndex(row, 1) // Kolom B
			// Coba parsing timestamp dari format "YYYY-MM-DD HH:MM:SS"
			checkTime, err := time.Parse("2006-01-02 15:04:05", timestampStr)
			if err != nil {
				continue
			} // Lewati jika format salah

			// Lakukan perbandingan waktu yang benar
			if !checkTime.Before(start) && !checkTime.After(end) {
				hadirLogs = append(hadirLogs, domain.LogAbsensi{
					RowNumber:   i + 2,
					Timestamp:   timestampStr,
					Username:    getStringFromCellByIndex(row, 2),
					NamaLengkap: "", // Akan diisi oleh usecase
					Status:      getStringFromCellByIndex(row, 4),
				})
			}
		}
	}

	// 3. Baca PengajuanIzin
	izinRange := "PengajuanIzin!A2:J"
	izinResp, _ := r.db.Spreadsheets.Values.Get(r.spreadsheetId, izinRange).Do()
	if izinResp != nil {
		for i, row := range izinResp.Values {
			tglMulaiStr := getStringFromCellByIndex(row, 5) // Kolom F
			checkTime, err := parseFlexibleDate(tglMulaiStr)
			if err != nil {
				continue
			}

			if !checkTime.Before(start) && !checkTime.After(end) {
				izinLogs = append(izinLogs, domain.PengajuanIzinLengkap{
					RowNumber:      i + 2,
					Timestamp:      getStringFromCellByIndex(row, 1),
					SiswaNISN:      getStringFromCellByIndex(row, 2),
					NamaLengkap:    "", // Akan diisi oleh usecase
					JenisIzin:      getStringFromCellByIndex(row, 4),
					TanggalMulai:   tglMulaiStr,
					TanggalSelesai: getStringFromCellByIndex(row, 6),
					Status:         getStringFromCellByIndex(row, 9),
				})
			}
		}
	}

	return hadirLogs, izinLogs, nil
}

func (r *absensiRepository) FindTodaysAttendanceLog(ctx context.Context, nisn string) (*domain.LogAbsensi, error) {
	today := time.Now().Format("2006-01-02")
	readRange := "LogAbsensi!A2:I" // Baca sampai TimestampPulang
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	for i, row := range resp.Values {
		rowNISN := getStringFromCellByIndex(row, 2)      // Kolom C
		rowTimestamp := getStringFromCellByIndex(row, 1) // Kolom B
		if rowNISN == nisn && strings.HasPrefix(rowTimestamp, today) {
			return &domain.LogAbsensi{
				RowNumber:       i + 2,
				TimestampPulang: getStringFromCellByIndex(row, 7), // Kolom H
			}, nil
		}
	}
	return nil, nil // Tidak ditemukan, bukan error
}

func (r *absensiRepository) UpdateClockOut(ctx context.Context, rowNumber int, clockOutTime string) error {
	// Update kolom H (TimestampPulang) dan I (KeteranganPulang)
	updateRange := fmt.Sprintf("LogAbsensi!I%d:J%d", rowNumber, rowNumber)
	var values [][]interface{}
	row := []interface{}{clockOutTime, "Scan QR Pulang"}
	values = append(values, row)

	valueRange := &sheets.ValueRange{Values: values}
	_, err := r.db.Spreadsheets.Values.Update(r.spreadsheetId, updateRange, valueRange).ValueInputOption("RAW").Do()
	return err
}

func (r *absensiRepository) GetAllLogsForUserInMonth(ctx context.Context, nisn string, year, month int) ([]domain.LogAbsensi, []domain.PengajuanIzinLengkap, error) {
	monthPrefix := fmt.Sprintf("%d-%02d", year, month)
	var hadirLogs []domain.LogAbsensi
	var izinLogs []domain.PengajuanIzinLengkap

	// Baca LogAbsensi
	logRange := "LogAbsensi!A2:K"
	logResp, _ := r.db.Spreadsheets.Values.Get(r.spreadsheetId, logRange).Do()
	if logResp != nil {
		for i, row := range logResp.Values {
			rowNISN := getStringFromCellByIndex(row, 2)
			timestamp := getStringFromCellByIndex(row, 1)
			if rowNISN == nisn && strings.HasPrefix(timestamp, monthPrefix) {
				hadirLogs = append(hadirLogs, domain.LogAbsensi{
					RowNumber:       i + 2,
					Timestamp:       timestamp,
					Username:        rowNISN,
					Status:          getStringFromCellByIndex(row, 4),
					TimestampPulang: getStringFromCellByIndex(row, 8),
				})
			}
		}
	}

	// Baca PengajuanIzin
	izinRange := "PengajuanIzin!A2:K"
	izinResp, _ := r.db.Spreadsheets.Values.Get(r.spreadsheetId, izinRange).Do()
	if izinResp != nil {
		for i, row := range izinResp.Values {
			rowNISN := getStringFromCellByIndex(row, 2)
			tglMulai := getStringFromCellByIndex(row, 5)
			if rowNISN == nisn && strings.HasPrefix(tglMulai, monthPrefix) {
				izinLogs = append(izinLogs, domain.PengajuanIzinLengkap{
					RowNumber:      i + 2,
					Timestamp:      getStringFromCellByIndex(row, 1),
					SiswaNISN:      rowNISN,
					JenisIzin:      getStringFromCellByIndex(row, 4),
					TanggalMulai:   tglMulai,
					TanggalSelesai: getStringFromCellByIndex(row, 6),
					Status:         getStringFromCellByIndex(row, 9),
				})
			}
		}
	}

	return hadirLogs, izinLogs, nil
}
