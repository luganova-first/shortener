package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"io"
	"os"
	"strconv"
)

// Формат данных в файле
type StorageFile struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileWriter struct {
	file *os.File
}

func NewFileWriter(filename string) (*FileWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file: file,
	}, nil
}

func (fw *FileWriter) Close() error {
	// закрываем файл
	return fw.file.Close()
}

type FileReader struct {
	file *os.File
}

func NewFileReader(filename string) (*FileReader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileReader{
		file: file,
	}, nil
}

func (fr *FileReader) Close() error {
	return fr.file.Close()
}

func FillStorageFromFile(s *model.Storage, cfg *config.Config) (*model.Storage, error) {
	storageFileName := cfg.StorageFileName

	fileReader, err := NewFileReader(storageFileName)
	if err != nil {
		return s, err
	}
	defer fileReader.Close()

	var storageRows []StorageFile

	// Создание JSON декодера
	decoder := json.NewDecoder(fileReader.file)

	// Чтение и парсинг JSON
	err = decoder.Decode(&storageRows)
	if err != nil && err != io.EOF {
		return s, err
	}

	for _, row := range storageRows {
		s.Shorted[row.ShortURL] = row.OriginalURL
		s.Full[row.OriginalURL] = row.ShortURL
	}

	return s, nil
}

func FillStorageFromDB(ctx context.Context, db *sql.DB, s *model.Storage) (*model.Storage, error) {
	defer db.Close()
	rows, err := db.QueryContext(ctx, "SELECT * FROM shorts")
	if err != nil {
		return s, err
	}
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var shorted string
		var fullURL string
		err = rows.Scan(&shorted, &fullURL)
		if err != nil {
			return s, err
		}
		s.Shorted[shorted] = fullURL
		s.Full[fullURL] = shorted
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return s, err
	}
	return s, nil
}

func FillStorage(s *model.Storage, cfg *config.Config) (*model.Storage, error) {
	switch {
	case cfg.DBconnStr != "":
		db, err := DB(cfg)
		if err != nil {
			return s, err
		}
		return FillStorageFromDB(context.Background(), db, s)
	case cfg.StorageFileName != "":
		return FillStorageFromFile(s, cfg)
	}

	return s, nil
}

func WriteStorageToFile(s *model.Storage, cfg *config.Config) error {
	storageFileName := cfg.StorageFileName

	fileWriter, err := NewFileWriter(storageFileName)
	if err != nil {
		return err
	}
	defer fileWriter.Close()

	i := 1
	var storageRows []StorageFile

	for key, value := range s.Shorted {
		id := strconv.Itoa(i)

		row := StorageFile{
			ID:          id,
			ShortURL:    key,
			OriginalURL: value,
		}

		storageRows = append(storageRows, row)
	}

	// Создаем encoder с отступами для красивого форматирования
	encoder := json.NewEncoder(fileWriter.file)
	encoder.SetIndent("", "  ") // два пробела для отступа

	// Записываем массив в файл
	err = encoder.Encode(storageRows)
	if err != nil {
		return err
	}

	return nil
}

func DB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DBconnStr)
	if err != nil {
		return db, err
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS shorts (
        shorted VARCHAR(8) NOT NULL DEFAULT '',
		full_url VARCHAR(250) NOT NULL DEFAULT ''
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return db, err
	}

	return db, nil
}

func InsertNewShort(db *sql.DB, shorted string, fullURL string) error {
	_, err := db.Exec("INSERT INTO shorts (shorted, full_url) VALUES ($1, $2)", shorted, fullURL)
	if err != nil {
		return err
	}
	return nil
}

func WriteStorageToDB(s *model.Storage, cfg *config.Config) error {
	db, err := DB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE TABLE shorts")
	if err != nil {
		return err
	}
	for key, value := range s.Shorted {
		err := InsertNewShort(db, key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func SaveStorage(s *model.Storage, cfg *config.Config) error {
	switch {
	case cfg.DBconnStr != "":
		return WriteStorageToDB(s, cfg)
	case cfg.StorageFileName != "":
		return WriteStorageToFile(s, cfg)
	}

	return nil
}
