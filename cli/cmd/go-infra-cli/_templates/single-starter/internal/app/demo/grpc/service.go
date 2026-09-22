package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	demodto "single-starter/internal/app/demo/dto"
	demoservice "single-starter/internal/app/demo/service"
	"single-starter/internal/app/demo/vo"
)

const (
	serviceName = "demo.DemoService"
	methodPing  = "/demo.DemoService/Ping"
)

// Service is a protobuf-free gRPC demo service based on built-in proto types.
type Service struct {
	demoService demoservice.DemoService
}

type Server interface {
	Ping(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
}

func NewService(demoService demoservice.DemoService) *Service {
	return &Service{demoService: demoService}
}

func (s *Service) Ping(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	resp, err := s.demoService.Ping(ctx, demodto.PingInput{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return wrapperspb.String(resp.Message), nil
}

func Register(server grpc.ServiceRegistrar, svc Server) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*Server)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Ping",
				Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
					req := &emptypb.Empty{}
					if err := dec(req); err != nil {
						return nil, err
					}
					impl := srv.(Server)
					if interceptor == nil {
						return impl.Ping(ctx, req)
					}
					info := &grpc.UnaryServerInfo{
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
		Streams:  []grpc.StreamDesc{},
		Metadata: "demo_service_manual",
	}, svc)
}

type Client interface {
	Ping(ctx context.Context) (*vo.PingResp, error)
}

type client struct {
	conn *grpc.ClientConn
}

func NewClient(conn *grpc.ClientConn) Client {
	return &client{conn: conn}
}

func (c *client) Ping(ctx context.Context) (*vo.PingResp, error) {
	out := new(wrapperspb.StringValue)
	if err := c.conn.Invoke(ctx, methodPing, &emptypb.Empty{}, out); err != nil {
		return nil, fmt.Errorf("grpc ping invoke: %w", err)
	}
	return &vo.PingResp{Message: out.Value}, nil
}
