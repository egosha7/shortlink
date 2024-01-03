package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net"
	"os"
	"regexp"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr          string `env:"SERVER_ADDRESS" json:"server_address"`       // Адрес сервера
	BaseURL       string `env:"BASE_URL" json:"base_url"`                   // Базовый адрес результирующего сокращенного URL
	FilePath      string `env:"FILE_STORAGE_PATH" json:"file_storage_path"` // Путь к файлу для сохранения данных
	DataBase      string `env:"DATABASE_DSN" json:"database_dsn"`           // Адрес базы данных
	EnableHTTPS   bool   `env:"ENABLE_HTTPS" json:"enable_https"`           // SSL
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`       // Строковое представление бесклассовой адресации (CIDR)
}

// Default - функция для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:          "localhost:8080",
		BaseURL:       "http://localhost:8080",
		FilePath:      "",
		DataBase:      "", // postgres://postgres:egosha@localhost:5432/shortlink
		EnableHTTPS:   false,
		TrustedSubnet: "",
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
	flag.BoolVar(&config.EnableHTTPS, "s", defaultValue.EnableHTTPS, "Переключатель HTTPS")
	flag.StringVar(&config.TrustedSubnet, "t", defaultValue.TrustedSubnet, "CIDR")
	configFile := flag.String("c", "", "Path to the configuration file")
	flag.Parse()

	godotenv.Load()

	// Парсинг переменных окружения в структуру Config
	if err := env.Parse(&config); err != nil {
		logger.Error("Ошибка при парсинге переменных окружения", zap.Error(err))
	}

	// Загрузка конфигурации из файла
	if *configFile != "" {
		fileConfig, err := loadConfig(*configFile)
		if err != nil {
			fmt.Println("Error loading config file:", err)
			os.Exit(1)
		}

		// Использование значений из файла конфигурации
		if fileConfig.Addr != "" {
			config.Addr = fileConfig.Addr
		}
		if fileConfig.BaseURL != "" {
			config.BaseURL = fileConfig.BaseURL
		}
		if fileConfig.FilePath != "" {
			config.FilePath = fileConfig.FilePath
		}
		if fileConfig.DataBase != "" {
			config.DataBase = fileConfig.DataBase
		}
		if fileConfig.TrustedSubnet != "" {
			config.TrustedSubnet = fileConfig.TrustedSubnet
		}
		config.EnableHTTPS = fileConfig.EnableHTTPS
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
