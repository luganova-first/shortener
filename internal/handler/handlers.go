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

// Хендлер сокращения url
func SetShort(storage *model.Storage, cfg *config.Config, users *userauth.Users) http.HandlerFunc {
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

		var tokenString string
		cookie, err := req.Cookie("jwt")
		if err != nil {
			log.Println(err)
		} else {
			tokenString = cookie.Value
		}

		userID := userauth.GetUserID(tokenString)
		if userID < 0 {
			userID = req.Context().Value("userID").(int)
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

		userItem := userauth.UserItem{
			ShortURL:    shortURL,
			OriginalURL: targetValue,
		}

		users.UserURLs[userID] = append(users.UserURLs[userID], userItem)

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

		if fullURL == "url is deleted" {
			res.WriteHeader(http.StatusGone)
			return
		}

		res.Header().Set("Location", fullURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte(fullURL))
	}
}

// Хендлер получения всех сокращений пользователя
func UserURLS(users *userauth.Users) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var tokenString string

		cookie, err := req.Cookie("jwt")
		if err != nil {
			log.Println(err)
		} else {
			tokenString = cookie.Value
		}

		userID := userauth.GetUserID(tokenString)
		if userID < 0 {
			userID = req.Context().Value("userID").(int)

			if userID <= 0 {
				res.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		if len(users.UserURLs[userID]) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		// Отправляем ответ
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(users.UserURLs[userID])
	}
}

// Хендлер удаления сокращений пользователя
func DeleteShorts(users *userauth.Users, storage *model.Storage, cfg *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var tokenString string

		cookie, err := req.Cookie("jwt")
		if err != nil {
			log.Println(err)
		} else {
			tokenString = cookie.Value
		}

		userID := userauth.GetUserID(tokenString)
		if userID < 0 {
			userID = req.Context().Value("userID").(int)

			if userID <= 0 {
				res.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		var buf bytes.Buffer

		// читаем тело запроса
		_, err = buf.ReadFrom(req.Body)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var shortsForDelete []string
		err = json.Unmarshal(buf.Bytes(), &shortsForDelete)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		s := service.NewShortenerService(storage, cfg)

		var dataForDelete []string

		for _, shortForDelete := range shortsForDelete {
			usersURLs := users.UserURLs[userID]

			shortURL, err := s.GetShortURL(shortForDelete)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			var newUserURLs []userauth.UserItem
			for _, item := range usersURLs {
				if shortURL == item.ShortURL {
					dataForDelete = append(dataForDelete, shortForDelete)
				} else {
					newUserURLs = append(newUserURLs, item)
				}
			}

			users.UserURLs[userID] = newUserURLs
		}

		err = s.DeleteBulkData(dataForDelete)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		res.WriteHeader(http.StatusAccepted)
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
