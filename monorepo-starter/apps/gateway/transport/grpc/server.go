package grpc

import (
	"context"

	infragrpc "github.com/liukunxin/go-infra/pkg/infra/grpc"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	serviceName = "gateway.RuntimeService"
	methodPing  = "/gateway.RuntimeService/Ping"
)

type pingService struct{}

type PingServer interface {
	Ping(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
}

func (s *pingService) Ping(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String("pong from gateway grpc"), nil
}

func registerPingService(server ggrpc.ServiceRegistrar) {
	server.RegisterService(&ggrpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*PingServer)(nil),
		Methods: []ggrpc.MethodDesc{
			{
				MethodName: "Ping",
				Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor ggrpc.UnaryServerInterceptor) (any, error) {
					req := &emptypb.Empty{}
					if err := dec(req); err != nil {
						return nil, err
					}
					impl := srv.(PingServer)
					if interceptor == nil {
						return impl.Ping(ctx, req)
					}
					info := &ggrpc.UnaryServerInfo{
						Server:     srv,
						FullMethod: methodPing,
					}
					handler := func(ctx context.Context, req any) (any, error) {
						return impl.Ping(ctx, req.(*emptypb.Empty))
					}
					return interceptor(ctx, req, info, handler)
				},
			},
		},
		Streams:  []ggrpc.StreamDesc{},
		Metadata: "gateway_runtime_manual",
	}, &pingService{})
}

func NewServer(address string) (*infragrpc.Server, error) {
	return infragrpc.NewServer(
		infragrpc.ServerConfig{
			Address:               address,
			EnableReflection:      true,
			RegisterHealthService: true,
		},
		func(gs *ggrpc.Server) {
			registerPingService(gs)
		},
	)
}
