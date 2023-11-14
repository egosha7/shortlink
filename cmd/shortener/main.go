// Пакет main - это точка входа для службы shortlink.
//
// Эта служба предоставляет функциональность сокращения URL.
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

var (
	// Version - это версия сборки приложения.
	Version string
	// BuildTime - это временная метка времени сборки приложения.
	BuildTime string
	// Commit - это хеш коммита приложения.
	Commit string
)

// main - это основная точка входа для службы shortlink.
func main() {
	// Вывести информацию о сборке.
	fmt.Printf("Версия сборки: %s\n", Version)
	fmt.Printf("Дата сборки: %s\n", BuildTime)
	fmt.Printf("Коммит: %s\n", Commit)

	// Настроить логгер.
	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Проверить конфигурацию из флагов и переменных окружения.
	cfg := config.OnFlag(logger)

	// Подключиться к базе данных.
	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Ошибка подключения к базе данных", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Настроить маршруты для приложения.
	r := routes.SetupRoutes(cfg, conn, logger)

	// Зарегистрировать маршруты профилирования.
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Запустить горутину для сервера pprof.
	go func() {
		if err := http.ListenAndServe(":6060", mux); err != nil {
			log.Fatal(err)
		}
	}()

	// Запустить основной сервер.
	err = http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
	if err != nil {
		logger.Error("Ошибка запуска сервера", zap.Error(err))
		os.Exit(1)
	}
}
