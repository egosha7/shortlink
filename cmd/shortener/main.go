package main

import (
	"flag"
	"fmt"
	"github.com/egosha7/shortlink/internal/common/config"
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
	addr := flag.String("a", "", "HTTP-адрес сервера")
	baseURL := flag.String("b", "", "Базовый адрес результирующего сокращённого URL")
	flag.Parse()

	if os.Getenv("SERVER_ADDRESS") != "" {
		*addr = os.Getenv("SERVER_ADDRESS")
	}
	if os.Getenv("BASE_URL") != "" {
		*baseURL = os.Getenv("BASE_URL")
	}

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

	runServer(cfg)
}

func runServer(cfg *config.Config) {
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
