// file: internal/domain/absensi.go
package domain

import "context"

type LogAbsensi struct {
	RowNumber       int    `json:"rowNumber"`
	Timestamp       string `json:"timestamp"`
	Username        string `json:"username"`
	NamaLengkap     string `json:"namaLengkap"`
	Status          string `json:"status"`
	TimestampPulang string `json:"timestampPulang"`
}

type PengajuanIzinLengkap struct {
	RowNumber      int    `json:"rowNumber"`
	Timestamp      string `json:"timestamp"`
	SiswaNISN      string `json:"siswaNISN"`
	NamaLengkap    string `json:"namaLengkap"`
	JenisIzin      string `json:"jenisIzin"`
	TanggalMulai   string `json:"tanggalMulai"`
	TanggalSelesai string `json:"tanggalSelesai"`
	Status         string `json:"status"`
}

type DashboardData struct {
	NamaLengkap         string
	Username            string
	TotalSiswa          int
	TotalHadirHariIni   int
	TotalIzinHariIni    int
	TotalBelumAdaKabar  int
	LogKehadiranHariIni []LogAbsensi
	SiswaIzinHariIni    []PengajuanIzinLengkap
}

type SmartDashboardData struct {
	NamaLengkapUser    string        `json:"namaLengkapUser"`
	IsHoliday          bool          `json:"isHoliday"`
	HolidayDescription string        `json:"holidayDescription"`
	TotalSiswa         int           `json:"totalSiswa"`
	TotalHadir         int           `json:"totalHadir"`
	TotalIzin          int           `json:"totalIzin"`
	TotalBelumAdaKabar int           `json:"totalBelumAdaKabar"`
	DaftarStatusSiswa  []SiswaStatus `json:"daftarStatusSiswa"`
}

type SiswaStatus struct {
	RowNumber   int    `json:"rowNumber"`
	NISN        string `json:"nisn"`
	NamaLengkap string `json:"namaLengkap"`
	Kelas       string `json:"kelas"`
	Status      string `json:"status"`
	Keterangan  string `json:"keterangan"`
}

type KehadiranManual struct {
	NISN        string `json:"NISN"`
	NamaSiswa   string `json:"NamaSiswa,omitempty"`
	Status      string `json:"Status"`
	Timestamp   string `json:"Timestamp,omitempty"`
	DicatatOleh string `json:"DicatatOleh,omitempty"`
}

type RekapSiswa struct {
	NISN        string `json:"nisn"`
	NamaLengkap string `json:"namaLengkap"`
	Hadir       int    `json:"hadir"`
	Izin        int    `json:"izin"`
	Sakit       int    `json:"sakit"`
	Alpa        int    `json:"alpa"`
}

// Struct baru untuk data terpadu di portal wali murid
type PortalDashboardData struct {
	Siswa          *Siswa          `json:"siswa"`
	Events         []CalendarEvent `json:"events"`
	StatistikBulan *StatistikData  `json:"statistikBulan"`
}

// Struct untuk event di kalender
type CalendarEvent struct {
	Title string `json:"title"`
	Start string `json:"start"`
	Color string `json:"color"`
}

type StatistikData struct {
	TotalHadir int `json:"totalHadir"`
	TotalIzin  int `json:"totalIzin"`
	TotalSakit int `json:"totalSakit"`
	TotalAlpa  int `json:"totalAlpa"`
}

// AbsensiUsecase mendefinisikan kontrak untuk logika bisnis absensi.
type AbsensiUsecase interface {
	GenerateQR(ctx context.Context, qrType string) ([]byte, error)
	VerifyAndRecordScan(ctx context.Context, qrData string, username string) (string, error)
	GetAttendanceForUser(ctx context.Context, username string) ([]LogAbsensi, error)
	GetAllLeaveRequests(ctx context.Context) ([]PengajuanIzinLengkap, error)
	GetDashboardData(ctx context.Context, username string) (*DashboardData, error)
	GetTodaysAttendanceAndLeave(ctx context.Context) ([]LogAbsensi, []PengajuanIzinLengkap, error)
	GetSmartDashboardData(ctx context.Context, username string, dateStr string) (*SmartDashboardData, error)
	CreateManualAttendance(ctx context.Context, data *KehadiranManual) error
	CreateBatchManualAttendance(ctx context.Context, data []KehadiranManual) error
	DeleteAttendance(ctx context.Context, rowNumber int) error
	UpdateAttendance(ctx context.Context, rowNumber int, data *KehadiranManual) error
	GetAttendanceByRow(ctx context.Context, rowNumber int) (*LogAbsensi, error)
	GetMonthlyStats(ctx context.Context, year, month int) (*StatistikData, error)
	GetRekapByDateRange(ctx context.Context, startDate, endDate string) ([]RekapSiswa, error)
	GetPortalDashboardData(ctx context.Context, username string, year, month int) (*PortalDashboardData, error)
}
