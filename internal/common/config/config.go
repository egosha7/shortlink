package config

// Config - структура конфигурации приложения
type Config struct {
	Addr    string // Адрес сервера
	BaseURL string // Базовый адрес результирующего сокращённого URL
}

// New - функци для создания новой конфигурации с значениями по умолчанию
func New() *Config {
	return &Config{
		Addr:    "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
}
