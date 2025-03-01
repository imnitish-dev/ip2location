// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/ip2location.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.


const (
	IP2LocationService_LookupIP_FullMethodName = "/ip2location.IP2LocationService/LookupIP"
)

// IP2LocationServiceClient is the client API for IP2LocationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IP2LocationServiceClient interface {
	LookupIP(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error)
}

type iP2LocationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIP2LocationServiceClient(cc grpc.ClientConnInterface) IP2LocationServiceClient {
	return &iP2LocationServiceClient{cc}
}

func (c *iP2LocationServiceClient) LookupIP(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LookupResponse)
	err := c.cc.Invoke(ctx, IP2LocationService_LookupIP_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IP2LocationServiceServer is the server API for IP2LocationService service.
// All implementations must embed UnimplementedIP2LocationServiceServer
// for forward compatibility.
type IP2LocationServiceServer interface {
	LookupIP(context.Context, *LookupRequest) (*LookupResponse, error)
	mustEmbedUnimplementedIP2LocationServiceServer()
}

// UnimplementedIP2LocationServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedIP2LocationServiceServer struct{}

func (UnimplementedIP2LocationServiceServer) LookupIP(context.Context, *LookupRequest) (*LookupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LookupIP not implemented")
}
func (UnimplementedIP2LocationServiceServer) mustEmbedUnimplementedIP2LocationServiceServer() {}
func (UnimplementedIP2LocationServiceServer) testEmbeddedByValue()                            {}

// UnsafeIP2LocationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IP2LocationServiceServer will
// result in compilation errors.
type UnsafeIP2LocationServiceServer interface {
	mustEmbedUnimplementedIP2LocationServiceServer()
}

func RegisterIP2LocationServiceServer(s grpc.ServiceRegistrar, srv IP2LocationServiceServer) {
	// If the following call pancis, it indicates UnimplementedIP2LocationServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&IP2LocationService_ServiceDesc, srv)
}

func _IP2LocationService_LookupIP_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IP2LocationServiceServer).LookupIP(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IP2LocationService_LookupIP_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IP2LocationServiceServer).LookupIP(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IP2LocationService_ServiceDesc is the grpc.ServiceDesc for IP2LocationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IP2LocationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ip2location.IP2LocationService",
	HandlerType: (*IP2LocationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LookupIP",
			Handler:    _IP2LocationService_LookupIP_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/ip2location.proto",
}
