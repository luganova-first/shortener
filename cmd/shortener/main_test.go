package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainPage(t *testing.T) {
	shorts := make(Shorted)
	var shortUrl string

	t.Run("wrong method", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/", nil)
		// создаём новый Recorder
		w := httptest.NewRecorder()
		h := http.HandlerFunc(MainPage(shorts))
		h(w, request)

		res := w.Result()
		// проверяем код ответа
		assert.Equal(t, 400, res.StatusCode)
	})

	t.Run("wrong content-type", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Content-Type", "application/json")
		// создаём новый Recorder
		w := httptest.NewRecorder()
		h := http.HandlerFunc(MainPage(shorts))
		h(w, request)

		res := w.Result()
		// проверяем код ответа
		assert.Equal(t, 400, res.StatusCode)
	})

	t.Run("wrong url", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/aaaaaaa", nil)
		request.Header.Set("Content-Type", "text/plain")
		// создаём новый Recorder
		w := httptest.NewRecorder()
		h := http.HandlerFunc(MainPage(shorts))
		h(w, request)

		res := w.Result()
		// проверяем код ответа
		assert.Equal(t, 400, res.StatusCode)
	})

	t.Run("no body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Content-Type", "text/plain")
		// создаём новый Recorder
		w := httptest.NewRecorder()
		h := http.HandlerFunc(MainPage(shorts))
		h(w, request)

		res := w.Result()
		// проверяем код ответа
		assert.Equal(t, 400, res.StatusCode)
	})

	t.Run("ok post", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
		request.Header.Set("Content-Type", "text/plain")
		// создаём новый Recorder
		w := httptest.NewRecorder()
		h := http.HandlerFunc(MainPage(shorts))
		h(w, request)

		res := w.Result()
		// проверяем код ответа
		assert.Equal(t, 201, res.StatusCode)
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		body, _ := io.ReadAll(res.Body)
		shortUrl = string(body)
	})

	t.Run("ok get", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, shortUrl, nil)
		// создаём новый Recorder
		w := httptest.NewRecorder()
		h := http.HandlerFunc(MainPage(shorts))
		h(w, request)

		res := w.Result()
		// проверяем код ответа
		assert.Equal(t, 307, res.StatusCode)
		body, _ := io.ReadAll(res.Body)
		assert.Equal(t, "https://practicum.yandex.ru/", string(body))
	})
}
