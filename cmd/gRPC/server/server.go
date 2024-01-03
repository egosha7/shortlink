package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
	"github.com/egosha7/shortlink/logic"
	"google.golang.org/grpc"
)

type grpcServer struct {
	shortlink.UnimplementedShortLinkServiceServer
	store   *storage.URLStore
	worker  *worker.Worker
	baseURL string
}

func (s *grpcServer) ShortenURL(ctx context.Context, req *shortlink.ShortenURLRequest) (*shortlink.ShortenURLResponse, error) {
	body := []byte(req.Body)
	userID := req.UserId
	result, err := logic.ShortenURL(body, userID, s.store, s.baseURL)
	if err != nil {
		return nil, err
	}
	return &shortlink.ShortenURLResponse{Result: result}, nil
}

// Добавьте реализацию других методов сервиса

func main() {

	// HTTP сервер
	httpRouter := setupHTTPRouter() // Ваша функция для настройки HTTP маршрутов
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: httpRouter,
	}

	// gRPC сервер
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	shortlink.RegisterShortLinkServiceServer(grpcServer, &grpcServer{store: yourStore, worker: yourWorker, baseURL: "your_base_url"})

	// Запуск HTTP и gRPC серверов в отдельных горутинах
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Ожидание сигнала завершения (например, Ctrl+C)
	select {}
}

// setupHTTPRouter - ваша функция для настройки HTTP маршрутов
func setupHTTPRouter() http.Handler {
	// Реализуйте настройку ваших HTTP маршрутов
	return http.NewServeMux()
}
