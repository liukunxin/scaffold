package bootstrap

import (
	kconfig "github.com/liukunxin/go-infra/pkg/base/config"
	klog "github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
)

type Config struct {
	AppName string `yaml:"app_name"`
	Server  struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
	GRPC struct {
		Address string `yaml:"address"`
	} `yaml:"grpc"`
	Runtime struct {
		DefaultPingName string `yaml:"default_ping_name"`
	} `yaml:"runtime"`
	Log   klog.Config  `yaml:"log"`
	Trace trace.Config `yaml:"trace"`
}

func loadConfig() (*Config, error) {
	return kconfig.Load[Config](kconfig.WithValidate(false))
}
