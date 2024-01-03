// grpc_service.go

package service

import (
	"context"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
	"github.com/egosha7/shortlink/logic"
)

// GRPCService реализует интерфейс вашей gRPC службы.
type GRPCService struct {
	shortlink.UnimplementedShortLinkServiceServer
	store   *storage.URLStore
	worker  *worker.Worker
	baseURL string
}

// NewGRPCService создает экземпляр GRPCService.
func NewGRPCService(store *storage.URLStore, worker *worker.Worker, baseURL string) *GRPCService {
	return &GRPCService{
		store:   store,
		worker:  worker,
		baseURL: baseURL,
	}
}

// ShortenURL реализует метод ShortenURL вашей gRPC службы.
func (s *GRPCService) ShortenURL(ctx context.Context, req *shortlink.ShortenURLRequest) (*shortlink.ShortenURLResponse, error) {
	// Реализация метода ShortenURL, используя вашу логику
	body := []byte(req.Body)
	userID := req.UserId
	result, err := logic.ShortenURL(body, userID, s.store, s.baseURL)
	if err != nil {
		return nil, err
	}
	return &shortlink.ShortenURLResponse{Result: result}, nil
}

// Добавьте реализацию других методов вашей gRPC службы
