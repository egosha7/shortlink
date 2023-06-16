package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"net"
	"os"
	"regexp"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr     string `env:"SERVER_ADDRESS"`    // Адрес сервера
	BaseURL  string `env:"BASE_URL"`          // Базовый адрес результирующего сокращенного URL
	FilePath string `env:"FILE_STORAGE_PATH"` // Путь к файлу для сохранения данных
}

// Default - функция для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:     "localhost:8080",
		BaseURL:  "http://localhost:8080",
		FilePath: "",
	}
}

// OnFlag - функция для чтения значений из флагов командной строки и записи их в структуру Config
func OnFlag() *Config {
	defaultValue := Default()

	// Инициализация флагов командной строки
	config := Config{}
	flag.StringVar(&config.Addr, "a", defaultValue.Addr, "HTTP-адрес сервера")
	flag.StringVar(&config.BaseURL, "b", defaultValue.BaseURL, "Базовый адрес результирующего сокращенного URL")
	flag.StringVar(&config.FilePath, "f", defaultValue.FilePath, "Путь к файлу данных")
	flag.Parse()

	godotenv.Load()

	// Парсинг переменных окружения в структуру Config
	if err := env.Parse(&config); err != nil {
		fmt.Println("Ошибка при парсинге переменных окружения:", err)
	}

	// Проверка существования файла
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		// Файл не существует
		fmt.Println("Файл не найден")
	} else {
		// Файл существует

		// Проверка прав доступа к файлу
		if err := checkFileAccess(config.FilePath); err != nil {
			// Ошибка доступа к файлу
			fmt.Println("Ошибка доступа к файлу:", err)
		} else {
			// Файл существует и доступен для чтения
			fmt.Println("Файл существует и доступен для чтения")
		}
	}

	// Проверка корректности введенных значений флагов
	if _, _, err := net.SplitHostPort(config.Addr); err != nil {
		panic(err)
	}
	if matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, config.BaseURL); !matched {
		panic("Invalid base URL")
	}

	return &config
}

// Функция для проверки доступа к файлу
func checkFileAccess(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
