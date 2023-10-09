// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: proto/service.proto

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
	Gossip_Handshake_FullMethodName      = "/Gossip/Handshake"
	Gossip_HandleEncDeck_FullMethodName  = "/Gossip/HandleEncDeck"
	Gossip_HandleTakeSeat_FullMethodName = "/Gossip/HandleTakeSeat"
)

// GossipClient is the client API for Gossip service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GossipClient interface {
	Handshake(ctx context.Context, in *Version, opts ...grpc.CallOption) (*Version, error)
	HandleEncDeck(ctx context.Context, in *EncDeck, opts ...grpc.CallOption) (*Ack, error)
	HandleTakeSeat(ctx context.Context, in *TakeSeat, opts ...grpc.CallOption) (*Ack, error)
}

type gossipClient struct {
	cc grpc.ClientConnInterface
}

func NewGossipClient(cc grpc.ClientConnInterface) GossipClient {
	return &gossipClient{cc}
}

func (c *gossipClient) Handshake(ctx context.Context, in *Version, opts ...grpc.CallOption) (*Version, error) {
	out := new(Version)
	err := c.cc.Invoke(ctx, Gossip_Handshake_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gossipClient) HandleEncDeck(ctx context.Context, in *EncDeck, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, Gossip_HandleEncDeck_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gossipClient) HandleTakeSeat(ctx context.Context, in *TakeSeat, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, Gossip_HandleTakeSeat_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GossipServer is the server API for Gossip service.
// All implementations must embed UnimplementedGossipServer
// for forward compatibility
type GossipServer interface {
	Handshake(context.Context, *Version) (*Version, error)
	HandleEncDeck(context.Context, *EncDeck) (*Ack, error)
	HandleTakeSeat(context.Context, *TakeSeat) (*Ack, error)
	mustEmbedUnimplementedGossipServer()
}

// UnimplementedGossipServer must be embedded to have forward compatible implementations.
type UnimplementedGossipServer struct {
}

func (UnimplementedGossipServer) Handshake(context.Context, *Version) (*Version, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Handshake not implemented")
}
func (UnimplementedGossipServer) HandleEncDeck(context.Context, *EncDeck) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleEncDeck not implemented")
}
func (UnimplementedGossipServer) HandleTakeSeat(context.Context, *TakeSeat) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleTakeSeat not implemented")
}
func (UnimplementedGossipServer) mustEmbedUnimplementedGossipServer() {}

// UnsafeGossipServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GossipServer will
// result in compilation errors.
type UnsafeGossipServer interface {
	mustEmbedUnimplementedGossipServer()
}

func RegisterGossipServer(s grpc.ServiceRegistrar, srv GossipServer) {
	s.RegisterService(&Gossip_ServiceDesc, srv)
}

func _Gossip_Handshake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Version)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GossipServer).Handshake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Gossip_Handshake_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GossipServer).Handshake(ctx, req.(*Version))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gossip_HandleEncDeck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EncDeck)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GossipServer).HandleEncDeck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Gossip_HandleEncDeck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GossipServer).HandleEncDeck(ctx, req.(*EncDeck))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gossip_HandleTakeSeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TakeSeat)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GossipServer).HandleTakeSeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Gossip_HandleTakeSeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GossipServer).HandleTakeSeat(ctx, req.(*TakeSeat))
	}
	return interceptor(ctx, in, info, handler)
}

// Gossip_ServiceDesc is the grpc.ServiceDesc for Gossip service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Gossip_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Gossip",
	HandlerType: (*GossipServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Handshake",
			Handler:    _Gossip_Handshake_Handler,
		},
		{
			MethodName: "HandleEncDeck",
			Handler:    _Gossip_HandleEncDeck_Handler,
		},
		{
			MethodName: "HandleTakeSeat",
			Handler:    _Gossip_HandleTakeSeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/service.proto",
}
