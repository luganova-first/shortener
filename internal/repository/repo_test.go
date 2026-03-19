package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFillStorageFromFile_Success(t *testing.T) {
	// Создаем временный файл с тестовыми данными
	tmpFile, err := os.CreateTemp("", "storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Подготовка тестовых данных
	testData := []StorageFile{
		{
			ID:          "1",
			ShortURL:    "abc123",
			OriginalURL: "https://example.com/1",
		},
		{
			ID:          "2",
			ShortURL:    "def456",
			OriginalURL: "https://example.com/2",
		},
	}

	// Запись данных в файл
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)
	tmpFile.Close()

	// Создаем конфиг с путем к временному файлу
	cfg := &config.Config{
		StorageFileName: tmpFile.Name(),
	}

	// Создаем пустое хранилище
	storage := &model.Storage{
		Shorted: make(map[string]string),
		Full:    make(map[string]string),
	}

	// Вызываем тестируемую функцию
	result, err := FillStorageFromFile(storage, cfg)

	// Проверки
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://example.com/1", result.Shorted["abc123"])
	assert.Equal(t, "https://example.com/2", result.Shorted["def456"])
	assert.Equal(t, "abc123", result.Full["https://example.com/1"])
	assert.Equal(t, "def456", result.Full["https://example.com/2"])
}

func TestFillStorageFromFile_EmptyFile(t *testing.T) {
	// Создаем пустой временный файл
	tmpFile, err := os.CreateTemp("", "storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.Config{
		StorageFileName: tmpFile.Name(),
	}

	storage := &model.Storage{
		Shorted: make(map[string]string),
		Full:    make(map[string]string),
	}

	result, err := FillStorageFromFile(storage, cfg)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Shorted)
	assert.Empty(t, result.Full)
}

func TestWriteStorageToFile_Success(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Подготовка тестовых данных
	storage := &model.Storage{
		Shorted: map[string]string{
			"abc123": "https://example.com/1",
			"def456": "https://example.com/2",
		},
		Full: map[string]string{
			"https://example.com/1": "abc123",
			"https://example.com/2": "def456",
		},
	}

	cfg := &config.Config{
		StorageFileName: tmpFile.Name(),
	}

	// Вызываем тестируемую функцию
	err = WriteStorageToFile(storage, cfg)

	// Проверяем, что ошибки нет
	require.NoError(t, err)

	// Читаем файл и проверяем содержимое
	fileContent, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var savedData []StorageFile
	err = json.Unmarshal(fileContent, &savedData)
	require.NoError(t, err)

	// Проверяем количество записей
	assert.Len(t, savedData, 2)

	// Создаем мапу для удобства проверки
	savedMap := make(map[string]string)
	for _, item := range savedData {
		savedMap[item.ShortURL] = item.OriginalURL
	}

	assert.Equal(t, "https://example.com/1", savedMap["abc123"])
	assert.Equal(t, "https://example.com/2", savedMap["def456"])
}

func TestWriteStorageToFile_EmptyStorage(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Пустое хранилище
	storage := &model.Storage{
		Shorted: make(map[string]string),
		Full:    make(map[string]string),
	}

	cfg := &config.Config{
		StorageFileName: tmpFile.Name(),
	}

	err = WriteStorageToFile(storage, cfg)

	require.NoError(t, err)

	// Читаем файл и проверяем
	fileContent, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var savedData []StorageFile
	err = json.Unmarshal(fileContent, &savedData)
	require.NoError(t, err)

	assert.Empty(t, savedData)
}
