// file: internal/handler/absensi_handler.go
package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"daarulilmi-presence/internal/domain" // Ganti dengan nama modul Anda

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AbsensiHandler struct {
	absensiUsecase domain.AbsensiUsecase
}

func NewAbsensiHandler(e *echo.Echo, api *echo.Group, absensiUsecase domain.AbsensiUsecase) {
	handler := &AbsensiHandler{absensiUsecase}

	// Rute API terproteksi
	api.GET("/qr/generate", handler.GenerateQR)
	api.POST("/absensi/scan", handler.Scan)
	api.GET("/portal-data", handler.GetPortalData)
	api.GET("/dashboard-data", handler.GetDashboardData)
	api.POST("/absensi/manual", handler.CreateManualAttendanceAPI)
	api.POST("/absensi/manual/batch", handler.CreateBatchManualAttendanceAPI)
	api.DELETE("/absensi/log/:row", handler.DeleteAttendanceAPI)
	api.PUT("/absensi/log/:row", handler.UpdateAttendanceAPI)
	api.GET("/absensi/log/:row", handler.GetAttendanceByRowAPI)
	api.GET("/absensi/rekap/:tanggal", handler.GetRekapByDateAPI)
	api.GET("/statistik/bulanan/:tahun/:bulan", handler.GetMonthlyStatsAPI)
	api.GET("/rekap", handler.GetRekapAPI)
	api.GET("/portal/dashboard-data/:tahun/:bulan", handler.GetPortalDashboardDataAPI)

	// Rute Halaman
	e.GET("/dashboard", handler.ShowDashboardPage)
	e.GET("/admin/izin", handler.ShowAdminIzinPage)
}

type ScanRequest struct {
	QRData string `json:"qr_data"`
}

func (h *AbsensiHandler) GetAttendanceByRowAPI(c echo.Context) error {
	rowStr := c.Param("row")
	row, err := strconv.Atoi(rowStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nomor baris tidak valid"})
	}

	logData, err := h.absensiUsecase.GetAttendanceByRow(c.Request().Context(), row)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, logData)
}

// GenerateQR menangani request untuk membuat QR code
func (h *AbsensiHandler) GenerateQR(c echo.Context) error {
	qrType := c.QueryParam("type")
	if qrType != "masuk" && qrType != "pulang" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tipe QR tidak valid"})
	}

	pngBytes, err := h.absensiUsecase.GenerateQR(c.Request().Context(), qrType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuat QR code"})
	}

	// Set content type header menjadi image/png dan kirim byte gambar sebagai response
	return c.Blob(http.StatusOK, "image/png", pngBytes)
}

func (h *AbsensiHandler) Scan(c echo.Context) error {
	// Ambil username dari token JWT
	userClaims := c.Get("user").(jwt.MapClaims)
	username := userClaims["username"].(string)

	// Ambil data QR dari body request
	req := new(ScanRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Request tidak valid"})
	}

	// Panggil usecase untuk verifikasi dan catat
	message, err := h.absensiUsecase.VerifyAndRecordScan(c.Request().Context(), req.QRData, username)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": message})
}

func (h *AbsensiHandler) GetPortalData(c echo.Context) error {
	// Logika untuk mengambil username dari JWT tetap sama
	userClaims := c.Get("user").(jwt.MapClaims)
	username := userClaims["username"].(string)

	// Logika untuk memanggil usecase tetap sama
	attendanceData, err := h.absensiUsecase.GetAttendanceForUser(c.Request().Context(), username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data absensi"})
	}

	// Siapkan data untuk dikirim
	data := map[string]interface{}{
		"Username":   username,
		"Attendance": attendanceData,
	}

	// Kembalikan data sebagai JSON, bukan render HTML
	return c.JSON(http.StatusOK, data)
}

func (h *AbsensiHandler) ShowDashboardPage(c echo.Context) error {
	// KEMBALIKAN KE KODE RENDER YANG BENAR
	return c.Render(http.StatusOK, "dashboard_walikelas.html", map[string]interface{}{})
}

func (h *AbsensiHandler) ShowAdminIzinPage(c echo.Context) error {
	requests, err := h.absensiUsecase.GetAllLeaveRequests(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "Gagal mengambil data izin")
	}
	data := map[string]interface{}{"LeaveRequests": requests}
	return c.Render(http.StatusOK, "admin_izin.html", data)
}

func (h *AbsensiHandler) GetDashboardData(c echo.Context) error {
	log.Println("==============================================")
	log.Println(">>> API /api/dashboard-data BERHASIL DIPANGGIL! <<<")
	log.Println("==============================================")

	// --- PERBAIKAN DI SINI ---
	// Ambil data pengguna dari konteks dengan aman
	user := c.Get("user")
	if user == nil {
		log.Println("CRITICAL ERROR: Middleware JWT tidak berhasil menyimpan data user ke konteks.")
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan internal pada server (konteks user nil)"})
	}

	// Lakukan type assertion dengan aman
	userClaims, ok := user.(jwt.MapClaims)
	if !ok {
		log.Println("CRITICAL ERROR: Gagal melakukan type assertion untuk jwt.MapClaims.")
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan internal pada server (type assertion gagal)"})
	}
	// --- AKHIR PERBAIKAN ---

	username := userClaims["username"].(string)

	// Panggil usecase yang sudah benar
	smartData, err := h.absensiUsecase.GetSmartDashboardData(c.Request().Context(), username, "")
	if err != nil {
		log.Printf("ERROR getting smart dashboard data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memuat data dashboard"})
	}

	return c.JSON(http.StatusOK, smartData)
}

func (h *AbsensiHandler) GetTodaysAttendanceAPI(c echo.Context) error {
	// Panggil logika "Smart Dashboard" yang sudah ada
	smartData, err := h.absensiUsecase.GetSmartDashboardData(c.Request().Context(), "", time.Now().Format("2006-01-02"))
	if err != nil {
		log.Printf("ERROR getting smart dashboard data for attendance page: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memuat data"})
	}

	// Kirim hanya daftar status siswa yang sudah diolah
	return c.JSON(http.StatusOK, smartData.DaftarStatusSiswa)
}

func (h *AbsensiHandler) CreateManualAttendanceAPI(c echo.Context) error {
	data := new(domain.KehadiranManual)
	if err := c.Bind(data); err != nil {
		log.Printf("ERROR binding data kehadiran manual: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data yang dikirim tidak valid"})
	}

	// Set timestamp ke waktu sekarang jika tidak disediakan oleh frontend
	if data.Timestamp == "" {
		data.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	err := h.absensiUsecase.CreateManualAttendance(c.Request().Context(), data)
	if err != nil {
		log.Printf("ERROR usecase CreateManualAttendance: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Kehadiran berhasil dicatat secara manual"})
}

func (h *AbsensiHandler) CreateBatchManualAttendanceAPI(c echo.Context) error {
	var data []domain.KehadiranManual
	if err := c.Bind(&data); err != nil {
		log.Printf("ERROR binding data absensi massal: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data yang dikirim tidak valid"})
	}

	err := h.absensiUsecase.CreateBatchManualAttendance(c.Request().Context(), data)
	if err != nil {
		log.Printf("ERROR usecase CreateBatchManualAttendance: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Data kehadiran massal berhasil dicatat"})
}

func (h *AbsensiHandler) DeleteAttendanceAPI(c echo.Context) error {
	rowStr := c.Param("row")
	row, err := strconv.Atoi(rowStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nomor baris tidak valid"})
	}

	err = h.absensiUsecase.DeleteAttendance(c.Request().Context(), row)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Data kehadiran berhasil dihapus"})
}

func (h *AbsensiHandler) UpdateAttendanceAPI(c echo.Context) error {
	rowStr := c.Param("row")
	row, err := strconv.Atoi(rowStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nomor baris tidak valid"})
	}

	data := new(domain.KehadiranManual)
	if err := c.Bind(data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data yang dikirim tidak valid"})
	}

	// PASTIKAN FUNGSI YANG DIPANGGIL ADALAH UpdateAttendance
	err = h.absensiUsecase.UpdateAttendance(c.Request().Context(), row, data)
	if err != nil {
		log.Printf("ERROR usecase UpdateAttendance: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Data kehadiran berhasil diperbarui"})
}

func (h *AbsensiHandler) GetRekapByDateAPI(c echo.Context) error {
	tanggal := c.Param("tanggal")
	smartData, err := h.absensiUsecase.GetSmartDashboardData(c.Request().Context(), "", tanggal)
	if err != nil {
		log.Printf("ERROR getting smart dashboard data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memuat data"})
	}
	return c.JSON(http.StatusOK, smartData)
}

func (h *AbsensiHandler) GetMonthlyStatsAPI(c echo.Context) error {
	tahunStr := c.Param("tahun")
	bulanStr := c.Param("bulan")
	tahun, _ := strconv.Atoi(tahunStr)
	bulan, _ := strconv.Atoi(bulanStr)

	stats, err := h.absensiUsecase.GetMonthlyStats(c.Request().Context(), tahun, bulan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}

func (h *AbsensiHandler) GetRekapAPI(c echo.Context) error {
	startDate := c.QueryParam("mulai")
	endDate := c.QueryParam("selesai")

	rekapData, err := h.absensiUsecase.GetRekapByDateRange(c.Request().Context(), startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, rekapData)
}
func (h *AbsensiHandler) GetPortalDashboardDataAPI(c echo.Context) error {
	userClaims := c.Get("user").(jwt.MapClaims)
	username := userClaims["username"].(string)

	tahunStr := c.Param("tahun")
	bulanStr := c.Param("bulan")
	tahun, _ := strconv.Atoi(tahunStr)
	bulan, _ := strconv.Atoi(bulanStr)

	data, err := h.absensiUsecase.GetPortalDashboardData(c.Request().Context(), username, tahun, bulan)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}
