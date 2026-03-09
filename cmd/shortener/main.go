package main

import (
	// "fmt"
	"github.com/luganova-first/shortener/internal/archiver"
	"github.com/luganova-first/shortener/internal/config"
	"github.com/luganova-first/shortener/internal/handler"
	"github.com/luganova-first/shortener/internal/logger"
	"github.com/luganova-first/shortener/internal/model"
	"github.com/luganova-first/shortener/internal/repository"
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

	storage, err := repository.FillStorage(model.NewStorage(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// Передаем базовый URL в хендлер
	r.Post("/", handler.SetShort(storage, cfg))
	r.Post("/api/shorten/batch", handler.BatchShort(storage, cfg))
	r.Post("/api/shorten", handler.JSONShort(storage, cfg))
	r.Get("/ping", handler.GetDB(cfg))
	r.Get("/{shortID}", handler.GetShort(storage, cfg))

	r.MethodNotAllowed(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
	})

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, archiver.GzipHandler(logger.WithLogging(r))))
}
