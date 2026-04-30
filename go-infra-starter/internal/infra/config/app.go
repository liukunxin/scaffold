package config

import (
	"fmt"

	klog "github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/infra/mysql"
	redisv8 "github.com/liukunxin/go-infra/pkg/infra/redis/v8"
)

type App struct {
	AppName string        `yaml:"app_name" validate:"required"`
	Server  ServerConfig  `yaml:"server" validate:"required"`
	Log     klog.Config   `yaml:"log"`
	Trace   TraceConfig   `yaml:"trace"`
	Metrics MetricsConfig `yaml:"metrics"`
	Pprof   PprofConfig   `yaml:"pprof"`
	Mysql   mysql.Config  `yaml:"mysql"`
	Redis   RedisConfig   `yaml:"redis"`
}

type ServerConfig struct {
	Address string `yaml:"address" validate:"required"`
}

type TraceConfig struct {
	ServiceName string   `yaml:"service_name"`
	SampleRatio *float64 `yaml:"sample_ratio"`
}

type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type PprofConfig struct {
	Enabled bool `yaml:"enabled"`
}

type RedisConfig struct {
	Enabled bool           `yaml:"enabled"`
	Config  redisv8.Config `yaml:"config"`
}

func (a *App) Validate() error {
	if a.AppName == "" {
		return fmt.Errorf("app_name is required")
	}
	if a.Server.Address == "" {
		return fmt.Errorf("server.address is required")
	}
	return nil
}

