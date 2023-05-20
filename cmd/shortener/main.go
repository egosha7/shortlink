package main

import (
	"flag"
	"fmt"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/go-chi/chi"
	"net/http"
	"os"
)

type Config struct {
	Address string
	BaseURL string
}

var config Config

func init() {
	flag.StringVar(&config.Address, "a", ":8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.Parse()

	// инициализация полей из переменных окружения
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		config.Address = addr
	}
	if url := os.Getenv("BASE_URL"); url != "" {
		config.BaseURL = url
	}
}

func main() {
	r := chi.NewRouter()
	r.Get(`/`, handlers.RedirectURL)
	r.Post(`/`, handlers.ShortenURL)
	r.NotFound(handlers.RedirectURL)

	err := http.ListenAndServe(config.Address, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
