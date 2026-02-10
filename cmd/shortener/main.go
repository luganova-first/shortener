package main

import (
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/handler"
	"github.com/luganova-first/shortener/internal/model"
	"net/http"
	"log"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Инициализация конфигурации
	cfg := config.NewConfig()

	// Валидация конфигурации
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	shorts := make(model.Shorted)

	r := chi.NewRouter()

	// Передаем базовый URL в хендлер
	r.Post("/", handler.SetShort(shorts, cfg.BaseURL))
	r.Get("/{shortID}", handler.GetShort(shorts))

	r.MethodNotAllowed(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
	})

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
