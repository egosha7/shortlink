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
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
)

// Глобальные переменные
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

	// Включение HTTPS, если задан соответствующий флаг или переменная окружени
	enableHTTPS := cfg.SwitchSSL

	if enableHTTPS {
		// autocert.Manager обеспечивает автоматическую генерацию и обновление сертификатов Let's Encrypt
		manager := &autocert.Manager{
			// Директория для хранения сертификатов
			Cache: autocert.DirCache("/cert"),
			// Функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// Ваш временный домен от ngrok
			HostPolicy: autocert.HostWhitelist("816d-178-214-245-167.ngrok-free.app"),
		}

		// Запуск веб-сервера с HTTPS и использованием autocert.Listener
		server := &http.Server{
			Addr:      ":8443", // Используйте порт HTTPS
			Handler:   loger.LogMiddleware(logger, r),
			TLSConfig: manager.TLSConfig(),
		}

		// Горутина для автоматического обновления сертификатов
		go func() {
			if err := http.Serve(manager.Listener(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, TLS"))
			})); err != nil {
				logger.Fatal("Error serving autocert", zap.Error(err))
			}
		}()

		// Запуск основного сервера с HTTPS
		err := server.ListenAndServeTLS("", "")
		if err != nil {
			logger.Error("Error starting HTTPS server", zap.Error(err))
			os.Exit(1)
		}
	} else {
		// Запуск веб-сервера с HTTP
		err := http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
		if err != nil {
			logger.Error("Error starting HTTP server", zap.Error(err))
			os.Exit(1)
		}
	}
}
