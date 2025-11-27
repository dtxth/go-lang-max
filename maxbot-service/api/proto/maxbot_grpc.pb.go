// Code generated manually to complement protoc --go_out output. DO NOT EDIT.
package maxbotproto

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

type MaxBotServiceClient interface {
	GetMaxIDByPhone(ctx context.Context, in *GetMaxIDByPhoneRequest, opts ...grpc.CallOption) (*GetMaxIDByPhoneResponse, error)
	ValidatePhone(ctx context.Context, in *ValidatePhoneRequest, opts ...grpc.CallOption) (*ValidatePhoneResponse, error)
}

type maxBotServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMaxBotServiceClient(cc grpc.ClientConnInterface) MaxBotServiceClient {
	return &maxBotServiceClient{cc}
}

func (c *maxBotServiceClient) GetMaxIDByPhone(ctx context.Context, in *GetMaxIDByPhoneRequest, opts ...grpc.CallOption) (*GetMaxIDByPhoneResponse, error) {
	out := new(GetMaxIDByPhoneResponse)
	err := c.cc.Invoke(ctx, "/maxbot.MaxBotService/GetMaxIDByPhone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maxBotServiceClient) ValidatePhone(ctx context.Context, in *ValidatePhoneRequest, opts ...grpc.CallOption) (*ValidatePhoneResponse, error) {
	out := new(ValidatePhoneResponse)
	err := c.cc.Invoke(ctx, "/maxbot.MaxBotService/ValidatePhone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type MaxBotServiceServer interface {
	GetMaxIDByPhone(context.Context, *GetMaxIDByPhoneRequest) (*GetMaxIDByPhoneResponse, error)
	ValidatePhone(context.Context, *ValidatePhoneRequest) (*ValidatePhoneResponse, error)
	mustEmbedUnimplementedMaxBotServiceServer()
}

type UnimplementedMaxBotServiceServer struct{}

func (UnimplementedMaxBotServiceServer) GetMaxIDByPhone(context.Context, *GetMaxIDByPhoneRequest) (*GetMaxIDByPhoneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMaxIDByPhone not implemented")
}

func (UnimplementedMaxBotServiceServer) ValidatePhone(context.Context, *ValidatePhoneRequest) (*ValidatePhoneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatePhone not implemented")
}

func (UnimplementedMaxBotServiceServer) mustEmbedUnimplementedMaxBotServiceServer() {}

type UnsafeMaxBotServiceServer interface {
	mustEmbedUnimplementedMaxBotServiceServer()
}

func RegisterMaxBotServiceServer(s grpc.ServiceRegistrar, srv MaxBotServiceServer) {
	s.RegisterService(&MaxBotService_ServiceDesc, srv)
}

func _MaxBotService_GetMaxIDByPhone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMaxIDByPhoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaxBotServiceServer).GetMaxIDByPhone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/maxbot.MaxBotService/GetMaxIDByPhone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaxBotServiceServer).GetMaxIDByPhone(ctx, req.(*GetMaxIDByPhoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MaxBotService_ValidatePhone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidatePhoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaxBotServiceServer).ValidatePhone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/maxbot.MaxBotService/ValidatePhone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaxBotServiceServer).ValidatePhone(ctx, req.(*ValidatePhoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var MaxBotService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "maxbot.MaxBotService",
	HandlerType: (*MaxBotServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetMaxIDByPhone",
			Handler:    _MaxBotService_GetMaxIDByPhone_Handler,
		},
		{
			MethodName: "ValidatePhone",
			Handler:    _MaxBotService_ValidatePhone_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/maxbot.proto",
}
