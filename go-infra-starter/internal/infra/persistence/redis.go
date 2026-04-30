package persistence

import (
	redisv8 "github.com/liukunxin/go-infra/pkg/infra/redis/v8"
	"go-infra-starter/internal/infra/config"
)

type RedisState struct {
	Enabled bool
}

func InitRedis(cfg *config.App) (*RedisState, error) {
	if !cfg.Features.Redis {
		return &RedisState{Enabled: false}, nil
	}
	if err := redisv8.Init(&cfg.Redis); err != nil {
		return nil, err
	}
	return &RedisState{Enabled: true}, nil
}

func (s *RedisState) Close() {
	if s == nil || !s.Enabled {
		return
	}
	_ = redisv8.GetClient().Close()
}

