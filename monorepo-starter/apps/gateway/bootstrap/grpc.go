package bootstrap

import (
	"fmt"
	"os"

	"github.com/liukunxin/go-infra/pkg/base/env"
	"github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	infragrpc "github.com/liukunxin/go-infra/pkg/infra/grpc"
	grpctransport "monorepo-starter/apps/gateway/transport/grpc"
)

type GRPCApp struct {
	server *infragrpc.Server
}

func NewGRPC() (*GRPCApp, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	env.SetName(cfg.AppName)
	env.SetEnv(os.Getenv("env"))
	if err = log.Init(cfg.Log); err != nil {
		return nil, err
	}
	traceCfg := cfg.Trace
	serviceName := cfg.AppName
	if traceCfg.ServiceName != nil && *traceCfg.ServiceName != "" {
		serviceName = *traceCfg.ServiceName
	}
	traceCfg.ServiceName = &serviceName
	if err = trace.Init(trace.WithConfig(&traceCfg)); err != nil {
		return nil, err
	}

	server, err := grpctransport.NewServer(cfg.GRPC.Address)
	if err != nil {
		return nil, fmt.Errorf("init grpc server: %w", err)
	}
	return &GRPCApp{server: server}, nil
}

func (a *GRPCApp) Run() error {
	return a.server.Start()
}

func (a *GRPCApp) Close() {
	_ = infragrpc.ShutdownWithTimeout(a.server, 0)
	trace.Flush()
	log.Close()
}
