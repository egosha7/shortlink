package config

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"net"
	"os"
	"regexp"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr    string // Адрес сервера
	BaseURL string // Базовый адрес результирующего сокращённого URL
}

// Default - функци для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:    "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
}

func OnFlag() *Config {
	defaultValue := Default()
	// Инициализация флагов командной строки
	addr := flag.String("a", defaultValue.Addr, "HTTP-адрес сервера")
	baseURL := flag.String("b", defaultValue.BaseURL, "Базовый адрес результирующего сокращённого URL")
	flag.Parse()

	godotenv.Load()

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
		fmt.Fprintf(os.Stderr, "Invalid base URL \n")
		os.Exit(1)
	}

	return &Config{
		Addr:    *addr,
		BaseURL: *baseURL,
	}
}
