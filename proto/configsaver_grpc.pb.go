// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

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

// ConfigSaverClient is the client API for ConfigSaver service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ConfigSaverClient interface {
	// Get a configuration from the server.
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigReply, error)
	// Save an updated configuration to the server.
	// The server is responsible for determining whether the configuration is valid, has changed
	// and how to persist it.
	UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigReply, error)
}

type configSaverClient struct {
	cc grpc.ClientConnInterface
}

func NewConfigSaverClient(cc grpc.ClientConnInterface) ConfigSaverClient {
	return &configSaverClient{cc}
}

func (c *configSaverClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigReply, error) {
	out := new(GetConfigReply)
	err := c.cc.Invoke(ctx, "/configsaver.ConfigSaver/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *configSaverClient) UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigReply, error) {
	out := new(UpdateConfigReply)
	err := c.cc.Invoke(ctx, "/configsaver.ConfigSaver/UpdateConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ConfigSaverServer is the server API for ConfigSaver service.
// All implementations must embed UnimplementedConfigSaverServer
// for forward compatibility
type ConfigSaverServer interface {
	// Get a configuration from the server.
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigReply, error)
	// Save an updated configuration to the server.
	// The server is responsible for determining whether the configuration is valid, has changed
	// and how to persist it.
	UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigReply, error)
	mustEmbedUnimplementedConfigSaverServer()
}

// UnimplementedConfigSaverServer must be embedded to have forward compatible implementations.
type UnimplementedConfigSaverServer struct {
}

func (UnimplementedConfigSaverServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedConfigSaverServer) UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConfig not implemented")
}
func (UnimplementedConfigSaverServer) mustEmbedUnimplementedConfigSaverServer() {}

// UnsafeConfigSaverServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ConfigSaverServer will
// result in compilation errors.
type UnsafeConfigSaverServer interface {
	mustEmbedUnimplementedConfigSaverServer()
}

func RegisterConfigSaverServer(s grpc.ServiceRegistrar, srv ConfigSaverServer) {
	s.RegisterService(&ConfigSaver_ServiceDesc, srv)
}

func _ConfigSaver_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigSaverServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/configsaver.ConfigSaver/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigSaverServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConfigSaver_UpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigSaverServer).UpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/configsaver.ConfigSaver/UpdateConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigSaverServer).UpdateConfig(ctx, req.(*UpdateConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ConfigSaver_ServiceDesc is the grpc.ServiceDesc for ConfigSaver service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ConfigSaver_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "configsaver.ConfigSaver",
	HandlerType: (*ConfigSaverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _ConfigSaver_GetConfig_Handler,
		},
		{
			MethodName: "UpdateConfig",
			Handler:    _ConfigSaver_UpdateConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/configsaver.proto",
}
