package config

import kconfig "github.com/liukunxin/go-infra/pkg/base/config"

// Load 读取 configs/ 下的配置，并按 env 变量叠加 config.<env>.yml。
func Load() (*App, error) {
	return kconfig.Load[App](
		kconfig.WithEnvFrom("env"),
		kconfig.WithValidate(true),
		kconfig.WithTagValidation(true),
	)
}
