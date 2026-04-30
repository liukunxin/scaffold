package runtime

import (
	"github.com/liukunxin/go-infra/pkg/infra/pprof"
	"go-infra-starter/internal/infra/config"
)

func InitPprof(cfg *config.App) {
	if cfg.Pprof.Enabled {
		pprof.Start()
	}
}

