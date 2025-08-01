// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.32.0--rc1
// source: trip.proto

package trip

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TripService_PreviewTrip_FullMethodName = "/trip.TripService/PreviewTrip"
)

// TripServiceClient is the client API for TripService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TripServiceClient interface {
	PreviewTrip(ctx context.Context, in *PreviewTripRequest, opts ...grpc.CallOption) (*PreviewTripResponse, error)
}

type tripServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTripServiceClient(cc grpc.ClientConnInterface) TripServiceClient {
	return &tripServiceClient{cc}
}

func (c *tripServiceClient) PreviewTrip(ctx context.Context, in *PreviewTripRequest, opts ...grpc.CallOption) (*PreviewTripResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PreviewTripResponse)
	err := c.cc.Invoke(ctx, TripService_PreviewTrip_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TripServiceServer is the server API for TripService service.
// All implementations must embed UnimplementedTripServiceServer
// for forward compatibility.
type TripServiceServer interface {
	PreviewTrip(context.Context, *PreviewTripRequest) (*PreviewTripResponse, error)
	mustEmbedUnimplementedTripServiceServer()
}

// UnimplementedTripServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTripServiceServer struct{}

func (UnimplementedTripServiceServer) PreviewTrip(context.Context, *PreviewTripRequest) (*PreviewTripResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreviewTrip not implemented")
}
func (UnimplementedTripServiceServer) mustEmbedUnimplementedTripServiceServer() {}
func (UnimplementedTripServiceServer) testEmbeddedByValue()                     {}

// UnsafeTripServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TripServiceServer will
// result in compilation errors.
type UnsafeTripServiceServer interface {
	mustEmbedUnimplementedTripServiceServer()
}

func RegisterTripServiceServer(s grpc.ServiceRegistrar, srv TripServiceServer) {
	// If the following call pancis, it indicates UnimplementedTripServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TripService_ServiceDesc, srv)
}

func _TripService_PreviewTrip_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PreviewTripRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TripServiceServer).PreviewTrip(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TripService_PreviewTrip_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TripServiceServer).PreviewTrip(ctx, req.(*PreviewTripRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TripService_ServiceDesc is the grpc.ServiceDesc for TripService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TripService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "trip.TripService",
	HandlerType: (*TripServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PreviewTrip",
			Handler:    _TripService_PreviewTrip_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "trip.proto", 
}
