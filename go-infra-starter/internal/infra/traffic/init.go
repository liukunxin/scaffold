package traffic

import (
	"go-infra-starter/internal/infra/config"

	kitraffic "github.com/liukunxin/go-infra/pkg/infra/traffic"
	"golang.org/x/time/rate"
)

type TrafficState struct {
	Enabled bool
}

func InitTraffic(cfg *config.App) error {
	if !cfg.Features.Traffic {
		return kitraffic.Init(kitraffic.WithController(&kitraffic.DummyController{}))
	}

	limit := cfg.Traffic.RateLimitQPS
	if limit <= 0 {
		limit = 200
	}
	burst := cfg.Traffic.RateLimitBurst
	if burst <= 0 {
		burst = 50
	}

	controller := kitraffic.NewRateLimitController(rate.Limit(limit), burst)
	return kitraffic.Init(kitraffic.WithController(controller))
}
