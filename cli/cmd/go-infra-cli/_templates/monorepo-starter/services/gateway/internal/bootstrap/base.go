package bootstrap

import (
	"os"

	"github.com/liukunxin/go-infra/pkg/base/env"
	"github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"monorepo-starter/services/gateway/internal/infra/config"
)

// loadBaseConfig 读取配置并初始化基线能力（应用名、环境、日志、链路）。
// 所有进程入口（cmd/http、cmd/grpc）都从这里启动，保证启动顺序一致。
func loadBaseConfig() (*config.App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err = initBase(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func initBase(cfg *config.App) error {
	env.SetName(cfg.AppName)
	env.SetEnv(os.Getenv("env"))

	if err := log.Init(cfg.Log); err != nil {
		return err
	}
	traceCfg := cfg.Trace
	serviceName := cfg.AppName
	if traceCfg.ServiceName != nil && *traceCfg.ServiceName != "" {
		serviceName = *traceCfg.ServiceName
	}
	traceCfg.ServiceName = &serviceName
	return trace.Init(trace.WithConfig(&traceCfg))
}

// closeBase 与 initBase 对称，进程退出前统一收尾。
func closeBase() {
	trace.Flush()
	log.Close()
}
