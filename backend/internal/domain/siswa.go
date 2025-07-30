// file: internal/domain/siswa.go
package domain

import "context"

// Siswa merepresentasikan data seorang siswa
type Siswa struct {
	NISN             string `json:"nisn"`
	NamaLengkap      string `json:"namaLengkap"`
	Kelas            string `json:"kelas"`
	NamaOrangTua     string `json:"namaOrangTua"` // <-- TAMBAHKAN INI
	NomorTeleponOrtu string `json:"nomorTeleponOrtu"`
	EmailOrtu        string `json:"emailOrtu"`
}

// Definisikan kontrak-kontrak baru untuk fitur manajemen siswa
type SiswaRepository interface {
	FindAll(ctx context.Context) ([]Siswa, error)
	FindByNISN(ctx context.Context, nisn string) (*Siswa, error)
	Save(ctx context.Context, siswa *Siswa) error
	Update(ctx context.Context, nisn string, siswa *Siswa) error
	Delete(ctx context.Context, nisn string) error
}

type SiswaUsecase interface {
	GetAll(ctx context.Context) ([]Siswa, error)
	Create(ctx context.Context, siswa *Siswa) error
	GetByNISN(ctx context.Context, nisn string) (*Siswa, error)
	Update(ctx context.Context, nisn string, siswa *Siswa) error
	Delete(ctx context.Context, nisn string) error
}
