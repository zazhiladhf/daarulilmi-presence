// file: internal/usecase/siswa_usecase.go
package usecase

import (
	"context"
	"errors"

	"daarulilmi-presence/internal/domain" // Ganti dengan nama modul Anda
)

type siswaUsecase struct {
	repo domain.SiswaRepository
}

func NewSiswaUsecase(repo domain.SiswaRepository) domain.SiswaUsecase {
	return &siswaUsecase{repo}
}

func (uc *siswaUsecase) GetAll(ctx context.Context) ([]domain.Siswa, error) {
	return uc.repo.FindAll(ctx)
}

// --- FUNGSI BARU UNTUK MEMBUAT SISWA ---
func (uc *siswaUsecase) Create(ctx context.Context, siswa *domain.Siswa) error {
	// Validasi agar NISN tidak duplikat
	existing, err := uc.repo.FindByNISN(ctx, siswa.NISN)
	if err != nil {
		return err // Error saat mengakses database
	}
	if existing != nil {
		return errors.New("NISN sudah terdaftar")
	}

	// Jika aman, simpan siswa baru
	return uc.repo.Save(ctx, siswa)
}

func (uc *siswaUsecase) GetByNISN(ctx context.Context, nisn string) (*domain.Siswa, error) {
	return uc.repo.FindByNISN(ctx, nisn)
}

func (uc *siswaUsecase) Update(ctx context.Context, nisn string, siswa *domain.Siswa) error {
	// Di sini bisa ditambahkan validasi, misalnya memastikan NISN tidak diubah ke yang sudah ada
	return uc.repo.Update(ctx, nisn, siswa)
}

func (uc *siswaUsecase) Delete(ctx context.Context, nisn string) error {
	return uc.repo.Delete(ctx, nisn)
}
