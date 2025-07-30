// file: internal/handler/user_handler.go
package handler

import (
	"log"
	"net/http"
	"strings"

	"daarulilmi-presence/internal/domain" // Ganti dengan nama modul Anda

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// UserHandler adalah "resepsionis" untuk semua urusan pengguna
type UserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUserHandler(e *echo.Echo, api *echo.Group, userUsecase domain.UserUsecase) {
	handler := &UserHandler{userUsecase}

	// Daftarkan rute-rute yang akan ditangani oleh handler ini
	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	e.POST("/request-reset", handler.RequestPasswordReset)
	e.POST("/reset-password", handler.ResetPassword)
	e.GET("/reset-password", handler.ShowResetPasswordForm)

	// Rute terproteksi
	api.POST("/user/update", handler.UpdateUser)
	api.GET("/user/profile", handler.GetUserProfileAPI)
}

// Login adalah fungsi yang dipanggil saat ada request ke POST /login
func (h *UserHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	token, err := h.userUsecase.Login(c.Request().Context(), username, password)
	if err != nil || token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau password salah"})
	}

	// Jika berhasil, kirim token dalam format JSON
	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}

func (h *UserHandler) Register(c echo.Context) error {
	// Buat objek user dari data form
	user := new(domain.User)
	user.Username = c.FormValue("username")
	user.PasswordHash = c.FormValue("password") // Ini masih password mentah
	user.Role = c.FormValue("role")

	// Panggil usecase untuk melakukan proses registrasi
	err := h.userUsecase.Register(c.Request().Context(), user)
	if err != nil {
		// Jika ada error (misal: username sudah ada), kembali ke halaman register
		// Nanti bisa ditambahkan pesan error
		errorMsg := "Terjadi kesalahan. Coba lagi."
		if err.Error() == "username sudah digunakan" {
			errorMsg = "Username sudah digunakan. Silakan pilih yang lain."
		}
		return c.Redirect(http.StatusSeeOther, "/register?error="+errorMsg)

	}

	// Jika berhasil, arahkan ke halaman login
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	// Ambil username saat ini dari token JWT
	userClaims := c.Get("user").(jwt.MapClaims)
	currentUsername := userClaims["username"].(string)

	// Ambil data baru dari form
	newUsername := c.FormValue("new_username")
	newPassword := c.FormValue("new_password")

	err := h.userUsecase.UpdateUser(c.Request().Context(), currentUsername, newUsername, newPassword)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Profil berhasil diperbarui."})
}

// --- MIDDLEWARE PENJAGA KEAMANAN JWT ---
func JWTMiddleware(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Missing authorization header"})
			}

			// Format header adalah "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid authorization header format"})
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validasi metode signing
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
				}
				return secret, nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
			}

			// Simpan data dari token ke context agar bisa diakses handler selanjutnya
			c.Set("user", token.Claims)
			return next(c)
		}
	}
}

func (h *UserHandler) RequestPasswordReset(c echo.Context) error {
	username := c.FormValue("username")
	token, err := h.userUsecase.RequestPasswordReset(c.Request().Context(), username)
	if err != nil {
		// Demi keamanan, jangan beri tahu jika username ada/tidak.
		return c.String(http.StatusOK, "Jika username terdaftar, link reset akan disimulasikan di konsol server.")
	}

	// Simulasi pengiriman email: cetak link ke konsol
	log.Printf("== SIMULASI EMAIL ==")
	log.Printf("Link Reset Password untuk %s: http://localhost:1412/reset-password?token=%s", username, token)
	log.Printf("====================")

	return c.String(http.StatusOK, "Jika username terdaftar, link reset akan disimulasikan di konsol server.")
}

func (h *UserHandler) ShowResetPasswordForm(c echo.Context) error {
	token := c.QueryParam("token")
	// Di aplikasi nyata, kita harus validasi token di sini sebelum menampilkan form
	return c.Render(http.StatusOK, "reset_password.html", map[string]interface{}{"token": token})
}

func (h *UserHandler) ResetPassword(c echo.Context) error {
	token := c.FormValue("token")
	newPassword := c.FormValue("new_password")

	err := h.userUsecase.ResetPassword(c.Request().Context(), token, newPassword)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Password berhasil direset. Silakan login dengan password baru Anda.")
}

func (h *UserHandler) GetUserProfileAPI(c echo.Context) error {
	// Ambil username dari token JWT yang sudah divalidasi middleware
	userClaims := c.Get("user").(jwt.MapClaims)
	username := userClaims["username"].(string)

	// Panggil usecase untuk mencari data user
	user, err := h.userUsecase.GetByUsername(c.Request().Context(), username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data profil"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Profil pengguna tidak ditemukan"})
	}

	return c.JSON(http.StatusOK, user)
}
