package main

import (
	"flag"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/go-chi/chi"
	"net/http"
)

func main() {
	// Инициализация конфигурации
	cfg := config.New()

	// Инициализация флагов командной строки
	addr := flag.String("a", "localhost:8080", "HTTP-адрес сервера")
	baseURL := flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
	flag.Parse()

	// Настройка конфигурации на основе флагов
	cfg.Addr = *addr
	cfg.BaseURL = *baseURL

	// Создание роутера
	r := chi.NewRouter()
	r.Get(`/`, handlers.RedirectURL)
	r.Post(`/`, handlers.ShortenURL)
	r.NotFound(handlers.RedirectURL)

	// Запуск сервера
	http.ListenAndServe(cfg.Addr, r)
}
