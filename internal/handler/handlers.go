package handler

import (
	"github.com/luganova-first/shortener/internal/model"
	"github.com/luganova-first/shortener/internal/service"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// Хендлер сокращения url
func SetShort(storage *model.Storage, baseURL string) http.HandlerFunc {
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

		// Проверяем, что URL валиден
		if _, err := url.ParseRequestURI(targetValue); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		shortURL, err := service.SetData(baseURL, targetValue, storage)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	}
}

// Хендлер получения полного url по сокращённой ссылке
func GetShort(storage *model.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortID := chi.URLParam(req, "shortID")

		fullURL := service.GetData(shortID, storage)

		if fullURL == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Location", fullURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte(fullURL))
	}
}
