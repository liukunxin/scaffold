package bootstrap

import (
	"fmt"

	infragrpc "github.com/liukunxin/go-infra/pkg/infra/grpc"
	"single-starter/internal/infra/config"
	"single-starter/internal/route"
)

type GRPCApp struct {
	cfg    *config.App
	server *infragrpc.Server
}

func NewGRPC() (*GRPCApp, error) {
	cfg, err := loadBaseConfig()
	if err != nil {
		return nil, err
	}

	srv, err := infragrpc.NewServer(
		infragrpc.ServerConfig{
			Address:               cfg.GRPC.Address,
			EnableReflection:      true,
			RegisterHealthService: true,
		},
		route.RegisterGRPC,
	)
	if err != nil {
		return nil, fmt.Errorf("init grpc server: %w", err)
	}

	return &GRPCApp{
		cfg:    cfg,
		server: srv,
	}, nil
}

func (a *GRPCApp) Run() error {
	return a.server.Start()
}

func (a *GRPCApp) Close() {
	_ = infragrpc.ShutdownWithTimeout(a.server, 0)
	closeBase()
}
