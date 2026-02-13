package service

import (
	"crypto/rand"
	"fmt"
	"net/url"

	"github.com/luganova-first/shortener/internal/model"
)

// Генератор хеша сокращения
func cryptoRandomString(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(bytes), nil
}

func SetData(baseURL string, targetValue string, storage *model.Storage) (string, error) {
	// Ищем body запроса в значениях уже сокращённых
	cryptoString := storage.GetFull(targetValue)

	if cryptoString == "" {
		// Если не нашли, генерируем новое сокращение и записываем его в Shorted и в Full
		for i := 1; i < 5; i++ {
			str, err := cryptoRandomString(8)
			if err != nil {
				return "", err
			}

			// Если такого ключа ещё нет в Shorted
			// то записываем новое сокращение в Shorted и в Full и цикл закончится.
			// Если такой ключ уже есть, цикл повторится и сгенерируется новое значение ключа
			if storage.Shorted[str] == "" {
				cryptoString = str
				storage.Shorted[cryptoString] = targetValue
				storage.Full[targetValue] = cryptoString
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

func GetData(shortID string, storage *model.Storage) string {
	return storage.GetShort(shortID)
}
