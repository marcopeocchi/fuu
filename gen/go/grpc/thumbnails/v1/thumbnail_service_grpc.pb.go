// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: grpc/thumbnails/v1/thumbnail_service.proto

package thumbnailspb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ThumbnailServiceClient is the client API for ThumbnailService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ThumbnailServiceClient interface {
	Generate(ctx context.Context, in *GenerateRequest, opts ...grpc.CallOption) (*GenerateResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	GetRange(ctx context.Context, in *GetRangeRequest, opts ...grpc.CallOption) (*GetRangeResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
}

type thumbnailServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewThumbnailServiceClient(cc grpc.ClientConnInterface) ThumbnailServiceClient {
	return &thumbnailServiceClient{cc}
}

func (c *thumbnailServiceClient) Generate(ctx context.Context, in *GenerateRequest, opts ...grpc.CallOption) (*GenerateResponse, error) {
	out := new(GenerateResponse)
	err := c.cc.Invoke(ctx, "/thumbnails.v1.ThumbnailService/Generate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thumbnailServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/thumbnails.v1.ThumbnailService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thumbnailServiceClient) GetRange(ctx context.Context, in *GetRangeRequest, opts ...grpc.CallOption) (*GetRangeResponse, error) {
	out := new(GetRangeResponse)
	err := c.cc.Invoke(ctx, "/thumbnails.v1.ThumbnailService/GetRange", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thumbnailServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/thumbnails.v1.ThumbnailService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ThumbnailServiceServer is the server API for ThumbnailService service.
// All implementations should embed UnimplementedThumbnailServiceServer
// for forward compatibility
type ThumbnailServiceServer interface {
	Generate(context.Context, *GenerateRequest) (*GenerateResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	GetRange(context.Context, *GetRangeRequest) (*GetRangeResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
}

// UnimplementedThumbnailServiceServer should be embedded to have forward compatible implementations.
type UnimplementedThumbnailServiceServer struct {
}

func (UnimplementedThumbnailServiceServer) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Generate not implemented")
}
func (UnimplementedThumbnailServiceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedThumbnailServiceServer) GetRange(context.Context, *GetRangeRequest) (*GetRangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRange not implemented")
}
func (UnimplementedThumbnailServiceServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

// UnsafeThumbnailServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ThumbnailServiceServer will
// result in compilation errors.
type UnsafeThumbnailServiceServer interface {
	mustEmbedUnimplementedThumbnailServiceServer()
}

func RegisterThumbnailServiceServer(s grpc.ServiceRegistrar, srv ThumbnailServiceServer) {
	s.RegisterService(&ThumbnailService_ServiceDesc, srv)
}

func _ThumbnailService_Generate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThumbnailServiceServer).Generate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/thumbnails.v1.ThumbnailService/Generate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThumbnailServiceServer).Generate(ctx, req.(*GenerateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThumbnailService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThumbnailServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/thumbnails.v1.ThumbnailService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThumbnailServiceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThumbnailService_GetRange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRangeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThumbnailServiceServer).GetRange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/thumbnails.v1.ThumbnailService/GetRange",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThumbnailServiceServer).GetRange(ctx, req.(*GetRangeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThumbnailService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThumbnailServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/thumbnails.v1.ThumbnailService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThumbnailServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ThumbnailService_ServiceDesc is the grpc.ServiceDesc for ThumbnailService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ThumbnailService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "thumbnails.v1.ThumbnailService",
	HandlerType: (*ThumbnailServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Generate",
			Handler:    _ThumbnailService_Generate_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ThumbnailService_Delete_Handler,
		},
		{
			MethodName: "GetRange",
			Handler:    _ThumbnailService_GetRange_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _ThumbnailService_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/thumbnails/v1/thumbnail_service.proto",
}