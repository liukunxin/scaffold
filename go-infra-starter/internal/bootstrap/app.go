package bootstrap

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/base/env"
	"go-infra-starter/internal/infra/ai"
	"go-infra-starter/internal/infra/config"
	"go-infra-starter/internal/infra/network"
	"go-infra-starter/internal/infra/observability"
	"go-infra-starter/internal/infra/persistence"
	kruntime "go-infra-starter/internal/infra/runtime"
	inftraffic "go-infra-starter/internal/infra/traffic"
	"go-infra-starter/internal/route"
)

type App struct {
	cfg        *config.App
	router     *gin.Engine
	mysql      *persistence.MySQLState
	redis      *persistence.RedisState
	httpClient *network.HTTPClientState
	llm        *ai.LLMState
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// 为 go-infra 的其他模块提供统一环境上下文。
	env.SetName(cfg.AppName)
	env.SetEnv(os.Getenv("env"))

	if err = observability.InitLog(cfg); err != nil {
		return nil, err
	}
	if err = observability.InitTrace(cfg); err != nil {
		return nil, err
	}

	mysqlState := &persistence.MySQLState{Enabled: false}
	if cfg.Features.MySQL {
		mysqlState, err = persistence.InitMySQL(cfg)
		if err != nil {
			return nil, err
		}
	}
	redisState := &persistence.RedisState{Enabled: false}
	if cfg.Features.Redis {
		redisState, err = persistence.InitRedis(cfg)
		if err != nil {
			return nil, err
		}
	}
	httpClientState := network.InitHTTPClient(cfg)
	if err = inftraffic.InitTraffic(cfg); err != nil {
		return nil, err
	}
	llmState, err := ai.InitLLM(cfg)
	if err != nil {
		return nil, err
	}

	kruntime.InitPprof(cfg)

	router := gin.New()
	observability.InitMetrics(cfg, router)
	route.Setup(router)

	return &App{
		cfg:        cfg,
		router:     router,
		mysql:      mysqlState,
		redis:      redisState,
		httpClient: httpClientState,
		llm:        llmState,
	}, nil
}

func (a *App) Run() error {
	return a.router.Run(a.cfg.Server.Address)
}

func (a *App) Close() {
	a.mysql.Close()
	a.redis.Close()
	observability.Close()
}
