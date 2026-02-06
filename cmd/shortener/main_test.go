package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetShort(t *testing.T) {
	shorts := make(Shorted)
	var shortURL string

	t.Run("wrong content-type", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(shorts))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 400, res.StatusCode)
		defer res.Body.Close()
	})

	t.Run("wrong url", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/aaaaaaa", nil)
		request.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(shorts))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 400, res.StatusCode)
		defer res.Body.Close()
	})

	t.Run("no body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(shorts))
		h(w, request)

		res := w.Result()
		assert.Equal(t, 400, res.StatusCode)
		defer res.Body.Close()
	})

	t.Run("ok post", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
		request.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(shorts))
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
		w := httptest.NewRecorder()
		h := http.HandlerFunc(SetShort(shorts))
		h(w, request)

		res := w.Result()
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		assert.Equal(t, string(resBody), shortURL)
	})
}

func TestGetShort(t *testing.T) {
	shorts := make(Shorted)

	testShort := "iPbLQebD"
	testURL := fmt.Sprintf("/%s", testShort)
	shorts[testShort] = "https://practicum.yandex.ru/"

	t.Run("ok get", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, testURL, nil)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/{short_id}", GetShort(shorts))

		r.ServeHTTP(w, request)

		res := w.Result()
		assert.Equal(t, 307, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		assert.Equal(t, "https://practicum.yandex.ru/", string(resBody))
	})
}
