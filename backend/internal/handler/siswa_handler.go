// file: internal/handler/siswa_handler.go

package handler

import (
	"log"
	"net/http"

	"daarulilmi-presence/internal/domain"

	"github.com/labstack/echo/v4"
)

type SiswaHandler struct {
	usecase domain.SiswaUsecase
}

func NewSiswaHandler(e *echo.Echo, api *echo.Group, usecase domain.SiswaUsecase) {
	handler := &SiswaHandler{usecase}

	// Rute Halaman
	e.GET("/admin/siswa", handler.ShowSiswaPage)
	e.GET("/admin/siswa/tambah", handler.ShowTambahSiswaPage)

	// Rute API
	api.GET("/siswa", handler.GetAllSiswaAPI)
	api.POST("/admin/siswa/tambah", handler.CreateSiswa)
	api.GET("/siswa/:nisn", handler.GetSiswaByNISNAPI)
	api.PUT("/siswa/:nisn", handler.UpdateSiswaAPI)
	api.DELETE("/siswa/:nisn", handler.DeleteSiswaAPI)
}

func (h *SiswaHandler) GetAllSiswaAPI(c echo.Context) error {
	siswaList, err := h.usecase.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data siswa"})
	}
	return c.JSON(http.StatusOK, siswaList)
}

func (h *SiswaHandler) ShowSiswaPage(c echo.Context) error {
	// KEMBALIKAN KE KODE RENDER YANG BENAR
	siswaList, err := h.usecase.GetAll(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "Gagal mengambil data siswa")
	}
	data := map[string]interface{}{
		"Siswa": siswaList,
	}
	return c.Render(http.StatusOK, "manajemen_siswa.html", data)
}

func (h *SiswaHandler) ShowTambahSiswaPage(c echo.Context) error {
	data := map[string]interface{}{
		"error": c.QueryParam("error"),
	}
	return c.Render(http.StatusOK, "tambah_siswa.html", data)
}

func (h *SiswaHandler) CreateSiswa(c echo.Context) error {
	siswa := new(domain.Siswa)
	if err := c.Bind(siswa); err != nil {
		// Jika data yang dikirim tidak sesuai format, kirim error
		log.Printf("ERROR binding data siswa: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data yang dikirim tidak valid"})
	}

	err := h.usecase.Create(c.Request().Context(), siswa)
	if err != nil {
		// Jika ada error dari usecase (misal: NISN duplikat)
		log.Printf("ERROR usecase CreateSiswa: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	// Jika sukses, kirim pesan sukses
	log.Printf("INFO: Siswa baru berhasil dibuat dengan NISN %s", siswa.NISN)
	return c.JSON(http.StatusCreated, map[string]string{"message": "Siswa baru berhasil ditambahkan"})
}

func (h *SiswaHandler) GetSiswaByNISNAPI(c echo.Context) error {
	nisn := c.Param("nisn")
	siswa, err := h.usecase.GetByNISN(c.Request().Context(), nisn)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data siswa"})
	}
	if siswa == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
	}
	return c.JSON(http.StatusOK, siswa)
}

func (h *SiswaHandler) UpdateSiswaAPI(c echo.Context) error {
	nisn := c.Param("nisn")
	siswa := new(domain.Siswa)
	if err := c.Bind(siswa); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	err := h.usecase.Update(c.Request().Context(), nisn, siswa)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Data siswa berhasil diperbarui"})
}

func (h *SiswaHandler) DeleteSiswaAPI(c echo.Context) error {
	nisn := c.Param("nisn")
	err := h.usecase.Delete(c.Request().Context(), nisn)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Data siswa berhasil dihapus"})
}
