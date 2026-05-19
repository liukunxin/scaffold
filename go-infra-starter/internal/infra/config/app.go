package config

import (
	"fmt"

	klog "github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"github.com/liukunxin/go-infra/pkg/infra/http_client"
	"github.com/liukunxin/go-infra/pkg/infra/llm"
	"github.com/liukunxin/go-infra/pkg/infra/mysql"
	redis "github.com/liukunxin/go-infra/pkg/infra/redis/v8"
)

type App struct {
	AppName string       `yaml:"app_name" validate:"required"`
	Server  ServerConfig `yaml:"server" validate:"required"`

	// 直接复用 go-infra 内置配置结构，避免脚手架重复定义。
	Log     klog.Config        `yaml:"log"`
	Trace   trace.Config       `yaml:"trace"`
	Mysql   mysql.Config       `yaml:"mysql"`
	Redis   redis.Config       `yaml:"redis"`
	HTTP    http_client.Config `yaml:"http_client"`
	Traffic TrafficConfig      `yaml:"traffic"`
	LLM     llm.Config         `yaml:"llm"`
}

type ServerConfig struct {
	Address string `yaml:"address" validate:"required"`
}

type TrafficConfig struct {
	RateLimitQPS   float64 `yaml:"rate_limit_qps"`
	RateLimitBurst int     `yaml:"rate_limit_burst"`
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
