package main

import (
	"context"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/loger"
	routes "github.com/egosha7/shortlink/internal/router"
	"github.com/egosha7/shortlink/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {

	// Проверка конфигурации флагов и переменных окружения
	cfg := config.OnFlag()

	// Создание хранилища
	store := storage.NewURLStore(cfg.FilePath)

	// Загрузка данных из файла
	err := store.LoadFromFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data from file: %v\n", err)
		os.Exit(1)
	}

	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Error connecting to database", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := routes.SetupRoutes(cfg, store, conn)

	// Запуск сервера
	err = http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
	if err != nil {
		logger.Error("Error starting server", zap.Error(err))
		os.Exit(1)
	}
}
