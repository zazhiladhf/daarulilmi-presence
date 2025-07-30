// file: internal/domain/user.go
package domain

import "context"

// User mendefinisikan struktur data utama dari pengguna
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // Sembunyikan dari JSON
	Role         string `json:"role"`
	NamaLengkap  string `json:"namaLengkap"`
	SiswaNISN    string `json:"siswaNisn"` // <-- TAMBAHKAN INI
}

type UserUsecase interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, currentUsername string, newUsername, newPassword string) error
	RequestPasswordReset(ctx context.Context, username string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
	GetByUsername(ctx context.Context, username string) (*User, error)
}
