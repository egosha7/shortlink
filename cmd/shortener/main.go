package main

import (
	"flag"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/go-chi/chi"
	"net"
	"net/http"
	"os"
	"regexp"
)

func main() {
	// Инициализация конфигурации
	cfg := config.New()

	// Инициализация флагов командной строки
	addr := flag.String("a", "localhost:8080", "HTTP-адрес сервера")
	baseURL := flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
	flag.Parse()

	// Проверка корректности введенных значений флагов
	if _, _, err := net.SplitHostPort(*addr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid address: %v\n", err)
		os.Exit(1)
	}
	if matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, *baseURL); !matched {
		fmt.Fprintf(os.Stderr, "Invalid base URL\n")
		os.Exit(1)
	}

	// Настройка конфигурации на основе флагов
	cfg.Addr = *addr
	cfg.BaseURL = *baseURL

	// Создание роутера
	r := chi.NewRouter()
	r.Get(`/`, handlers.RedirectURL)
	r.Post(`/`, handlers.ShortenURL)
	r.NotFound(handlers.RedirectURL)

	// Запуск сервера
	err := http.ListenAndServe(cfg.Addr, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
