package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runAdd(args []string) error {
	return runConfigFeatureToggle(args, true)
}

func runRemove(args []string) error {
	return runConfigFeatureToggle(args, false)
}

// syncFeature 按能力类型分派：接线型只改锚点，覆盖型还要落地/回收 _features 文件。
func syncFeature(projectDir, feature string, enable bool) (bool, error) {
	if isStructuralFeature(feature) {
		return syncStructuralFeature(projectDir, feature, enable)
	}
	return syncConfigFeatureArtifacts(projectDir, feature, enable)
}

func runConfigFeatureToggle(args []string, enable bool) error {
	cmdName := "add"
	if !enable {
		cmdName = "remove"
	}

	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	projectDir := fs.String("dir", "", "project root directory (default: auto-detect from current directory)")

	var featuresArg string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		featuresArg = strings.TrimSpace(args[0])
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
	} else {
		if err := fs.Parse(args); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return fmt.Errorf("feature list is required, e.g. go-infra-cli %s redis", cmdName)
		}
		featuresArg = strings.TrimSpace(fs.Arg(0))
	}

	features, err := parseFeatureArg(featuresArg)
	if err != nil {
		return err
	}

	root, err := resolveProjectDir(*projectDir)
	if err != nil {
		return err
	}

	anyChanged := false
	for _, feature := range features {
		changed, err := syncFeature(root, feature, enable)
		if err != nil {
			return fmt.Errorf("%s: %w", feature, err)
		}
		if changed {
			anyChanged = true
			fmt.Printf("%s: %s\n", feature, featureState(enable))
		} else if isFeatureInstalled(root, feature) == enable {
			fmt.Printf("%s: already %s (skipped)\n", feature, featureState(enable))
		}
		// 只在真的接线了才提醒「去哪配参数」；已经装过的不重复刷。
		if enable && changed {
			printAddFeatureHints(feature)
		}
	}

	// 兜底：若覆盖文件里还带着模板时代写死的 import 前缀，改回当前 module 路径。
	if err := rewriteStaleStarterImports(root); err != nil {
		return fmt.Errorf("rewrite module paths: %w", err)
	}
	if err := formatGoSources(root); err != nil {
		return fmt.Errorf("format go sources: %w", err)
	}

	// 新增能力往往带来新依赖（如 traffic 需要 golang.org/x/time/rate），
	// 只改代码不同步依赖的话，下一条 go build 必然失败。与 init 对齐：改了就收敛依赖并自检。
	if anyChanged {
		if err := runGoModTidy(root); err != nil {
			return err
		}
		if err := verifyGeneratedProject(root); err != nil {
			return err
		}
	}

	if !anyChanged {
		fmt.Println("no file changes (already in target state)")
	} else {
		fmt.Printf("\nproject: %s\nnext: adjust configs if needed and restart the service\n", root)
	}
	return nil
}

func resolveProjectDir(explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		if err := validateScaffoldProject(abs); err != nil {
			return "", err
		}
		return abs, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if err := validateScaffoldProject(wd); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", errors.New("cannot find a single-starter style project (need go.mod, configs/config.yml, and standard internal layout); use --dir")
}

func validateScaffoldProject(dir string) error {
	if err := validateProjectRoot(dir); err != nil {
		return err
	}
	return validateStarterLayout(dir)
}

func validateProjectRoot(dir string) error {
	if !fileExists(filepath.Join(dir, "go.mod")) {
		return errors.New("missing go.mod")
	}
	if !fileExists(filepath.Join(dir, "configs", "config.yml")) {
		return errors.New("missing configs/config.yml")
	}
	return nil
}

// featureConfigHints 是「能力 → 必须人工填写的配置项」对照表，只列不填参数就用不起来的能力。
// redis 模板自带可用的 127.0.0.1:6379，metrics/pprof/http-client/traffic 都有默认值，
// 都不进来 —— 否则每 add 一次就刷一堆噪音。
//
// 刻意不读生成物的 configs/config.yml：刚生成的项目配置必然是空的，读它判断不出任何新信息。
// 「没配置就不启用」由注入的接线自己用运行时守卫表达（见 feature_sync.go 的 configFeatureSpecs：
// `if cfg.Mysql.DSN != ""`、`if len(cfg.Redis.Addresses) > 0`），不需要 CLI 替它把关。
var featureConfigHints = map[string]string{
	"mysql": "mysql.dsn",
	"llm":   "llm.default_provider / llm.providers",
}

// addFeatureHintLine 生成 add 成功后的单行提示；返回空表示该能力开箱可用、不需要提醒。
func addFeatureHintLine(feature string) string {
	setting, ok := featureConfigHints[feature]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s: wired; set %s in configs/config.yml to enable it (left empty = feature off)", feature, setting)
}

func printAddFeatureHints(feature string) {
	if line := addFeatureHintLine(feature); line != "" {
		fmt.Println(line)
	}
}

// configHintList 返回已启用能力里需要手填配置的项，顺序稳定（按能力名排序），
// 供 init 的 next steps 汇总成一行。
func configHintList(configFlags map[string]bool, withLLM bool) []string {
	enabled := make(map[string]bool, len(configFlags)+1)
	for name, on := range configFlags {
		enabled[name] = on
	}
	if withLLM {
		enabled["llm"] = true
	}

	var out []string
	for _, name := range sortedFeatureNames() {
		if setting, ok := featureConfigHints[name]; ok && enabled[name] {
			out = append(out, setting)
		}
	}
	return out
}
