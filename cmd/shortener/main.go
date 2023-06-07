package main

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/loger"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	cfg := config.OnFlag() // Проверка конфигурации флагов и переменных окружения
	runServer(cfg)
}

func runServer(cfg *config.Config) {
	// Создание роутера
	store := storage.NewURLStore(cfg.FilePath)

	// Загрузка данных из файла
	err := store.LoadFromFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data from file: %v\n", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(handlers.GzipMiddleware)
	r.HandleFunc(
		"/{id}", func(w http.ResponseWriter, r *http.Request) {
			handlers.RedirectURL(w, r, store)
		},
	)
	r.HandleFunc(
		`/`, func(w http.ResponseWriter, r *http.Request) {
			handlers.ShortenURL(w, r, cfg, store)
		},
	)
	r.HandleFunc(
		`/api/shorten`, func(w http.ResponseWriter, r *http.Request) {
			handlers.HandleShortenURL(w, r, cfg, store)
		},
	)

	r.NotFound(
		func(w http.ResponseWriter, r *http.Request) {
			handlers.RedirectURL(w, r, store)
		},
	)

	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Запуск сервера
	err = http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
	if err != nil {
		logger.Error("Error starting server", zap.Error(err))
		os.Exit(1)
	}
}
