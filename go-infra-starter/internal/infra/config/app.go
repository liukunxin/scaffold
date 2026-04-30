package config

import (
	"fmt"

	klog "github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"github.com/liukunxin/go-infra/pkg/infra/mysql"
	redis "github.com/liukunxin/go-infra/pkg/infra/redis/v8"
)

type App struct {
	AppName string       `yaml:"app_name" validate:"required"`
	Server  ServerConfig `yaml:"server" validate:"required"`

	// 直接复用 go-infra 内置配置结构，避免脚手架重复定义。
	Log   klog.Config   `yaml:"log"`
	Trace trace.Config  `yaml:"trace"`
	Mysql mysql.Config  `yaml:"mysql"`
	Redis redis.Config  `yaml:"redis"`

	// 以下为项目级能力开关（编排层），不是 go-infra SDK 内置 Config。
	Features FeaturesConfig `yaml:"features"`
}

type ServerConfig struct {
	Address string `yaml:"address" validate:"required"`
}

type FeaturesConfig struct {
	Redis   bool `yaml:"redis"`
	Metrics bool `yaml:"metrics"`
	Pprof   bool `yaml:"pprof"`
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

