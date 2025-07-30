// file: backend/generate_hash.go
package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := []byte("siswa123") // Ganti dengan password awal siswa
	hashedPassword, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	fmt.Println(string(hashedPassword))
}
