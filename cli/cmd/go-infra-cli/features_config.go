package main

import (
	"fmt"
	"strings"
)

// configFeatureKeys 为可通过 add/remove/init 安装的配置型能力。
var configFeatureKeys = map[string]string{
	"mysql":       "mysql",
	"redis":       "redis",
	"metrics":     "metrics",
	"pprof":       "pprof",
	"http-client": "http_client",
	"traffic":     "traffic",
}

var structuralFeatureKeys = map[string]struct{}{
	"llm": {},
}

func parseConfigFeaturesArg(features string) ([]string, error) {
	features = strings.TrimSpace(features)
	if features == "" {
		return nil, fmt.Errorf("at least one feature is required, allowed: %s", strings.Join(sortedConfigFeatureNames(), ","))
	}

	seen := make(map[string]struct{})
	var ordered []string
	for _, item := range strings.Split(features, ",") {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		if _, structural := structuralFeatureKeys[key]; structural {
			return nil, fmt.Errorf("feature %q requires code overlay; use `init --features %s` or manage files manually", key, key)
		}
		if _, ok := configFeatureKeys[key]; !ok {
			return nil, fmt.Errorf("unsupported feature %q, allowed: %s", key, strings.Join(sortedConfigFeatureNames(), ","))
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		ordered = append(ordered, key)
	}
	if len(ordered) == 0 {
		return nil, fmt.Errorf("at least one feature is required, allowed: %s", strings.Join(sortedConfigFeatureNames(), ","))
	}
	return ordered, nil
}

func parseInitFeaturesArg(features string) (map[string]bool, bool, error) {
	flags := map[string]bool{
		"mysql":       false,
		"redis":       false,
		"metrics":     false,
		"pprof":       false,
		"http-client": false,
		"traffic":     false,
	}
	withLLM := false

	if strings.TrimSpace(features) == "" {
		return flags, withLLM, nil
	}

	for _, item := range strings.Split(features, ",") {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		if key == "llm" {
			withLLM = true
			continue
		}
		if _, ok := configFeatureKeys[key]; !ok {
			return nil, false, fmt.Errorf("unsupported feature %q, allowed: mysql,redis,metrics,pprof,http-client,traffic,llm", key)
		}
		flags[key] = true
	}
	return flags, withLLM, nil
}

func sortedConfigFeatureNames() []string {
	names := make([]string, 0, len(configFeatureKeys))
	for name := range configFeatureKeys {
		names = append(names, name)
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}
