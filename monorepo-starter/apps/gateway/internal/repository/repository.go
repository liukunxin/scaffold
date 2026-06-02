package repository

import "strings"

// ConfigRepository provides app-scoped default config values.
type ConfigRepository interface {
	DefaultPingName() string
}

type staticConfigRepository struct {
	defaultPingName string
}

func NewConfigRepository(defaultPingName string) ConfigRepository {
	return &staticConfigRepository{
		defaultPingName: strings.TrimSpace(defaultPingName),
	}
}

func (r *staticConfigRepository) DefaultPingName() string {
	if r.defaultPingName != "" {
		return r.defaultPingName
	}
	return "go-infra"
}
