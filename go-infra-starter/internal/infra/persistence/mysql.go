package persistence

import (
	"github.com/liukunxin/go-infra/pkg/infra/mysql"
	"go-infra-starter/internal/infra/config"
)

type MySQLState struct {
	Enabled bool
}

func InitMySQL(cfg *config.App) (*MySQLState, error) {
	if cfg.Mysql.DSN == "" {
		return &MySQLState{Enabled: false}, nil
	}
	if err := mysql.Init(cfg.Mysql); err != nil {
		return nil, err
	}
	return &MySQLState{Enabled: true}, nil
}

func (s *MySQLState) Close() {
	if s == nil || !s.Enabled {
		return
	}
	_ = mysql.GetClient().Close()
}

