package main

import (
	"strings"
	"testing"
)

func TestParseInitFeaturesArg_EmptyMeansNone(t *testing.T) {
	flags, withLLM, err := parseInitFeaturesArg("")
	if err != nil {
		t.Fatal(err)
	}
	if withLLM {
		t.Fatal("expected llm false")
	}
	for name, enabled := range flags {
		if enabled {
			t.Fatalf("expected %s false by default", name)
		}
	}
}

func TestParseInitFeaturesArg_Selective(t *testing.T) {
	flags, withLLM, err := parseInitFeaturesArg("mysql,redis,llm")
	if err != nil {
		t.Fatal(err)
	}
	if !withLLM || !flags["mysql"] || !flags["redis"] || flags["metrics"] {
		t.Fatalf("unexpected flags: %+v llm=%v", flags, withLLM)
	}
}

// init 的 flags 必须覆盖全部配置型能力，否则 syncAllFeatures 会漏装。
func TestParseInitFeaturesArg_CoversEveryConfigFeature(t *testing.T) {
	flags, _, err := parseInitFeaturesArg("")
	if err != nil {
		t.Fatal(err)
	}
	for name := range configFeatureKeys {
		if _, ok := flags[name]; !ok {
			t.Fatalf("config feature %q missing from init flags", name)
		}
	}
	if len(flags) != len(configFeatureKeys) {
		t.Fatalf("init flags = %d entries, configFeatureKeys = %d", len(flags), len(configFeatureKeys))
	}
}

func TestParseInitFeaturesArg_ErrorMentionsLLM(t *testing.T) {
	_, _, err := parseInitFeaturesArg("nope")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "llm") {
		t.Fatalf("error should list llm: %v", err)
	}
}

// llm 是覆盖文件型能力，但从使用者角度看它和 redis 一样只是「一项能力」。
// 曾经 add/remove 直接拒绝 llm，结果是「init 装得上、remove 卸不掉」。
func TestParseFeatureArg_AcceptsBothConfigAndStructuralFeatures(t *testing.T) {
	got, err := parseFeatureArg("redis,llm,mysql")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"redis", "llm", "mysql"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v (order must be preserved)", got, want)
		}
	}
}

func TestParseFeatureArg_Deduplicates(t *testing.T) {
	got, err := parseFeatureArg("llm,llm")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "llm" {
		t.Fatalf("got %v, want [llm]", got)
	}
}

func TestParseFeatureArg_RejectsUnknownAndEmpty(t *testing.T) {
	if _, err := parseFeatureArg("nope"); err == nil {
		t.Fatal("expected error for an unknown feature")
	}
	if _, err := parseFeatureArg("  "); err == nil {
		t.Fatal("expected error for an empty feature list")
	}
}

// 错误提示必须列出全部可安装能力，否则使用者不知道 llm 也能 add。
func TestParseFeatureArg_ErrorListsEveryFeature(t *testing.T) {
	_, err := parseFeatureArg("nope")
	if err == nil {
		t.Fatal("expected error")
	}
	for _, name := range sortedFeatureNames() {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("error should list %q: %v", name, err)
		}
	}
}

// 提示表里的能力名必须真的可安装，否则用户照着提示去配也装不上。
func TestFeatureConfigHints_NamesAreInstallable(t *testing.T) {
	allowed := map[string]bool{}
	for _, name := range sortedFeatureNames() {
		allowed[name] = true
	}
	for feature := range featureConfigHints {
		if !allowed[feature] {
			t.Fatalf("hint registered for unknown feature %q", feature)
		}
	}
}

// 只有「不填参数就不能用」的能力才该有提示。
// redis 模板自带 127.0.0.1:6379，metrics/pprof/http-client/traffic 都有默认值 ——
// 给它们刷提示就是噪音；而 mysql / llm 不提醒，用户根本不知道去哪配。
func TestFeatureConfigHints_Coverage(t *testing.T) {
	for _, feature := range []string{"mysql", "llm"} {
		if addFeatureHintLine(feature) == "" {
			t.Fatalf("%s requires manual config, so add must print a hint", feature)
		}
	}
	for _, feature := range []string{"redis", "metrics", "pprof", "http-client", "traffic"} {
		if line := addFeatureHintLine(feature); line != "" {
			t.Fatalf("%s works out of the box, it must not print a hint, got %q", feature, line)
		}
	}
}

// 提示必须是静态的：曾经它读生成物的 config.yml 决定措辞，结果在不带 mysql 的 init 里打印
// 「no `dsn:` key found」，而模板里 `dsn: ""` 明明就在。
func TestFeatureConfigHints_AreStaticAndPointAtConfig(t *testing.T) {
	for feature := range featureConfigHints {
		line := addFeatureHintLine(feature)
		if !strings.Contains(line, "configs/config.yml") {
			t.Fatalf("%s hint should point at configs/config.yml: %s", feature, line)
		}
		for _, banned := range []string{"currently empty", "is empty in", "no `dsn:` key"} {
			if strings.Contains(line, banned) {
				t.Fatalf("%s hint claims to have inspected the config (%q): %s", feature, banned, line)
			}
		}
	}
}

// init 的 next steps 汇总行：只列「已启用且需要手填」的项，顺序稳定。
func TestConfigHintList(t *testing.T) {
	none, _, err := parseInitFeaturesArg("")
	if err != nil {
		t.Fatal(err)
	}
	if got := configHintList(none, false); len(got) != 0 {
		t.Fatalf("no feature enabled, want no hint line, got %v", got)
	}

	flags, withLLM, err := parseInitFeaturesArg("mysql,redis,llm")
	if err != nil {
		t.Fatal(err)
	}
	// redis 开箱可用，不进列表；mysql 与 llm 都要，且按能力名排序（llm < mysql）。
	want := []string{featureConfigHints["llm"], featureConfigHints["mysql"]}
	got := configHintList(flags, withLLM)
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v (order must be stable)", got, want)
		}
	}
}
