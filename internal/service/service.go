package service

import (
	"fmt"
	"net/url"

	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/luganova-first/shortener/internal/repository"
	"github.com/luganova-first/shortener/pkg"
)

// ShortenerService представляет сервис для работы с сокращением ссылок
type ShortenerService struct {
	storage *model.Storage
	config  *config.Config
}

// NewShortenerService создает новый экземпляр сервиса
func NewShortenerService(storage *model.Storage, cfg *config.Config) *ShortenerService {
	return &ShortenerService{
		storage: storage,
		config:  cfg,
	}
}

// MakeShort создаёт новое сокращение и записывает в мапу
func (s *ShortenerService) MakeShort(targetValue string) (string, error) {
	var cryptoString string

	// Ищем body запроса в значениях уже сокращённых
	if s.config.DBconnStr != "" {
		cryptoString = repository.GetFullFromDB(s.config, targetValue)
	} else {
		cryptoString = s.storage.GetFull(targetValue)
	}

	if cryptoString == "" {
		// Если не нашли, генерируем новое сокращение и записываем его в Shorted и в Full
		for i := 1; i < 5; i++ {
			str, err := pkg.CryptoRandomString(8)
			if err != nil {
				return "", err
			}

			// Если такого ключа ещё нет в Shorted
			// то записываем новое сокращение в Shorted и в Full и цикл закончится.
			// Если такой ключ уже есть, цикл повторится и сгенерируется новое значение ключа
			if s.storage.Shorted[str] == "" {
				cryptoString = str

				break
			}
		}

		if cryptoString == "" {
			return "", fmt.Errorf("no cryptoString")
		}
	}

	s.storage.Shorted[cryptoString] = targetValue
	s.storage.Full[targetValue] = cryptoString

	return cryptoString, nil
}

// SetData сохраняет данные и возвращает сокращенный URL
func (s *ShortenerService) SetData(targetValue string) (string, error) {
	short, err := s.MakeShort(targetValue)

	shortURL, err := s.GetShortURL(short)
	if err != nil {
		return "", err
	}

	switch {
	case s.config.DBconnStr != "":
		err = repository.WriteStorageToDB(s.config, short, targetValue)
		if err != nil {
			return shortURL, err
		}
	case s.config.StorageFileName != "":
		err = repository.WriteStorageToFile(s.storage, s.config)
		if err != nil {
			return shortURL, err
		}
	}

	return shortURL, nil
}

// SetBulkData сохраняет множественные данные
func (s *ShortenerService) SetBulkData(data map[string]string) error {
	switch {
	case s.config.DBconnStr != "":
		err := repository.WriteBulkStorageToDB(s.config, data)
		if err != nil {
			return err
		}
	case s.config.StorageFileName != "":
		err := repository.WriteStorageToFile(s.storage, s.config)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetData получает данные по короткому идентификатору
func (s *ShortenerService) GetData(shortID string) string {
	if s.config.DBconnStr != "" {
		return repository.GetShortFromDB(s.config, shortID)
	} else {
		return s.storage.GetShort(shortID)
	}
}

// GetData получает данные по короткому идентификатору
func (s *ShortenerService) GetShortURL(short string) (string, error) {
	baseURL := s.config.BaseURL

	shortURL, err := url.JoinPath(baseURL, short)
	if err != nil {
		return "", err
	}

	return shortURL, nil
}
