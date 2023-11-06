package main

import (
	"context"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/loger"
	routes "github.com/egosha7/shortlink/internal/router"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
)

func main() {

	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating loger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Проверка конфигурации флагов и переменных окружения
	cfg := config.OnFlag(logger)

	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Error connecting to database", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := routes.SetupRoutes(cfg, conn, logger)

	// Зарегистрируем маршруты для профилирования
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		if err := http.ListenAndServe(":6060", mux); err != nil {
			log.Fatal(err)
		}
	}()

	// Запуск сервера
	err = http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
	if err != nil {
		logger.Error("Error starting server", zap.Error(err))
		os.Exit(1)
	}
}
