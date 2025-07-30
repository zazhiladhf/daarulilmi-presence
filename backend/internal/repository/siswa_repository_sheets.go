// file: internal/repository/siswa_repository_sheets.go (Versi Final & Aman)
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"daarulilmi-presence/internal/domain" // Pastikan nama modul sudah benar

	"google.golang.org/api/sheets/v4"
)

// Fungsi bantuan untuk mengambil string dari sel dengan aman
func getStringFromCell(cell interface{}) string {
	if cell == nil {
		return ""
	}
	return fmt.Sprintf("%v", cell)
}

// Fungsi bantuan untuk mengambil sel dari baris dengan aman
func getStringFromCellByIndex(row []interface{}, index int) string {
	if index < len(row) {
		return getStringFromCell(row[index])
	}
	return ""
}

type siswaRepository struct {
	db            *sheets.Service
	spreadsheetId string
}

func NewSiswaRepository(db *sheets.Service, spreadsheetId string) domain.SiswaRepository {
	return &siswaRepository{db, spreadsheetId}
}

// --- FUNGSI FindAll DIPERBAIKI MENJADI LEBIH AMAN ---
func (r *siswaRepository) FindAll(ctx context.Context) ([]domain.Siswa, error) {
	readRange := "DataSiswa!A2:K"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	var siswaList []domain.Siswa
	for _, row := range resp.Values {
		nisn := getStringFromCellByIndex(row, 0)
		if nisn == "" {
			continue
		}
		// if len(row) == 0 {
		// 	continue
		// }

		siswa := domain.Siswa{
			NISN:             nisn,
			NamaLengkap:      getStringFromCellByIndex(row, 1),
			Kelas:            getStringFromCellByIndex(row, 2),
			NamaOrangTua:     getStringFromCellByIndex(row, 8),
			NomorTeleponOrtu: getStringFromCellByIndex(row, 3),
			EmailOrtu:        getStringFromCellByIndex(row, 4),
		}
		siswaList = append(siswaList, siswa)
	}
	return siswaList, nil
}

// --- FUNGSI FindByNISN DIPERBAIKI MENJADI LEBIH AMAN ---
func (r *siswaRepository) FindByNISN(ctx context.Context, nisn string) (*domain.Siswa, error) {
	// ... (Fungsi FindByNISN Anda yang sudah aman)
	readRange := "DataSiswa!A2:K"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	for _, row := range resp.Values {
		if len(row) > 0 && getStringFromCellByIndex(row, 0) == nisn {
			siswa := &domain.Siswa{
				NISN:             getStringFromCellByIndex(row, 0),
				NamaLengkap:      getStringFromCellByIndex(row, 1),
				Kelas:            getStringFromCellByIndex(row, 2),
				NamaOrangTua:     getStringFromCellByIndex(row, 8),
				NomorTeleponOrtu: getStringFromCellByIndex(row, 3),
				EmailOrtu:        getStringFromCellByIndex(row, 4),
			}
			return siswa, nil
		}
	}
	return nil, nil
}

// --- FUNGSI Save (tetap sama, sudah aman) ---
func (r *siswaRepository) Save(ctx context.Context, siswa *domain.Siswa) error {
	writeRange := "DataSiswa"
	var values [][]interface{}
	row := []interface{}{siswa.NISN, siswa.NamaLengkap, siswa.Kelas, siswa.NomorTeleponOrtu, siswa.EmailOrtu}
	values = append(values, row)

	valueRange := &sheets.ValueRange{Values: values}
	_, err := r.db.Spreadsheets.Values.Append(r.spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("Gagal menyimpan data siswa ke sheet: %v", err)
		return err
	}
	return nil
}

// --- FUNGSI BARU UNTUK UPDATE SISWA ---
func (r *siswaRepository) Update(ctx context.Context, nisn string, siswa *domain.Siswa) error {
	readRange := "DataSiswa!A2:E"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return err
	}

	rowIndex := -1
	for i, row := range resp.Values {
		if len(row) > 0 && getStringFromCellByIndex(row, 0) == nisn {
			rowIndex = i + 2 // Baris di sheet = indeks array + 2 (karena data mulai dari A2)
			break
		}
	}

	if rowIndex == -1 {
		return errors.New("NISN tidak ditemukan untuk diupdate")
	}

	updateRange := fmt.Sprintf("DataSiswa!A%d:E%d", rowIndex, rowIndex)
	var values [][]interface{}
	row := []interface{}{siswa.NISN, siswa.NamaLengkap, siswa.Kelas, siswa.NomorTeleponOrtu, siswa.EmailOrtu}
	values = append(values, row)

	valueRange := &sheets.ValueRange{Values: values}
	_, err = r.db.Spreadsheets.Values.Update(r.spreadsheetId, updateRange, valueRange).ValueInputOption("RAW").Do()
	return err
}

// --- FUNGSI BARU UNTUK DELETE SISWA ---
func (r *siswaRepository) Delete(ctx context.Context, nisn string) error {
	// Implementasi delete di Google Sheets agak rumit.
	// Cara termudah adalah dengan MENGHAPUS ISI BARIS, bukan barisnya.
	// Implementasi yang lebih canggih akan menggunakan BatchUpdate API.

	// Untuk saat ini, kita akan gunakan cara yang sama seperti Update, tapi dengan nilai kosong.
	readRange := "DataSiswa!A2:A"
	resp, err := r.db.Spreadsheets.Values.Get(r.spreadsheetId, readRange).Do()
	if err != nil {
		return err
	}

	rowIndex := -1
	for i, row := range resp.Values {
		if len(row) > 0 && getStringFromCellByIndex(row, 0) == nisn {
			rowIndex = i + 2
			break
		}
	}

	if rowIndex == -1 {
		return errors.New("NISN tidak ditemukan untuk dihapus")
	}

	clearRange := fmt.Sprintf("DataSiswa!A%d:E%d", rowIndex, rowIndex)
	_, err = r.db.Spreadsheets.Values.Clear(r.spreadsheetId, clearRange, &sheets.ClearValuesRequest{}).Do()
	return err
}
