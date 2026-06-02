package bootstrap

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/base/env"
	"github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"go-infra-monorepo-starter/apps/gateway/internal/handler"
	"go-infra-monorepo-starter/apps/gateway/internal/repository"
	"go-infra-monorepo-starter/apps/gateway/internal/service"
	httptransport "go-infra-monorepo-starter/apps/gateway/transport/http"
	wstransport "go-infra-monorepo-starter/apps/gateway/transport/ws"
	demoapi "go-infra-monorepo-starter/domains/domain-demo/api"
	demoruntime "go-infra-monorepo-starter/domains/domain-demo/runtime"
	demoservice "go-infra-monorepo-starter/domains/domain-demo/service"
	"go-infra-monorepo-starter/internal/event"
	runtimecore "go-infra-monorepo-starter/internal/runtime"
)

type App struct {
	cfg  *Config
	http *httptransport.Server
}

func New() (*App, error) {
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

	dispatcher := event.NewDispatcher()
	dispatcher.Register(demoapi.EventPing, demoruntime.NewHandler(demoservice.NewCore()))
	router := runtimecore.NewRouter(dispatcher)
	engine := runtimecore.NewEngine(router)

	configRepo := repository.NewConfigRepository(cfg.Runtime.DefaultPingName)
	runtimeService := service.NewRuntimeService(engine, configRepo)
	runtimeHandler := handler.NewRuntimeHandler(runtimeService)
	wsServer := wstransport.NewServer()
	httpSrv := httptransport.New(gin.New(), runtimeHandler, wsServer)
	return &App{
		cfg:  cfg,
		http: httpSrv,
	}, nil
}

func (a *App) Run() error {
	return a.http.Run(a.cfg.Server.Address)
}

func (a *App) Close() {
	trace.Flush()
	log.Close()
}
