package config

// Config - структура конфигурации приложения
type Config struct {
	Addr    string // Адрес сервера
	BaseURL string // Базовый адрес результирующего сокращённого URL
}

// Default - функци для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:    "123",
		BaseURL: "123",
	}
}

func OnFlag(addr string, baseURL string) *Config {
	return &Config{
		Addr:    addr,
		BaseURL: baseURL,
	}
}
