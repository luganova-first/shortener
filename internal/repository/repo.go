package repository

import (
	"encoding/json"
	"flag"
	"github.com/luganova-first/shortener/internal/model"
	"log"
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

// Получаем имя файла для записи данных.
// Сначала из переменной окружения "FILE_STORAGE_PATH",
// если её нет, то из флага -f, если и его нет, то ставим по умолчанию
func GetStorageFileName() string {
	defaultStorageFileName := "Storage.txt"

	storageFileName := os.Getenv("FILE_STORAGE_PATH")

	if storageFileName == "" {
		fileName := flag.String("f", defaultStorageFileName, "Файл для записи Storage")
		flag.Parse()

		if storageFileName == "" {
			if *fileName == "" {
				*fileName = defaultStorageFileName
			}

			storageFileName = *fileName
		}
	}

	return storageFileName
}

func FillStorageFromFile(s *model.Storage) *model.Storage {
	storageFileName := GetStorageFileName()

	fileReader, err := NewFileReader(storageFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fileReader.Close()

	var storageRows []StorageFile

	// Создание JSON декодера
	decoder := json.NewDecoder(fileReader.file)

	// Чтение и парсинг JSON
	err = decoder.Decode(&storageRows)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range storageRows {
		s.Shorted[row.ShortURL] = row.OriginalURL
		s.Full[row.OriginalURL] = row.ShortURL
	}

	return s
}

func WriteStorageToFile(s *model.Storage) {
	storageFileName := GetStorageFileName()

	fileWriter, err := NewFileWriter(storageFileName)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
}
