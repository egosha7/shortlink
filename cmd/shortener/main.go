package main

import (
	"bufio"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/loger"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
)

func main() {
	cfg := config.OnFlag() // Проверка конфигурации флагов и переменных окружения
	store := handlers.NewURLStore()
	LoadLinksFromFile(cfg.FileStorage, store)
	runServer(cfg)
}

func LoadLinksFromFile(filePath string, store *handlers.URLStore) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			store.AddURL(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func runServer(cfg *config.Config) {
	// Создание роутера
	store := handlers.NewURLStore()
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
