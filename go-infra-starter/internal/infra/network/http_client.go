package network

import (
	httpclient "github.com/liukunxin/go-infra/pkg/infra/http_client"
	"go-infra-starter/internal/infra/config"
)

type HTTPClientState struct {
	Enabled bool
}

func InitHTTPClient(cfg *config.App) *HTTPClientState {
	if !cfg.Features.HTTPClient {
		return &HTTPClientState{Enabled: false}
	}
	httpclient.Init(cfg.HTTP)
	return &HTTPClientState{Enabled: true}
}

