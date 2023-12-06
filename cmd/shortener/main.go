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
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	Version   string
	BuildTime string
	Commit    string
)

func main() {
	fmt.Printf("Версия сборки: %s\n", Version)
	fmt.Printf("Дата сборки: %s\n", BuildTime)
	fmt.Printf("Коммит: %s\n", Commit)

	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	cfg := config.OnFlag(logger)
	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Ошибка подключения к базе данных", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := routes.SetupRoutes(cfg, conn, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		if err := http.ListenAndServe(":6060", mux); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	enableHTTPS := cfg.EnableHTTPS
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var wg sync.WaitGroup

	go func() {
		sig := <-signalCh
		fmt.Printf("Received signal %v. Shutting down...\n", sig)

		// Дождемся завершения оставшихся запросов
		wg.Wait()

		// Завершаем программу
		os.Exit(0)
	}()

	if enableHTTPS {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("/cert"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("816d-178-214-245-167.ngrok-free.app"),
		}

		server := &http.Server{
			Addr:      ":8443",
			Handler:   loger.LogMiddleware(logger, r),
			TLSConfig: manager.TLSConfig(),
		}

		go func() {
			if err := http.Serve(manager.Listener(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, TLS"))
			})); err != nil {
				logger.Fatal("Error serving autocert", zap.Error(err))
			}
		}()

		err := server.ListenAndServeTLS("", "")
		if err != nil {
			logger.Error("Error starting HTTPS server", zap.Error(err))
			os.Exit(1)
		}
	} else {
		err := http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
		if err != nil {
			logger.Error("Error starting HTTP server", zap.Error(err))
			os.Exit(1)
		}
	}
}
