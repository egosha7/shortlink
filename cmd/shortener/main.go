package main

import (
	"context"
	"fmt"
	"github.com/egosha7/shortlink/cmd/gRPC/proto/pb"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/loger"
	"github.com/egosha7/shortlink/internal/router"
	"github.com/egosha7/shortlink/service"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	fmt.Printf("Версия сборки: %s\n", Version)
	fmt.Printf("Дата сборки: %s\n", BuildTime)
	fmt.Printf("Коммит: %s\n", Commit)

	// Настройка логгера.
	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Проверка конфигурации из флагов и переменных окружения.
	cfg := config.OnFlag(logger)

	// Подключение к базе данных.
	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Ошибка подключения к базе данных", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Настройка маршрутов для приложения.
	r, store, wkr := routes.SetupRoutes(cfg, conn, logger)

	// Регистрация маршрутов профилирования.
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Запуск горутины для сервера pprof.
	go func() {
		if err := http.ListenAndServe(":6060", mux); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	// Включение HTTPS, если установлен соответствующий флаг или переменная окружения.
	enableHTTPS := cfg.EnableHTTPS

	// Настройка обработки сигналов для грациозного завершения.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var wg sync.WaitGroup

	// Запуск горутины для обработки сигналов.
	go func() {
		sig := <-signalCh
		fmt.Printf("Получен сигнал %v. Завершение работы...\n", sig)

		// Дождемся завершения оставшихся запросов.
		wg.Wait()

		// Завершаем программу.
		os.Exit(0)
	}()

	// Запуск HTTP или HTTPS сервера в зависимости от конфигурации.
	if enableHTTPS {
		// Настройка менеджера autocert для автоматической генерации сертификатов Let's Encrypt.
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("/cert"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("816d-178-214-245-167.ngrok-free.app"),
		}

		// Создание HTTPS сервера с использованием autocert.Listener.
		server := &http.Server{
			Addr:      ":8443",
			Handler:   loger.LogMiddleware(logger, r),
			TLSConfig: manager.TLSConfig(),
		}

		// Запуск горутины для автоматического обновления сертификатов.
		go func() {
			if err := http.Serve(manager.Listener(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, TLS"))
			})); err != nil {
				logger.Fatal("Ошибка при обслуживании autocert", zap.Error(err))
			}
		}()

		// Запуск HTTPS сервера.
		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil {
				logger.Error("Ошибка запуска HTTPS сервера", zap.Error(err))
				os.Exit(1)
			}
		}()
	} else {
		// Запуск HTTP сервера.
		go func() {
			if err := http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r)); err != nil {
				logger.Error("Ошибка запуска HTTP сервера", zap.Error(err))
				os.Exit(1)
			}
		}()
	}

	// gRPC

	// Запуск gRPC сервера
	grpcService := service.NewGRPCService(store, wkr, cfg)
	grpcServer := grpc.NewServer()
	pb.RegisterShortLinkServiceServer(grpcServer, grpcService)

	// Запуск горутины для gRPC сервера
	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			logger.Fatal("Ошибка при прослушивании порта для gRPC", zap.Error(err))
		}
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Ошибка запуска gRPC сервера", zap.Error(err))
		}
	}()
}
