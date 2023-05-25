package main

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/go-chi/chi"
	"net/http"
	"os"
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
	r.NotFound(
		func(w http.ResponseWriter, r *http.Request) {
			handlers.RedirectURL(w, r, store)
		},
	)

	// Запуск сервера
	err := http.ListenAndServe(cfg.Addr, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}

}
