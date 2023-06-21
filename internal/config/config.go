package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net"
	"os"
	"regexp"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr     string `env:"SERVER_ADDRESS"`    // Адрес сервера
	BaseURL  string `env:"BASE_URL"`          // Базовый адрес результирующего сокращенного URL
	FilePath string `env:"FILE_STORAGE_PATH"` // Путь к файлу для сохранения данных
	DataBase string `env:"DATABASE_DSN"`      // Адрес базы данных
}

// Default - функция для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:     "localhost:8080",
		BaseURL:  "http://localhost:8080",
		FilePath: "",
		DataBase: "", // postgres://postgres:egosha@localhost:5432/shortlink
	}
}

// OnFlag - функция для чтения значений из флагов командной строки и записи их в структуру Config
func OnFlag(logger *zap.Logger) *Config {
	defaultValue := Default()

	// Инициализация флагов командной строки
	config := Config{}
	flag.StringVar(&config.Addr, "a", defaultValue.Addr, "HTTP-адрес сервера")
	flag.StringVar(&config.BaseURL, "b", defaultValue.BaseURL, "Базовый адрес результирующего сокращенного URL")
	flag.StringVar(&config.FilePath, "f", defaultValue.FilePath, "Путь к файлу данных")
	flag.StringVar(&config.DataBase, "d", defaultValue.DataBase, "Адрес базы данных")
	flag.Parse()

	godotenv.Load()

	// Парсинг переменных окружения в структуру Config
	if err := env.Parse(&config); err != nil {
		logger.Error("Ошибка при парсинге переменных окружения", zap.Error(err))
	}

	// Проверка существования файла
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		// Файл не существует
		logger.Error("Файл не найден", zap.Error(err))
	} else {
		// Файл существует

		// Проверка прав доступа к файлу
		if err := checkFileAccess(config.FilePath); err != nil {
			// Ошибка доступа к файлу
			logger.Error("Ошибка доступа к файлу", zap.Error(err))
		} else {
			// Файл существует и доступен для чтения
			logger.Info("Ошибка доступа к файлу")
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
