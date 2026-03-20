package repository

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/pressly/goose/v3"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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

var ErrAlreadyExists = errors.New("url already exists")

var embedMigrations embed.FS

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
		return s, fmt.Errorf("failed to read file with shorts: %w", err)
	}
	defer fileReader.Close()

	var storageRows []StorageFile

	// Создание JSON декодера
	decoder := json.NewDecoder(fileReader.file)

	// Чтение и парсинг JSON
	err = decoder.Decode(&storageRows)
	if err != nil && err != io.EOF {
		return s, fmt.Errorf("failed to read file with shorts: %w", err)
	}

	for _, row := range storageRows {
		s.Shorted[row.ShortURL] = row.OriginalURL
		s.Full[row.OriginalURL] = row.ShortURL
	}

	return s, nil
}

func FillStorageFromDB(s *model.Storage, cfg *config.Config) (*model.Storage, error) {
	db, err := DB(cfg)
	if err != nil {
		return s, err
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), "SELECT shorted, full_url FROM shorts")
	if err != nil {
		return s, err
	}
	defer rows.Close()

	for rows.Next() {
		var short string
		var fullURL string

		err = rows.Scan(&short, &fullURL)
		if err != nil {
			return s, err
		}

		s.Shorted[short] = fullURL
		s.Full[fullURL] = short
	}

	return s, nil
}

func GetShortFromDB(cfg *config.Config, shortID string) string {
	db, err := DB(cfg)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer db.Close()

	row := db.QueryRowContext(context.Background(), "SELECT full_url FROM shorts WHERE shorted = $1", shortID)

	var fullURL string
	err = row.Scan(&fullURL)
	if err != nil {
		log.Println(err)
		return ""
	}

	return fullURL
}

func GetFullFromDB(cfg *config.Config, targetValue string) string {
	db, err := DB(cfg)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer db.Close()

	row := db.QueryRowContext(context.Background(), "SELECT shorted FROM shorts WHERE full_url = $1", targetValue)

	var shorted string
	err = row.Scan(&shorted)
	if err != nil {
		log.Println(err)
		return ""
	}

	return shorted
}

func UpDBMigrations(db *sql.DB) error {
	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("failed up migrations: %w", err)
	}
	defer db.Close()

	return nil
}

func ClearDB(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE TABLE shorts")
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

func FillStorage(s *model.Storage, cfg *config.Config) (*model.Storage, error) {
	switch {
	case cfg.DBconnStr != "":
		db, err := DB(cfg)
		if err != nil {
			return s, err
		}

		err = UpDBMigrations(db)
		if err != nil {
			return s, err
		}

		return s, nil
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

	return db, nil
}

func InsertNewShort(db *sql.DB, shorted string, fullURL string) error {
	_, err := db.Exec("INSERT INTO shorts (shorted, full_url) VALUES ($1, $2)", shorted, fullURL)
	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// Нарушение уникальности - такой URL уже существует
			return fmt.Errorf("%w: %s", ErrAlreadyExists, fullURL)
		}
		return err
	}

	return nil
}

func WriteStorageToDB(cfg *config.Config, key string, value string) error {
	db, err := DB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	err = InsertNewShort(db, key, value)
	if err != nil {
		return err
	}

	return nil
}

func WriteBulkStorageToDB(cfg *config.Config, data map[string]string) error {
	db, err := DB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Создаем слайс для значений и аргументов
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*2)

	i := 0
	for key, value := range data {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)",
			i*2+1, i*2+2))
		valueArgs = append(valueArgs, key, value)
		i++
	}

	query := fmt.Sprintf("INSERT INTO shorts (shorted, full_url) VALUES %s", strings.Join(valueStrings, ","))

	_, err = db.Exec(query, valueArgs...)

	return err
}
