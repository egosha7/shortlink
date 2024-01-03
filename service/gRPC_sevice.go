// grpc_service.go

package service

import (
	"context"
	pb "github.com/egosha7/shortlink/cmd/gRPC/proto"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/egosha7/shortlink/internal/worker"
	"github.com/egosha7/shortlink/logic"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GRPCService реализует интерфейс вашей gRPC службы.
type GRPCService struct {
	pb.UnimplementedShortLinkServiceServer
	store  *storage.URLStore
	worker *worker.Worker
	cfg    *config.Config
}

// NewGRPCService создает экземпляр GRPCService.
func NewGRPCService(store *storage.URLStore, worker *worker.Worker, cfg *config.Config) *GRPCService {
	return &GRPCService{
		store:  store,
		worker: worker,
		cfg:    cfg,
	}
}

// mustEmbedUnimplementedShortLinkServiceServer реализует метод из интерфейса.
func (s *GRPCService) mustEmbedUnimplementedShortLinkServiceServer() {}

// ShortenURL реализует gRPC-метод ShortenURL.
func (s *GRPCService) ShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	body := req.GetBody()
	userID := req.GetUserId()

	shortURL, err := logic.ShortenURL([]byte(body), userID, s.store, s.cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return &pb.ShortenURLResponse{Result: shortURL}, nil
}

// DeleteUserURLs реализует gRPC-метод DeleteUserURLs.
func (s *GRPCService) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*emptypb.Empty, error) {
	body := req.GetBody()
	userID := req.GetUserId()

	logic.DeleteUserURLs([]byte(body), userID, s.worker)

	return &emptypb.Empty{}, nil
}

// GetUserURLs реализует gRPC-метод GetUserURLs.
func (s *GRPCService) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	baseURL := req.GetBaseUrl()
	userID := req.GetUserId()

	response, err := logic.GetUserURLs(baseURL, userID, s.store)
	if err != nil {
		return nil, err
	}

	var urlInfos []*pb.URLInfo
	for _, u := range response {
		urlInfos = append(urlInfos, &pb.URLInfo{
			ShortUrl:    u["short_url"],
			OriginalUrl: u["original_url"],
		})
	}

	return &pb.GetUserURLsResponse{UrlInfo: urlInfos}, nil
}

// HandleShortenURL реализует gRPC-метод HandleShortenURL.
func (s *GRPCService) HandleShortenURL(ctx context.Context, req *pb.HandleShortenURLRequest) (*pb.HandleShortenURLResponse, error) {
	body := req.GetBody()
	userID := req.GetUserId()

	shortURL, err := logic.HandleShortenURL([]byte(body), userID, s.store, s.cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return &pb.HandleShortenURLResponse{Result: shortURL}, nil
}

// HandleShortenBatch реализует gRPC-метод HandleShortenBatch.
func (s *GRPCService) HandleShortenBatch(ctx context.Context, req *pb.HandleShortenBatchRequest) (*pb.HandleShortenBatchResponse, error) {
	records := req.GetRecords()
	baseURL := req.GetBaseUrl()
	userID := req.GetUserId()

	var recordInfos []map[string]string
	for _, r := range records {
		recordInfos = append(recordInfos, map[string]string{
			"key":   r.GetKey(),
			"value": r.GetValue(),
		})
	}

	response, err := logic.HandleShortenBatch(recordInfos, ctx, baseURL, s.store, userID)
	if err != nil {
		return nil, err
	}

	var urlInfos []*pb.URLInfo
	for _, u := range response {
		urlInfos = append(urlInfos, &pb.URLInfo{
			ShortUrl:    u["short_url"],
			OriginalUrl: u["original_url"],
		})
	}

	return &pb.HandleShortenBatchResponse{Result: urlInfos}, nil
}
