// file: internal/usecase/user_usecase.go
package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"daarulilmi-presence/internal/domain" // Ganti dengan nama modul Anda

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Map untuk menyimpan token reset [token] -> username
var resetTokens = make(map[string]string)

// Map untuk menyimpan waktu kedaluwarsa token [token] -> waktu
var resetTokenExpiry = make(map[string]time.Time)

// Definisikan "kontrak" atau job description untuk repository
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Save(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, currentUsername string, user *domain.User) error
}

type userUsecase struct {
	userRepo  UserRepository
	jwtSecret []byte
}

// NewUserUsecase adalah "pabrik" untuk usecase
func NewUserUsecase(userRepo UserRepository, jwtSecret []byte) domain.UserUsecase {
	return &userUsecase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// --- FUNGSI BARU UNTUK MEMBUAT TOKEN ---
func (uc *userUsecase) generateJWT(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Gunakan secret dari struct: uc.jwtSecret
	tokenString, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// --- MODIFIKASI FUNGSI LOGIN ---
// Sekarang fungsi Login akan mengembalikan token (string) bukan user object
func (uc *userUsecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", err // Error saat koneksi ke DB
	}
	if user == nil {
		return "", nil // User tidak ditemukan
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", nil // Password salah
	}

	// Jika login berhasil, panggil fungsi untuk membuat token
	token, err := uc.generateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil // Kembalikan token sebagai string
}

func (uc *userUsecase) Register(ctx context.Context, user *domain.User) error {
	// 1. Cek apakah username sudah ada
	existingUser, err := uc.userRepo.FindByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("username sudah digunakan")
	}

	// 2. Hash password sebelum disimpan
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// 3. Simpan pengguna baru via repository
	return uc.userRepo.Save(ctx, user)
}

func (uc *userUsecase) UpdateUser(ctx context.Context, currentUsername string, newUsername, newPassword string) error {
	// Ambil data user saat ini untuk mendapatkan data yang tidak berubah
	currentUser, err := uc.userRepo.FindByUsername(ctx, currentUsername)
	if err != nil || currentUser == nil {
		return errors.New("pengguna saat ini tidak ditemukan")
	}

	// Siapkan data untuk diupdate
	updateData := domain.User{
		Username:     currentUser.Username,
		PasswordHash: currentUser.PasswordHash, // Defaultnya pakai yg lama
		Role:         currentUser.Role,
	}

	// Jika username baru diisi, gunakan itu
	if newUsername != "" {
		// Cek apakah username baru sudah dipakai orang lain
		existingUser, err := uc.userRepo.FindByUsername(ctx, newUsername)
		if err != nil {
			return err
		}
		if existingUser != nil {
			return errors.New("username baru sudah digunakan")
		}
		updateData.Username = newUsername
	}

	// Jika password baru diisi, hash dan gunakan itu
	if newPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updateData.PasswordHash = string(hashedPassword)
	}

	// Panggil repository untuk menyimpan perubahan
	return uc.userRepo.Update(ctx, currentUsername, &updateData)
}

// --- FUNGSI BARU UNTUK MINTA RESET ---
func (uc *userUsecase) RequestPasswordReset(ctx context.Context, username string) (string, error) {
	// Cek apakah user ada
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil || user == nil {
		return "", errors.New("username tidak ditemukan")
	}

	// Buat token acak yang aman
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	// Simpan token dan waktu kedaluwarsa (15 menit)
	resetTokens[token] = username
	resetTokenExpiry[token] = time.Now().Add(15 * time.Minute)

	return token, nil
}

// --- FUNGSI BARU UNTUK PROSES RESET ---
func (uc *userUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Cek token di penyimpanan
	username, exists := resetTokens[token]
	if !exists {
		return errors.New("token tidak valid")
	}

	// Cek waktu kedaluwarsa
	if time.Now().After(resetTokenExpiry[token]) {
		delete(resetTokens, token)
		delete(resetTokenExpiry, token)
		return errors.New("token sudah kedaluwarsa")
	}

	// Hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Siapkan data update
	updateData := &domain.User{
		Username:     username, // Username tidak berubah
		PasswordHash: string(hashedPassword),
	}

	// Panggil repo untuk update (kita perlu modifikasi repo.Update sedikit)
	err = uc.userRepo.Update(ctx, username, updateData)
	if err != nil {
		return err
	}

	// Hapus token setelah berhasil digunakan
	delete(resetTokens, token)
	delete(resetTokenExpiry, token)

	return nil
}

func (uc *userUsecase) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	// Tugasnya hanya meneruskan permintaan ke repository
	return uc.userRepo.FindByUsername(ctx, username)
}
