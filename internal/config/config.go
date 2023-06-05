package config

import (
	"flag"
	"github.com/joho/godotenv"
	"net"
	"os"
	"regexp"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr     string // Адрес сервера
	BaseURL  string // Базовый адрес результирующего сокращенного URL
	FilePath string // Путь к файлу для сохранения данных
}

// Default - функци для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:     "localhost:8080",
		BaseURL:  "http://localhost:8080",
		FilePath: "C:\\Users\\eg898\\shortlink\\cmd\\shortener\\tmp\\some.json",
	}
}

func OnFlag() *Config {
	defaultValue := Default()
	// Инициализация флагов командной строки
	addr := flag.String("a", defaultValue.Addr, "HTTP-адрес сервера")
	baseURL := flag.String("b", defaultValue.BaseURL, "Базовый адрес результирующего сокращённого URL")
	filePath := flag.String("f", defaultValue.FilePath, "Путь к файлу данных")
	flag.Parse()

	godotenv.Load()

	if os.Getenv("SERVER_ADDRESS") != "" {
		*addr = os.Getenv("SERVER_ADDRESS")
	}
	if os.Getenv("BASE_URL") != "" {
		*baseURL = os.Getenv("BASE_URL")
	}
	if os.Getenv("FILE_STORAGE_PATH") != "" {
		*filePath = os.Getenv("FILE_STORAGE_PATH")
	}

	// Проверка корректности введенных значений флагов
	if _, _, err := net.SplitHostPort(*addr); err != nil {
		panic(err)
	}
	if matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, *baseURL); !matched {
		panic("Invalid base URL")
	}

	return &Config{
		Addr:     *addr,
		BaseURL:  *baseURL,
		FilePath: *filePath,
	}
}
