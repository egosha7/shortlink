// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: shortlink.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ShortLinkService_ShortenURL_FullMethodName         = "/shortlink.ShortLinkService/ShortenURL"
	ShortLinkService_DeleteUserURLs_FullMethodName     = "/shortlink.ShortLinkService/DeleteUserURLs"
	ShortLinkService_GetUserURLs_FullMethodName        = "/shortlink.ShortLinkService/GetUserURLs"
	ShortLinkService_HandleShortenURL_FullMethodName   = "/shortlink.ShortLinkService/HandleShortenURL"
	ShortLinkService_HandleShortenBatch_FullMethodName = "/shortlink.ShortLinkService/HandleShortenBatch"
	ShortLinkService_Stats_FullMethodName              = "/shortlink.ShortLinkService/Stats"
)

// ShortLinkServiceClient is the client API for ShortLinkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortLinkServiceClient interface {
	ShortenURL(ctx context.Context, in *ShortenURLRequest, opts ...grpc.CallOption) (*ShortenURLResponse, error)
	DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetUserURLs(ctx context.Context, in *GetUserURLsRequest, opts ...grpc.CallOption) (*GetUserURLsResponse, error)
	HandleShortenURL(ctx context.Context, in *HandleShortenURLRequest, opts ...grpc.CallOption) (*HandleShortenURLResponse, error)
	HandleShortenBatch(ctx context.Context, in *HandleShortenBatchRequest, opts ...grpc.CallOption) (*HandleShortenBatchResponse, error)
	Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error)
}

type shortLinkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShortLinkServiceClient(cc grpc.ClientConnInterface) ShortLinkServiceClient {
	return &shortLinkServiceClient{cc}
}

func (c *shortLinkServiceClient) ShortenURL(ctx context.Context, in *ShortenURLRequest, opts ...grpc.CallOption) (*ShortenURLResponse, error) {
	out := new(ShortenURLResponse)
	err := c.cc.Invoke(ctx, ShortLinkService_ShortenURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortLinkServiceClient) DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, ShortLinkService_DeleteUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortLinkServiceClient) GetUserURLs(ctx context.Context, in *GetUserURLsRequest, opts ...grpc.CallOption) (*GetUserURLsResponse, error) {
	out := new(GetUserURLsResponse)
	err := c.cc.Invoke(ctx, ShortLinkService_GetUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortLinkServiceClient) HandleShortenURL(ctx context.Context, in *HandleShortenURLRequest, opts ...grpc.CallOption) (*HandleShortenURLResponse, error) {
	out := new(HandleShortenURLResponse)
	err := c.cc.Invoke(ctx, ShortLinkService_HandleShortenURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortLinkServiceClient) HandleShortenBatch(ctx context.Context, in *HandleShortenBatchRequest, opts ...grpc.CallOption) (*HandleShortenBatchResponse, error) {
	out := new(HandleShortenBatchResponse)
	err := c.cc.Invoke(ctx, ShortLinkService_HandleShortenBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortLinkServiceClient) Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error) {
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, ShortLinkService_Stats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortLinkServiceServer is the server API for ShortLinkService service.
// All implementations must embed UnimplementedShortLinkServiceServer
// for forward compatibility
type ShortLinkServiceServer interface {
	ShortenURL(context.Context, *ShortenURLRequest) (*ShortenURLResponse, error)
	DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*emptypb.Empty, error)
	GetUserURLs(context.Context, *GetUserURLsRequest) (*GetUserURLsResponse, error)
	HandleShortenURL(context.Context, *HandleShortenURLRequest) (*HandleShortenURLResponse, error)
	HandleShortenBatch(context.Context, *HandleShortenBatchRequest) (*HandleShortenBatchResponse, error)
	Stats(context.Context, *StatsRequest) (*StatsResponse, error)
	mustEmbedUnimplementedShortLinkServiceServer()
}

// UnimplementedShortLinkServiceServer must be embedded to have forward compatible implementations.
type UnimplementedShortLinkServiceServer struct {
}

func (UnimplementedShortLinkServiceServer) ShortenURL(context.Context, *ShortenURLRequest) (*ShortenURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShortenURL not implemented")
}
func (UnimplementedShortLinkServiceServer) DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}
func (UnimplementedShortLinkServiceServer) GetUserURLs(context.Context, *GetUserURLsRequest) (*GetUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserURLs not implemented")
}
func (UnimplementedShortLinkServiceServer) HandleShortenURL(context.Context, *HandleShortenURLRequest) (*HandleShortenURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleShortenURL not implemented")
}
func (UnimplementedShortLinkServiceServer) HandleShortenBatch(context.Context, *HandleShortenBatchRequest) (*HandleShortenBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleShortenBatch not implemented")
}
func (UnimplementedShortLinkServiceServer) Stats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stats not implemented")
}
func (UnimplementedShortLinkServiceServer) mustEmbedUnimplementedShortLinkServiceServer() {}

// UnsafeShortLinkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortLinkServiceServer will
// result in compilation errors.
type UnsafeShortLinkServiceServer interface {
	mustEmbedUnimplementedShortLinkServiceServer()
}

func RegisterShortLinkServiceServer(s grpc.ServiceRegistrar, srv ShortLinkServiceServer) {
	s.RegisterService(&ShortLinkService_ServiceDesc, srv)
}

func _ShortLinkService_ShortenURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShortenURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortLinkServiceServer).ShortenURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortLinkService_ShortenURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortLinkServiceServer).ShortenURL(ctx, req.(*ShortenURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortLinkService_DeleteUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortLinkServiceServer).DeleteUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortLinkService_DeleteUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortLinkServiceServer).DeleteUserURLs(ctx, req.(*DeleteUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortLinkService_GetUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortLinkServiceServer).GetUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortLinkService_GetUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortLinkServiceServer).GetUserURLs(ctx, req.(*GetUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortLinkService_HandleShortenURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandleShortenURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortLinkServiceServer).HandleShortenURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortLinkService_HandleShortenURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortLinkServiceServer).HandleShortenURL(ctx, req.(*HandleShortenURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortLinkService_HandleShortenBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandleShortenBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortLinkServiceServer).HandleShortenBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortLinkService_HandleShortenBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortLinkServiceServer).HandleShortenBatch(ctx, req.(*HandleShortenBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortLinkService_Stats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortLinkServiceServer).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortLinkService_Stats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortLinkServiceServer).Stats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ShortLinkService_ServiceDesc is the grpc.ServiceDesc for ShortLinkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ShortLinkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shortlink.ShortLinkService",
	HandlerType: (*ShortLinkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ShortenURL",
			Handler:    _ShortLinkService_ShortenURL_Handler,
		},
		{
			MethodName: "DeleteUserURLs",
			Handler:    _ShortLinkService_DeleteUserURLs_Handler,
		},
		{
			MethodName: "GetUserURLs",
			Handler:    _ShortLinkService_GetUserURLs_Handler,
		},
		{
			MethodName: "HandleShortenURL",
			Handler:    _ShortLinkService_HandleShortenURL_Handler,
		},
		{
			MethodName: "HandleShortenBatch",
			Handler:    _ShortLinkService_HandleShortenBatch_Handler,
		},
		{
			MethodName: "Stats",
			Handler:    _ShortLinkService_Stats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shortlink.proto",
}