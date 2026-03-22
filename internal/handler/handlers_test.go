package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/luganova-first/shortener/internal/userauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"context"
)

var cfg *config.Config
var storage *model.Storage
var users *userauth.Users

func TestSetShort(t *testing.T) {
	cfg = config.NewConfig()
	storage = model.NewStorage()
	users = userauth.NewUsers()
	userID := 17

	var shortURL string

	t.Run("wrong content-type", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(storage, cfg, users))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 400, res.StatusCode)
		defer res.Body.Close()
	})

	t.Run("wrong url", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/aaaaaaa", nil)
		request.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(storage, cfg, users))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 400, res.StatusCode)
		defer res.Body.Close()
	})

	t.Run("no body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Content-Type", "text/plain")
		ctx := request.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		request = request.WithContext(ctx)
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(storage, cfg, users))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 400, res.StatusCode)
		defer res.Body.Close()
	})

	t.Run("ok post", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
		request.Header.Set("Content-Type", "text/plain")
		ctx := request.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		request = request.WithContext(ctx)
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(storage, cfg, users))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 201, res.StatusCode)
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		shortURL = string(resBody)
		assert.Contains(t, shortURL, "http://")
	})

	t.Run("no duplicate", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
		request.Header.Set("Content-Type", "text/plain")
		ctx := request.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		request = request.WithContext(ctx)
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(storage, cfg, users))
		h(w, request)

		res := w.Result()
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		assert.Equal(t, string(resBody), shortURL)
	})
}

func TestGetShort(t *testing.T) {
	testShort := "iPbLQebD"
	testURL := fmt.Sprintf("/%s", testShort)
	storage.Shorted[testShort] = "https://practicum.yandex.ru/"

	t.Run("ok get", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, testURL, nil)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/{shortID}", GetShort(storage, cfg))

		r.ServeHTTP(w, request)

		res := w.Result()
		assert.Equal(t, 307, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		assert.Equal(t, "https://practicum.yandex.ru/", string(resBody))
	})
}

func TestJSONShort(t *testing.T) {
	baseURL := cfg.BaseURL

	tests := []struct {
		name              string
		contentType       string
		inputJSON         interface{}
		expectedStatus    int
		expectedResultKey string
		validateResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "successful request",
			contentType:    "application/json",
			inputJSON:      map[string]string{"url": "https://example.com"},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				result, ok := response["result"]
				assert.True(t, ok, "response should have 'result' field")
				assert.Contains(t, result, baseURL+"/", "result should contain base URL")
			},
		},
		{
			name:           "wrong content type",
			contentType:    "text/plain",
			inputJSON:      map[string]string{"url": "https://example.com"},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Empty(t, body, "body should be empty for bad content type")
			},
		},
		{
			name:           "invalid json",
			contentType:    "application/json",
			inputJSON:      "invalid json", // это будет передано как строка, но мы обработаем отдельно
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				assert.NotEmpty(t, body, "body should contain error message")
			},
		},
		{
			name:           "invalid url format",
			contentType:    "application/json",
			inputJSON:      map[string]string{"url": "not-a-valid-url"},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Empty(t, body, "body should be empty for invalid URL")
			},
		},
		{
			name:           "empty url",
			contentType:    "application/json",
			inputJSON:      map[string]string{"url": ""},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Empty(t, body, "body should be empty for empty URL")
			},
		},
		{
			name:           "extra fields in json",
			contentType:    "application/json",
			inputJSON:      map[string]string{"url": "https://example.com", "extra": "field"},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "result")
				assert.NotContains(t, response, "extra")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			// Подготовка тела запроса
			switch v := tt.inputJSON.(type) {
			case string:
				if tt.name == "invalid json" {
					reqBody = []byte("{invalid json}")
				} else {
					reqBody = []byte(v)
				}
			default:
				reqBody, err = json.Marshal(tt.inputJSON)
				require.NoError(t, err)
			}

			// Создаем запрос
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", tt.contentType)

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			handler := JSONShort(storage, cfg)
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			// Проверяем Content-Type для успешных запросов
			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "wrong content type header")
			}

			// Валидируем ответ
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestBatchShort_Success(t *testing.T) {
	handler := BatchShort(storage, cfg)

	// Тестовые данные
	requests := []RequestItem{
		{
			CorrelationID: "123",
			OriginalURL:   "https://example.com/very/long/url/1",
		},
		{
			CorrelationID: "456",
			OriginalURL:   "https://example.com/very/long/url/2",
		},
		{
			CorrelationID: "789",
			OriginalURL:   "https://example.com/very/long/url/3",
		},
	}

	body, err := json.Marshal(requests)
	require.NoError(t, err)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Декодируем ответ
	var responses []ResponseItem
	err = json.NewDecoder(w.Body).Decode(&responses)
	require.NoError(t, err)

	// Проверяем количество ответов
	assert.Len(t, responses, len(requests))

	// Проверяем соответствие correlation_id
	for i, resp := range responses {
		assert.Equal(t, requests[i].CorrelationID, resp.CorrelationID)
		assert.NotEmpty(t, resp.ShortURL)
	}
}
