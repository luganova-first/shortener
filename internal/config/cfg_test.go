package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig_DefaultValues(t *testing.T) {
	// Сбрасываем флаги перед каждым тестом
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Очищаем переменные окружения
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	// Сохраняем оригинальные аргументы и восстанавливаем после теста
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем пустые аргументы (без флагов)
	os.Args = []string{"cmd"}

	cfg := NewConfig()

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
}

func TestNewConfig_WithFlags(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем аргументы с флагами
	os.Args = []string{"cmd", "-a", "127.0.0.1:9090", "-b", "https://short.url"}

	cfg := NewConfig()

	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddress)
	assert.Equal(t, "https://short.url", cfg.BaseURL)
}

func TestNewConfig_WithEnvVars(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", "0.0.0.0:3000")
	os.Setenv("BASE_URL", "https://example.com")
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
	}()

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Флаги должны игнорироваться, так как есть переменные окружения
	os.Args = []string{"cmd", "-a", "127.0.0.1:8080", "-b", "http://ignored.url"}

	cfg := NewConfig()

	assert.Equal(t, "0.0.0.0:3000", cfg.ServerAddress)
	assert.Equal(t, "https://example.com", cfg.BaseURL)
}

func TestNewConfig_MixedEnvAndFlags(t *testing.T) {
	tests := []struct {
		name           string
		envServer      string
		envBase        string
		flagServer     string
		flagBase       string
		expectedServer string
		expectedBase   string
	}{
		{
			name:           "only env SERVER_ADDRESS",
			envServer:      "192.168.1.1:8000",
			envBase:        "",
			flagServer:     "",
			flagBase:       "",
			expectedServer: "192.168.1.1:8000",
			expectedBase:   "http://192.168.1.1:8000",
		},
		{
			name:           "only env BASE_URL",
			envServer:      "",
			envBase:        "https://custom.url",
			flagServer:     "",
			flagBase:       "",
			expectedServer: "localhost:8080",
			expectedBase:   "https://custom.url",
		},
		{
			name:           "env overrides flag for SERVER_ADDRESS",
			envServer:      "10.0.0.1:9000",
			envBase:        "",
			flagServer:     "172.16.0.1:8080",
			flagBase:       "http://flag.url",
			expectedServer: "10.0.0.1:9000",
			expectedBase:   "http://flag.url",
		},
		{
			name:           "env overrides flag for BASE_URL",
			envServer:      "",
			envBase:        "https://env.url",
			flagServer:     "localhost:8080",
			flagBase:       "http://flag.url",
			expectedServer: "localhost:8080",
			expectedBase:   "https://env.url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Устанавливаем переменные окружения
			if tt.envServer != "" {
				os.Setenv("SERVER_ADDRESS", tt.envServer)
			} else {
				os.Unsetenv("SERVER_ADDRESS")
			}

			if tt.envBase != "" {
				os.Setenv("BASE_URL", tt.envBase)
			} else {
				os.Unsetenv("BASE_URL")
			}

			defer func() {
				os.Unsetenv("SERVER_ADDRESS")
				os.Unsetenv("BASE_URL")
			}()

			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Формируем аргументы командной строки
			args := []string{"cmd"}
			if tt.flagServer != "" {
				args = append(args, "-a", tt.flagServer)
			}
			if tt.flagBase != "" {
				args = append(args, "-b", tt.flagBase)
			}
			os.Args = args

			cfg := NewConfig()

			assert.Equal(t, tt.expectedServer, cfg.ServerAddress)
			assert.Equal(t, tt.expectedBase, cfg.BaseURL)
		})
	}
}

func TestNewConfig_BaseURLFormation(t *testing.T) {
	tests := []struct {
		name         string
		serverAddr   string
		flagBase     string
		expectedBase string
	}{
		{
			name:         "server without protocol",
			serverAddr:   "localhost:8080",
			flagBase:     "",
			expectedBase: "http://localhost:8080",
		},
		{
			name:         "server with http protocol",
			serverAddr:   "http://localhost:8080",
			flagBase:     "",
			expectedBase: "http://localhost:8080",
		},
		{
			name:         "server with https protocol",
			serverAddr:   "https://example.com:8443",
			flagBase:     "",
			expectedBase: "https://example.com:8443",
		},
		{
			name:         "flag base URL overrides automatic formation",
			serverAddr:   "localhost:8080",
			flagBase:     "https://custom.url",
			expectedBase: "https://custom.url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")

			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			args := []string{"cmd", "-a", tt.serverAddr}
			if tt.flagBase != "" {
				args = append(args, "-b", tt.flagBase)
			}
			os.Args = args

			cfg := NewConfig()

			assert.Equal(t, tt.serverAddr, cfg.ServerAddress)
			assert.Equal(t, tt.expectedBase, cfg.BaseURL)
		})
	}
}

func TestNewConfig_EnvVarsPriority(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Устанавливаем обе переменные окружения
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "https://env.url")
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
	}()

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем флаги с другими значениями
	os.Args = []string{"cmd", "-a", "flag:9090", "-b", "http://flag.url"}

	cfg := NewConfig()

	// Должны использоваться значения из переменных окружения
	assert.Equal(t, "env:8080", cfg.ServerAddress)
	assert.Equal(t, "https://env.url", cfg.BaseURL)
}

func TestNewConfig_EmptyEnvAndFlags(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Пустые флаги (используем значения по умолчанию)
	os.Args = []string{"cmd", "-a", "", "-b", ""}

	cfg := NewConfig()

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr string
	}{
		{
			name: "valid config",
			config: &Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			expectedErr: "",
		},
		{
			name: "valid config with custom values",
			config: &Config{
				ServerAddress: "0.0.0.0:3000",
				BaseURL:       "https://example.com",
			},
			expectedErr: "",
		},
		{
			name: "empty server address",
			config: &Config{
				ServerAddress: "",
				BaseURL:       "http://localhost:8080",
			},
			expectedErr: "server address cannot be empty",
		},
		{
			name: "empty base URL",
			config: &Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "",
			},
			expectedErr: "base URL cannot be empty",
		},
		{
			name: "both fields empty",
			config: &Config{
				ServerAddress: "",
				BaseURL:       "",
			},
			expectedErr: "server address cannot be empty",
		},
		{
			name: "server address with spaces only",
			config: &Config{
				ServerAddress: "   ",
				BaseURL:       "http://localhost:8080",
			},
			expectedErr: "server address cannot be empty",
		},
		{
			name: "base URL with spaces only",
			config: &Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "   ",
			},
			expectedErr: "base URL cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
