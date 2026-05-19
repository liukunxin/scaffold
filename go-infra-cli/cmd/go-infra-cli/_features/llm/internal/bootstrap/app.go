package bootstrap

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/base/env"
	"github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"go-infra-starter/internal/infra/ai"
	"go-infra-starter/internal/infra/config"
	"go-infra-starter/internal/route"
	// FEATURE_IMPORTS_START
	// FEATURE_IMPORTS_END
)

type App struct {
	cfg    *config.App
	router *gin.Engine
}

func New() (*App, error) {
	cfg, err := config.Load()
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
	// FEATURE_INIT_START
	// FEATURE_INIT_END
	if err = ai.InitLLM(cfg); err != nil {
		return nil, err
	}

	router := gin.New()
	// FEATURE_ROUTER_START
	// FEATURE_ROUTER_END
	route.Setup(router)

	return &App{
		cfg:    cfg,
		router: router,
	}, nil
}

func (a *App) Run() error {
	return a.router.Run(a.cfg.Server.Address)
}

func (a *App) Close() {
	// FEATURE_CLOSE_START
	// FEATURE_CLOSE_END
	trace.Flush()
	log.Close()
}
