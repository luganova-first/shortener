package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config хранит конфигурацию сервера
type Config struct {
	ServerAddress   string // Адрес запуска HTTP-сервера
	BaseURL         string // Базовый адрес для сокращенных URL
	StorageFileName string // Имя файла для записи данных Storage
	DBconnStr       string // Строка с адресом подключения к БД
}

// NewConfig создает и инициализирует конфигурацию из аргументов командной строки
func NewConfig() *Config {
	cfg := &Config{}

	defaultServerAddr := "localhost:8080"
	defaultStorageFileName := "Storage.txt"

	// Парсим флаги
	serverAddr := flag.String("a", defaultServerAddr, "Адрес запуска HTTP-сервера")
	baseURL := flag.String("b", "", "Базовый адрес результирующего сокращённого URL")
	fileName := flag.String("f", defaultStorageFileName, "Файл для записи Storage")
	dbStr := flag.String("d", "", "Строка с адресом подключения к БД")
	flag.Parse()

	var ok bool

	// Берём адрес из переменной окружения
	cfg.ServerAddress, ok = os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		// Если адреса нет, определяем флаги
		if *serverAddr == "" {
			*serverAddr = defaultServerAddr
		}

		// Устанавливаем значение
		cfg.ServerAddress = *serverAddr
	}

	// Берём адрес из переменной окружения
	cfg.BaseURL, ok = os.LookupEnv("BASE_URL")
	if !ok {
		// Если адреса нет, определяем флаги
		if *baseURL == "" {
			// Добавляем http:// если его нет в адресе сервера
			if !strings.HasPrefix(cfg.ServerAddress, "http://") && !strings.HasPrefix(cfg.ServerAddress, "https://") {
				*baseURL = "http://" + cfg.ServerAddress
			} else {
				*baseURL = cfg.ServerAddress
			}
		}

		// Устанавливаем значение
		cfg.BaseURL = *baseURL
	}

	// Берём имя файла из переменной окружения
	cfg.StorageFileName, ok = os.LookupEnv("FILE_STORAGE_PATH")
	if !ok {
		// Если имени файла нет, определяем флаги
		if *fileName == "" {
			*fileName = defaultStorageFileName
		}

		// Устанавливаем значения
		cfg.StorageFileName = *fileName
	}

	// Берём настройки БД из переменной окружения
	cfg.DBconnStr, ok = os.LookupEnv("DATABASE_DSN")
	if !ok {
		// Берём настроейк БД нет, определяем флаги
		cfg.DBconnStr = *dbStr
	}

	return cfg
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	c.ServerAddress = strings.ReplaceAll(c.ServerAddress, " ", "")
	c.BaseURL = strings.ReplaceAll(c.BaseURL, " ", "")

	if c.ServerAddress == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	return nil
}
