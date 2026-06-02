package config

import kconfig "github.com/liukunxin/go-infra/pkg/base/config"

func Load() (*App, error) {
	return kconfig.Load[App](
		kconfig.WithEnvFrom("env"),
		kconfig.WithValidate(true),
		kconfig.WithTagValidation(true),
	)
}

