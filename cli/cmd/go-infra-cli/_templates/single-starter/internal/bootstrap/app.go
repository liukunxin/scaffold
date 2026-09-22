package bootstrap

import (
	"github.com/gin-gonic/gin"
	"single-starter/internal/infra/config"
	"single-starter/internal/route"
	// FEATURE_IMPORTS_START
	// FEATURE_IMPORTS_END
)

type App struct {
	cfg    *config.App
	router *gin.Engine
}

func New() (*App, error) {
	cfg, err := loadBaseConfig()
	if err != nil {
		return nil, err
	}
	// FEATURE_INIT_START
	// FEATURE_INIT_END

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
	closeBase()
}
