// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: proto/grpc_info.proto

package proto

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

const (
	URLService_ShortenURL_FullMethodName     = "/grpc.URLService/ShortenURL"
	URLService_RegisterUser_FullMethodName   = "/grpc.URLService/RegisterUser"
	URLService_Login_FullMethodName          = "/grpc.URLService/Login"
	URLService_GetFullURL_FullMethodName     = "/grpc.URLService/GetFullURL"
	URLService_DeleteUserURLs_FullMethodName = "/grpc.URLService/DeleteUserURLs"
)

// URLServiceClient is the client API for URLService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type URLServiceClient interface {
	ShortenURL(ctx context.Context, in *AddURLRequest, opts ...grpc.CallOption) (*AddURLResponse, error)
	RegisterUser(ctx context.Context, in *RegisterUserRequest, opts ...grpc.CallOption) (*RegisterUserResponse, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	GetFullURL(ctx context.Context, in *GetURLRequest, opts ...grpc.CallOption) (*GetURLResponse, error)
	DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*DeleteUserURLsResponse, error)
}

type uRLServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewURLServiceClient(cc grpc.ClientConnInterface) URLServiceClient {
	return &uRLServiceClient{cc}
}

func (c *uRLServiceClient) ShortenURL(ctx context.Context, in *AddURLRequest, opts ...grpc.CallOption) (*AddURLResponse, error) {
	out := new(AddURLResponse)
	err := c.cc.Invoke(ctx, URLService_ShortenURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLServiceClient) RegisterUser(ctx context.Context, in *RegisterUserRequest, opts ...grpc.CallOption) (*RegisterUserResponse, error) {
	out := new(RegisterUserResponse)
	err := c.cc.Invoke(ctx, URLService_RegisterUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, URLService_Login_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLServiceClient) GetFullURL(ctx context.Context, in *GetURLRequest, opts ...grpc.CallOption) (*GetURLResponse, error) {
	out := new(GetURLResponse)
	err := c.cc.Invoke(ctx, URLService_GetFullURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLServiceClient) DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*DeleteUserURLsResponse, error) {
	out := new(DeleteUserURLsResponse)
	err := c.cc.Invoke(ctx, URLService_DeleteUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// URLServiceServer is the server API for URLService service.
// All implementations must embed UnimplementedURLServiceServer
// for forward compatibility
type URLServiceServer interface {
	ShortenURL(context.Context, *AddURLRequest) (*AddURLResponse, error)
	RegisterUser(context.Context, *RegisterUserRequest) (*RegisterUserResponse, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	GetFullURL(context.Context, *GetURLRequest) (*GetURLResponse, error)
	DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*DeleteUserURLsResponse, error)
	mustEmbedUnimplementedURLServiceServer()
}

// UnimplementedURLServiceServer must be embedded to have forward compatible implementations.
type UnimplementedURLServiceServer struct {
}

func (UnimplementedURLServiceServer) ShortenURL(context.Context, *AddURLRequest) (*AddURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShortenURL not implemented")
}
func (UnimplementedURLServiceServer) RegisterUser(context.Context, *RegisterUserRequest) (*RegisterUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterUser not implemented")
}
func (UnimplementedURLServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedURLServiceServer) GetFullURL(context.Context, *GetURLRequest) (*GetURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFullURL not implemented")
}
func (UnimplementedURLServiceServer) DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*DeleteUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}
func (UnimplementedURLServiceServer) mustEmbedUnimplementedURLServiceServer() {}

// UnsafeURLServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to URLServiceServer will
// result in compilation errors.
type UnsafeURLServiceServer interface {
	mustEmbedUnimplementedURLServiceServer()
}

func RegisterURLServiceServer(s grpc.ServiceRegistrar, srv URLServiceServer) {
	s.RegisterService(&URLService_ServiceDesc, srv)
}

func _URLService_ShortenURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLServiceServer).ShortenURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLService_ShortenURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLServiceServer).ShortenURL(ctx, req.(*AddURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLService_RegisterUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLServiceServer).RegisterUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLService_RegisterUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLServiceServer).RegisterUser(ctx, req.(*RegisterUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLService_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLService_GetFullURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLServiceServer).GetFullURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLService_GetFullURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLServiceServer).GetFullURL(ctx, req.(*GetURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLService_DeleteUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLServiceServer).DeleteUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLService_DeleteUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLServiceServer).DeleteUserURLs(ctx, req.(*DeleteUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// URLService_ServiceDesc is the grpc.ServiceDesc for URLService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var URLService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.URLService",
	HandlerType: (*URLServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ShortenURL",
			Handler:    _URLService_ShortenURL_Handler,
		},
		{
			MethodName: "RegisterUser",
			Handler:    _URLService_RegisterUser_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _URLService_Login_Handler,
		},
		{
			MethodName: "GetFullURL",
			Handler:    _URLService_GetFullURL_Handler,
		},
		{
			MethodName: "DeleteUserURLs",
			Handler:    _URLService_DeleteUserURLs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/grpc_info.proto",
}
