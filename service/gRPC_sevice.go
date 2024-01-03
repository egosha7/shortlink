// grpc_service.go

package service

import (
	pb "github.com/egosha7/shortlink/cmd/gRPC/proto"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
)

// GRPCService реализует интерфейс вашей gRPC службы.
type GRPCService struct {
	pb.UnimplementedShortLinkServiceServer
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

// mustEmbedUnimplementedShortLinkServiceServer реализует метод из интерфейса.
func (s *GRPCService) mustEmbedUnimplementedShortLinkServiceServer() {}
