// file: cmd/web/main.go (Versi Final & Lengkap)

package main

import (
	"context"
	"log"
	"os"

	"daarulilmi-presence/internal/handler"
	"daarulilmi-presence/internal/repository"
	"daarulilmi-presence/internal/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// --- KONFIGURASI GLOBAL ---
var (
	// Ganti nama sekolah di sini jika perlu
	// schoolName = "SMA Islam Daarul Ilmi Depok"
	// Kunci rahasia untuk JWT, harus sama persis dengan yang di usecase
	jwtSecret = []byte("daarulilmi-presence")
	// ID Spreadsheet dari URL Google Sheet Anda
	spreadsheetId = "1TFLV9ezeLt-q3uyNvArMfWwYoz5tDOGD-25zoPHXM3E"
)

func main() {
	// --- SETUP KONEKSI GOOGLE SHEETS ---
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Gagal membaca file kredensial: %v", err)
	}
	srv, err := sheets.NewService(context.Background(), option.WithCredentialsJSON(b))
	if err != nil {
		log.Fatalf("Gagal membuat koneksi ke Sheets: %v", err)
	}

	// === DEPENDENCY INJECTION (MERAKIT SEMUA KOMPONEN) ===
	// 1. Buat semua Repository (Kurir)
	userRepo := repository.NewUserRepository(srv, spreadsheetId)
	absensiRepo := repository.NewAbsensiRepository(srv, spreadsheetId)
	siswaRepo := repository.NewSiswaRepository(srv, spreadsheetId)

	// 2. Buat semua Usecase (Otak Bisnis)
	userUsecase := usecase.NewUserUsecase(userRepo, jwtSecret)
	absensiUsecase := usecase.NewAbsensiUsecase(absensiRepo, siswaRepo, userRepo)
	siswaUsecase := usecase.NewSiswaUsecase(siswaRepo)

	// --- SETUP SERVER ECHO ---
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// --- SETUP ROUTING ---
	// Grup untuk rute API yang butuh login
	apiGroup := e.Group("/api")
	apiGroup.Use(handler.JWTMiddleware(jwtSecret))

	// Daftarkan semua Handler (Resepsionis)
	handler.NewUserHandler(e, apiGroup, userUsecase)
	handler.NewAbsensiHandler(e, apiGroup, absensiUsecase)
	handler.NewSiswaHandler(e, apiGroup, siswaUsecase)

	// Rute Halaman Publik (tidak butuh login)
	// e.GET("/", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "login.html", nil)
	// })
	// e.GET("/register", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "register.html", map[string]interface{}{"error": c.QueryParam("error")})
	// })
	// e.GET("/forgot-password", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "forgot_password.html", nil)
	// })
	// e.GET("/reset-password", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "reset_password.html", map[string]interface{}{"token": c.QueryParam("token")})
	// })
	// e.GET("/scan", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "scanner.html", nil)
	// })

	// // Rute Halaman yang butuh login (keamanan ditangani oleh JS di frontend)
	// e.GET("/portal", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "portal_walimurid.html", nil)
	// })
	// e.GET("/user/update", func(c echo.Context) error {
	// 	return c.Render(http.StatusOK, "update_user.html", nil)
	// })

	// --- MULAI SERVER ---
	log.Println("Server berjalan di http://localhost:1412")
	e.Logger.Fatal(e.Start(":1412"))
}
