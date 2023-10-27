package main

import (
	"context"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/loger"
	routes "github.com/egosha7/shortlink/internal/router"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func BenchmarkShortenURL(b *testing.B) {
	// Создаем фейковый HTTP-запрос и респонс
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"url": "http://example.com"}`))
	w := httptest.NewRecorder()

	// Инициализируем необходимые зависимости, включая конфиг и хранилище
	logger, _ := loger.SetupLogger()
	cfg := config.OnFlag(logger)

	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Error connecting to database", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := routes.SetupRoutes(cfg, conn, logger)

	// Запускаем бенчмарк
	b.ResetTimer()
	for i := 0; i < 1000; i++ {
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			b.Fatalf("Unexpected status code: %d", w.Code)
		}
	}
}
