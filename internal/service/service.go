package service

import (
    "crypto/rand"

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

func SetData(baseURL string, targetValue string, shorts model.Shorted) string {
    // Ищем body запроса в значениях уже сокращённых
    var cryptoString string
    for key, value := range shorts {
        if value == targetValue {
            cryptoString = key
            break
        }
    }

    if cryptoString == "" {
        // Если не нашли, генерируем новое сокращение и записываем его в shorts
        for {
            str, err := cryptoRandomString(8)
            if err != nil {
                panic(err)
            }

            // Если такого ключа ещё нет в shorts
            // то записываем новое сокращение в shorts и цикл закончится.
            // Если такой ключ уже есть, цикл повторится и сгенерируется новое значение ключа
            if shorts[str] == "" {
                cryptoString = str
                shorts[cryptoString] = targetValue
                break
            }
        }
    }

    return baseURL + "/" + cryptoString
}

func GetData(shortID string, shorts model.Shorted) string {
    return shorts[shortID]
}