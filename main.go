// main.go - VERSI FINAL MODUL 4
package main

import (
	"context" // Pustaka untuk mengelola konteks permintaan
	"html/template"
	"io"
	"log" // Pustaka untuk logging
	"net/http"
	"os" // Pustaka untuk membaca file dari sistem operasi

	"github.com/labstack/echo/v4"
	// Pustaka baru untuk otentikasi & akses Google API
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Definisikan struktur data untuk Siswa
type Siswa struct {
	NISN        string
	NamaLengkap string
	Kelas       string
}

// Template Renderer (SAMA SEPERTI SEBELUMNYA)
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// --- FUNGSI BARU: Untuk membaca data dari Google Sheets ---
func bacaDataSiswa() []Siswa {
	// Membaca file kredensial yang sudah kita pindahkan
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Tidak dapat membaca file kredensial: %v", err)
	}

	// Konfigurasi akses ke Sheets API
	srv, err := sheets.NewService(context.Background(), option.WithCredentialsJSON(b))
	if err != nil {
		log.Fatalf("Tidak dapat membuat koneksi ke Sheets: %v", err)
	}

	// === GANTI DENGAN ID SPREADSHEET ANDA ===
	spreadsheetId := "1TFLV9ezeLt-q3uyNvArMfWwYoz5tDOGD-25zoPHXM3E"
	// Menentukan rentang data yang akan dibaca
	readRange := "DataSiswa!A2:C" // Baca dari sheet DataSiswa, kolom A sampai C, mulai dari baris 2

	// Memanggil API untuk mengambil data
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Tidak dapat mengambil data dari sheet: %v", err)
	}

	// Proses data yang diterima
	var semuaSiswa []Siswa
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			// Membuat objek Siswa dari setiap baris data
			siswa := Siswa{
				NISN:        row[0].(string),
				NamaLengkap: row[1].(string),
				Kelas:       row[2].(string),
			}
			semuaSiswa = append(semuaSiswa, siswa)
		}
	}
	return semuaSiswa
}

func main() {
	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	// Rute Halaman Login
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	})

	// --- RUTE BARU: Halaman Dashboard ---
	e.GET("/dashboard", func(c echo.Context) error {
		// Panggil fungsi untuk membaca data dari Google Sheets
		dataSiswa := bacaDataSiswa()
		// Kirim data tersebut ke halaman dashboard_walikelas.html
		return c.Render(http.StatusOK, "dashboard_walikelas.html", dataSiswa)
	})

	log.Println("Server berjalan di http://localhost:1412")
	e.Logger.Fatal(e.Start(":1412"))
}