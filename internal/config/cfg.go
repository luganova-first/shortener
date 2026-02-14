package config

import (
	"flag"
	"fmt"
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

	// Определяем флаги
	serverAddr := flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера")
	baseURL := flag.String("b", "", "Базовый адрес результирующего сокращённого URL")

	flag.Parse()

	// Устанавливаем значения из флагов
	cfg.ServerAddress = *serverAddr

	// Если базовый URL не указан, формируем его из адреса сервера
	if *baseURL == "" {
		// Добавляем http:// если его нет в адресе сервера
		if !strings.HasPrefix(cfg.ServerAddress, "http://") && !strings.HasPrefix(cfg.ServerAddress, "https://") {
			*baseURL = "http://" + cfg.ServerAddress
		} else {
			*baseURL = cfg.ServerAddress
		}
	}
	cfg.BaseURL = *baseURL

	return cfg
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.ServerAddress == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	return nil
}
