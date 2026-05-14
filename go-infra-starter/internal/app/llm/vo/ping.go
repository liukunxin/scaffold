package vo

type PingResp struct {
	Enabled bool   `json:"enabled"`
	Reply   string `json:"reply"`
}
