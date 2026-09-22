package main

import (
	"fmt"
	"sort"
	"strings"
)

// configFeatureKeys 是「纯接线型」能力：只在 bootstrap / route 的锚点里插入几行
// 接线代码，不产生新文件。
var configFeatureKeys = map[string]string{
	"mysql":       "mysql",
	"redis":       "redis",
	"metrics":     "metrics",
	"pprof":       "pprof",
	"http-client": "http_client",
	"traffic":     "traffic",
}

// structuralFeatureKeys 是「覆盖文件型」能力：除锚点接线外，还要把 _features/<name>/
// 下的整套业务/适配代码拷进项目。add 装、remove 卸，两个方向必须对称——
// 只允许 init 安装的能力等于"装上就卸不掉"，那是 bug 不是特性。
var structuralFeatureKeys = map[string]struct{}{
	"llm": {},
}

func isStructuralFeature(name string) bool {
	_, ok := structuralFeatureKeys[name]
	return ok
}

func isInstallableFeature(name string) bool {
	if isStructuralFeature(name) {
		return true
	}
	_, ok := configFeatureKeys[name]
	return ok
}

// parseFeatureArg 解析 add/remove 的能力列表。配置型与覆盖型都接受，
// 因为从使用者的角度看两者都是"一项能力"，区别只是实现方式。
func parseFeatureArg(features string) ([]string, error) {
	features = strings.TrimSpace(features)
	if features == "" {
		return nil, fmt.Errorf("at least one feature is required, allowed: %s", strings.Join(sortedFeatureNames(), ","))
	}

	seen := make(map[string]struct{})
	var ordered []string
	for _, item := range strings.Split(features, ",") {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		if !isInstallableFeature(key) {
			return nil, fmt.Errorf("unsupported feature %q, allowed: %s", key, strings.Join(sortedFeatureNames(), ","))
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		ordered = append(ordered, key)
	}
	if len(ordered) == 0 {
		return nil, fmt.Errorf("at least one feature is required, allowed: %s", strings.Join(sortedFeatureNames(), ","))
	}
	return ordered, nil
}

// parseInitFeaturesArg 解析 init --features。flags 直接由 configFeatureKeys 派生，
// 避免两处清单各自维护导致漂移；llm 是覆盖型能力，单独返回。
func parseInitFeaturesArg(features string) (map[string]bool, bool, error) {
	flags := make(map[string]bool, len(configFeatureKeys))
	for name := range configFeatureKeys {
		flags[name] = false
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
		if _, ok := configFeatureKeys[key]; ok {
			flags[key] = true
			continue
		}
		if isStructuralFeature(key) {
			// 目前覆盖型只有 llm 一个；将来加第二个时这里要改成 map[string]bool。
			withLLM = true
			continue
		}
		return nil, false, fmt.Errorf("unsupported feature %q, allowed: %s", key, strings.Join(sortedFeatureNames(), ","))
	}
	return flags, withLLM, nil
}

// sortedFeatureNames 返回全部可安装能力名，供错误提示复用。
func sortedFeatureNames() []string {
	names := make([]string, 0, len(configFeatureKeys)+len(structuralFeatureKeys))
	for name := range configFeatureKeys {
		names = append(names, name)
	}
	for name := range structuralFeatureKeys {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
