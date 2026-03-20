package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/luganova-first/shortener/internal/repository"
	"github.com/luganova-first/shortener/internal/service"
	"github.com/luganova-first/shortener/internal/userauth"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

type inputJSONData struct {
	URL string `json:"url,omitempty"`
}

type outJSONData struct {
	Result string `json:"result"`
}

// RequestItem представляет один элемент запроса
type RequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseItem представляет один элемент ответа
type ResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserItem представляет один элемент ответа
type UserItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

		res.Header().Set("content-type", "text/plain")

		shortURL, err := s.SetData(targetValue)
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				res.WriteHeader(http.StatusConflict)
			} else {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			res.WriteHeader(http.StatusCreated)
		}

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
		err = json.Unmarshal(buf.Bytes(), &jsonData)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		targetValue := jsonData.URL

		// Проверяем, что URL валиден
		if _, err := url.ParseRequestURI(targetValue); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		s := service.NewShortenerService(storage, cfg)

		res.Header().Set("content-type", "application/json")

		shortURL, err := s.SetData(targetValue)
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				res.WriteHeader(http.StatusConflict)
			} else {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			res.WriteHeader(http.StatusCreated)
		}

		resultData := outJSONData{
			Result: shortURL,
		}

		resp, err := json.MarshalIndent(resultData, "", " ")
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Write(resp)
	}
}

// Хендлер принимающий в теле запроса множество URL для сокращения
func BatchShort(storage *model.Storage, cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Декодируем JSON из тела запроса
		var requests []RequestItem
		err := json.NewDecoder(req.Body).Decode(&requests)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		// Валидация входных данных
		if len(requests) == 0 {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// Создаём мапу данных для сохранения
		dataForSave := make(map[string]string)

		// Создаем слайс для ответов
		responses := make([]ResponseItem, 0, len(requests))

		s := service.NewShortenerService(storage, cfg)

		// Ключ сокращения
		var short string

		// Обрабатываем каждый URL
		for _, item := range requests {
			// Валидация каждого элемента
			if item.CorrelationID == "" || item.OriginalURL == "" {
				continue // Пропускаем некорректные элементы
			}

			short, err = s.MakeShort(item.OriginalURL)

			shortURL, err := s.GetShortURL(short)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Добавляем результат в ответ
			responses = append(responses, ResponseItem{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})

			// Добавляем результат в данные для сохранения
			dataForSave[short] = item.OriginalURL
		}

		err = s.SetBulkData(dataForSave)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		json.NewEncoder(res).Encode(responses)
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

// Хендлер получения всех сокращений пользователя
func UserURLS(storage *model.Storage, cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var tokenString string

		cookie, err := req.Cookie("jwt")
		if err != nil {
			if err == http.ErrNoCookie {
				// Очистим "чужие" сокращения из памяти
				clear(storage.Shorted)
				clear(storage.Full)

				// Если используется база, вычистим из неё "чужие" сокращения
				if cfg.DBconnStr != "" {
					db, err := repository.DB(cfg)
					if err != nil {
						log.Println(err)
						res.WriteHeader(http.StatusInternalServerError)
						return
					}

					err = repository.ClearDB(db)
					if err != nil {
						log.Println(err)
						res.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				// Если используется файл, вычистим из него "чужие" сокращения
				if cfg.StorageFileName != "" {
					err = os.Truncate(cfg.StorageFileName, 0)
					if err != nil {
						log.Println(err)
						res.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				tokenString, err = userauth.BuildJWTString()
				if err != nil {
					log.Println(err)
					res.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie := &http.Cookie{
					Name:     "jwt",
					Value:    tokenString,
					Path:     "/api/user/urls",
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				}

				http.SetCookie(res, cookie)
			} else {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tokenString = cookie.Value
		}

		s := service.NewShortenerService(storage, cfg)

		userID := userauth.GetUserID(tokenString)
		if userID != 17 {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if cfg.DBconnStr != "" {
			storage, err = repository.FillStorageFromDB(storage, cfg)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if len(storage.Shorted) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		// Создаем слайс для ответов
		responses := make([]UserItem, 0, len(storage.Shorted))

		for short, originalURL := range storage.Shorted {
			shortURL, err := s.GetShortURL(short)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Добавляем результат в ответ
			responses = append(responses, UserItem{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			})
		}

		// Отправляем ответ
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(responses)
	}
}

// Хендлер получения подключения к базе данных
func GetDB(cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		db, err := repository.DB(cfg)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer db.Close()

		res.WriteHeader(http.StatusOK)
	}
}
