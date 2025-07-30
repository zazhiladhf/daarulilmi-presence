// file: internal/repository/user_repository_sheets.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"daarulilmi-presence/internal/domain"

	"google.golang.org/api/sheets/v4"
)

type userRepository struct {
	db            *sheets.Service
	spreadsheetId string
}

// NewUserRepository adalah "pabrik" yang membuat objek repository baru
func NewUserRepository(db *sheets.Service, spreadsheetId string) *userRepository {
	return &userRepository{db, spreadsheetId}
}

// FindByUsername adalah implementasi nyata untuk mencari user
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	readRange := "DataPengguna!A2:E" // Baca sampai kolom E (SiswaNISN)
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	for _, row := range resp.Values {
		if len(row) > 0 && getStringFromCellByIndex(row, 0) == username {
			user := &domain.User{
				Username:     getStringFromCellByIndex(row, 0),
				PasswordHash: getStringFromCellByIndex(row, 1),
				Role:         getStringFromCellByIndex(row, 2),
				NamaLengkap:  getStringFromCellByIndex(row, 3),
				SiswaNISN:    getStringFromCellByIndex(row, 4), // <-- TAMBAHKAN INI
			}
			return user, nil
		}
	}
	return nil, nil
}

func (r *userRepository) Save(ctx context.Context, user *domain.User) error {
	log.Println("--- FUNGSI REPOSITORY SAVE (APPEND) DIPANGGIL ---")
	// Tentukan sheet dan baris data baru yang akan ditambahkan
	writeRange := "DataPengguna"
	var values [][]interface{}
	row := []interface{}{user.Username, user.PasswordHash, user.Role}
	values = append(values, row)

	// Siapkan data untuk API
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Panggil API untuk menambahkan baris baru (Append)
	_, err := r.db.Spreadsheets.Values.Append(r.spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("Gagal menyimpan data ke sheet: %v", err)
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, currentUsername string, user *domain.User) error {
	log.Println("--- FUNGSI REPOSITORY UPDATE DIPANGGIL ---")
	// 1. Baca semua data untuk menemukan nomor baris yang benar
	readRange := "DataPengguna!A:C" // Baca semua kolom untuk dapat nomor baris
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return err
	}

	rowIndex := -1
	for i, row := range resp.Values {
		if len(row) > 0 && row[0].(string) == currentUsername {
			// Kita temukan barisnya! Ingat bahwa spreadsheet mulai dari 1, dan array dari 0.
			// Jika data mulai dari A2, maka baris ke-i di array adalah baris ke i+2 di sheet.
			rowIndex = i + 2
			break
		}
	}

	if rowIndex == -1 {
		return errors.New("tidak dapat menemukan pengguna untuk diupdate")
	}

	// 2. Siapkan data baru dan panggil API Update
	updateRange := fmt.Sprintf("DataPengguna!A%d:C%d", rowIndex, rowIndex)
	var values [][]interface{}
	row := []interface{}{user.Username, user.PasswordHash, user.Role}
	values = append(values, row)

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err = r.db.Spreadsheets.Values.Update(r.spreadsheetId, updateRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("Gagal mengupdate data di sheet: %v", err)
		return err
	}

	return nil
}
