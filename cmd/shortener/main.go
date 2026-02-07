package main

import (
	"crypto/rand"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

// Хранилище сокращений, ключ -- хеш сокращения, значение -- сокращаемый URL
type Shorted map[string]string

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

// Хендлер сокращения url
func SetShort(shorts Shorted) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// POST запрос должен быть с Content-Type `text/plain`
		if req.Header.Get("Content-Type") != "text/plain" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		defer req.Body.Close()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		if string(body) == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		targetValue := string(body)

		// Ищем body запроса в значениях уже сокращённых
		var cryptoString string
		for key, value := range shorts {
			if value == targetValue {
				cryptoString = key
			}
		}

		if cryptoString == "" {
			// Если не нашли, генерируем новое сокращение и записываем его в shorts
			str, err := cryptoRandomString(8)
			if err != nil {
				panic(err)
			}

			cryptoString = str

			shorts[cryptoString] = targetValue
		}

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte("http://" + req.Host + "/" + cryptoString))
	}
}

// Хендлер получения полного url по сокращённой ссылке
func GetShort(shorts Shorted) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortID := chi.URLParam(req, "shortID")
		// GET запрос, пытаемся найти сокращение в shorts по ключу
		if shorts[shortID] == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Location", shorts[shortID])
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte(shorts[shortID]))
	}
}

func main() {
	shorts := make(Shorted)

	r := chi.NewRouter()

	r.Post("/", SetShort(shorts))
	r.Get("/{shortID}", GetShort(shorts))

	r.MethodNotAllowed(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
		return
	})

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
