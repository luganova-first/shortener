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

// SetData сохраняет данные и возвращает сокращенный URL
func (s *ShortenerService) SetData(targetValue string) (string, error) {
	baseURL := s.config.BaseURL

	// Ищем body запроса в значениях уже сокращённых
	cryptoString := s.storage.GetFull(targetValue)

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
				s.storage.Shorted[cryptoString] = targetValue
				s.storage.Full[targetValue] = cryptoString

				// Перезаписываем файл с данными Storage
				if err := repository.WriteStorageToFile(s.storage, s.config); err != nil {
					return "", err
				}

				break
			}
		}

		if cryptoString == "" {
			return "", fmt.Errorf("no cryptoString")
		}
	}

	shortURL, err := url.JoinPath(baseURL, cryptoString)
	if err != nil {
		return "", err
	}

	return shortURL, nil
}

// GetData получает данные по короткому идентификатору
func (s *ShortenerService) GetData(shortID string) string {
	return s.storage.GetShort(shortID)
}
