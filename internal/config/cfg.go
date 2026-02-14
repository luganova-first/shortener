package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config хранит конфигурацию сервера
type Config struct {
	ServerAddress string // Адрес запуска HTTP-сервера
	BaseURL       string // Базовый адрес для сокращенных URL
}

// NewConfig создает и инициализирует конфигурацию из аргументов командной строки
func NewConfig() *Config {
	cfg := &Config{}

	defaultServerAddr := "localhost:8080"

	// Берём адреса из переменных окружения
	cfg.ServerAddress = os.Getenv("SERVER_ADDRESS")
	cfg.BaseURL = os.Getenv("BASE_URL")

	// Если адресов нет, определяем флаги
	if cfg.ServerAddress == "" || cfg.BaseURL == "" {
		serverAddr := flag.String("a", defaultServerAddr, "Адрес запуска HTTP-сервера")
		baseURL := flag.String("b", "", "Базовый адрес результирующего сокращённого URL")
		flag.Parse()

		if cfg.ServerAddress == "" {
			// Если адрес запуска HTTP-сервера никак не указан, ставим по умолчанию
			if *serverAddr == "" {
				*serverAddr = defaultServerAddr
			}

			// Устанавливаем значения
			cfg.ServerAddress = *serverAddr
		}

		if cfg.BaseURL == "" {
			// Если базовый URL никак не указан, формируем его из адреса сервера
			if *baseURL == "" {
				// Добавляем http:// если его нет в адресе сервера
				if !strings.HasPrefix(cfg.ServerAddress, "http://") && !strings.HasPrefix(cfg.ServerAddress, "https://") {
					*baseURL = "http://" + cfg.ServerAddress
				} else {
					*baseURL = cfg.ServerAddress
				}
			}
			cfg.BaseURL = *baseURL
		}
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
