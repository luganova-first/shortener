package handler

import (
	"bytes"
	"encoding/json"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/luganova-first/shortener/internal/repository"
	"github.com/luganova-first/shortener/internal/service"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type inputJSONData struct {
	URL string `json:"url,omitempty"`
}

type outJSONData struct {
	Result string `json:"result"`
}

// Хендлер сокращения url
func SetShort(storage *model.Storage, cfg *config.Config) http.HandlerFunc {
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

		s := service.NewShortenerService(storage, cfg)

		shortURL, err := s.SetData(targetValue)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	}
}

// Хендлер сокращения url, который будет принимать в теле запроса JSON-объект {"url":"<some_url>"}
// и возвращать в ответ объект {"result":"<short_url>"}.
func JSONShort(storage *model.Storage, cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// POST запрос должен быть с Content-Type `application/json`
		if req.Header.Get("Content-Type") != "application/json" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var jsonData inputJSONData
		var buf bytes.Buffer

		// читаем тело запроса
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		// десериализуем JSON в url
		if err = json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		targetValue := jsonData.URL

		// Проверяем, что URL валиден
		if _, err := url.ParseRequestURI(targetValue); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		s := service.NewShortenerService(storage, cfg)

		shortURL, err := s.SetData(targetValue)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		resultData := outJSONData{
			Result: shortURL,
		}

		resp, err := json.MarshalIndent(resultData, "", " ")
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resp)
	}
}

// Хендлер получения полного url по сокращённой ссылке
func GetShort(storage *model.Storage, cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortID := chi.URLParam(req, "shortID")

		s := service.NewShortenerService(storage, cfg)

		fullURL := s.GetData(shortID)

		if fullURL == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Location", fullURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte(fullURL))
	}
}

// Хендлер получения подключения к базе данных
func GetDB(cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		_, err := repository.DB(cfg)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}
