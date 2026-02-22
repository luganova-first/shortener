package main

import (
	"github.com/luganova-first/shortener/internal/archiver"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/handler"
	"github.com/luganova-first/shortener/internal/logger"
	"github.com/luganova-first/shortener/internal/model"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Инициализация конфигурации
	cfg := config.NewConfig()

	// Валидация конфигурации
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	storage := model.NewStorage()

	r := chi.NewRouter()

	// Передаем базовый URL в хендлер
	r.Post("/", handler.SetShort(storage, cfg.BaseURL))
	r.Post("/api/shorten", handler.JSONShort(storage, cfg.BaseURL))
	r.Get("/{shortID}", handler.GetShort(storage))

	r.MethodNotAllowed(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
	})

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, archiver.GzipHandler(logger.WithLogging(r))))
}
