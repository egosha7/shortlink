package main

import (
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
	runServer(cfg)
}

func runServer(cfg *config.Config) {
	// Создание роутера
	store := handlers.NewURLStore()
	r := chi.NewRouter()
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

	// Добавление middleware для сжатия gzip
	r.Use(gzipMiddleware)

	// Запуск сервера
	err = http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
	if err != nil {
		logger.Error("Error starting server", zap.Error(err))
		os.Exit(1)
	}
}

// gzipMiddleware обрабатывает сжатие gzip для запросов и ответов
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовок Accept-Encoding клиента на наличие gzip
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				// Если клиент поддерживает gzip, добавляем заголовок Content-Encoding
				w.Header().Set("Content-Encoding", "gzip")
				// Создаем gzip.Writer поверх текущего ResponseWriter
				gz := handlers.NewGzipResponseWriter(w)
				defer gz.Close()
				// Передаем обработку запроса и сжатый ResponseWriter следующему обработчику
				next.ServeHTTP(gz, r)
			} else {
				// Если клиент не поддерживает gzip, просто передаем управление следующему обработчику
				next.ServeHTTP(w, r)
			}
		},
	)
}
