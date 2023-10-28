package benchmark_handlers

import (
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/loger"
	routes "github.com/egosha7/shortlink/internal/router"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkShortenURL(b *testing.B) {
	// Создаем фейковый HTTP-запрос и респонс
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"url": "http://example.com"}`))
	w := httptest.NewRecorder()

	logger, _ := loger.SetupLogger()

	cfg := &config.Config{
		Addr:     "localhost:8080",
		BaseURL:  "http://localhost:8080",
		FilePath: "", // укажите путь к файлу
		DataBase: "", // укажите адрес базы данных
	}

	conn, _ := db.ConnectToDB(cfg)
	r := routes.SetupRoutes(cfg, conn, logger)

	// Запускаем бенчмарк
	b.ResetTimer()
	for i := 0; i < 1000000; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic occurred in BenchmarkShortenURL", zap.Any("panic", r))
				}
			}()

			r.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				b.Fatalf("Unexpected status code: %d", w.Code)
			}
		}()
	}
}
