package observability

import (
	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"github.com/liukunxin/go-infra/pkg/infra/metrics"
	"go-infra-starter/internal/infra/config"
)

func InitLog(cfg *config.App) error {
	return log.Init(cfg.Log)
}

func InitTrace(cfg *config.App) error {
	serviceName := cfg.Trace.ServiceName
	if serviceName == "" {
		serviceName = cfg.AppName
	}
	traceCfg := &trace.Config{
		ServiceName: &serviceName,
		SampleRatio: cfg.Trace.SampleRatio,
	}
	return trace.Init(trace.WithConfig(traceCfg))
}

func InitMetrics(cfg *config.App, router *gin.Engine) {
	if cfg.Metrics.Enabled {
		metrics.InitMetrics(cfg.AppName, router)
	}
}

func Close() {
	trace.Flush()
	log.Close()
}

