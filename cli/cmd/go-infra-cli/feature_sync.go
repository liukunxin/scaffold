package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// featureSection 描述注入到某对锚点之间的一段代码。缩进统一按一级 tab
// （bootstrap 里是 import 块 / 函数体，route 里是 Setup 函数体）。
type featureSection struct {
	name  string
	start string
	end   string
	lines []string
}

// featureTarget 描述一个 feature 要改的文件，以及在该文件里插入的各段。
type featureTarget struct {
	file     string
	sections []featureSection
}

func bootstrapTarget(sections ...featureSection) featureTarget {
	return featureTarget{file: bootstrapFilePath, sections: sections}
}

func routeTarget(sections ...featureSection) featureTarget {
	return featureTarget{file: routeFilePath, sections: sections}
}

// configFeatureSpecs 是 add/remove/init 可安装的配置型能力。
// 每个能力只声明「往哪些锚点插哪些行」，文件定位与幂等标记由 syncFeatureArtifacts 处理。
var configFeatureSpecs = map[string][]featureTarget{
	"mysql": {bootstrapTarget(
		featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
			`"github.com/liukunxin/go-infra/pkg/infra/mysql"`,
		}},
		featureSection{name: "INIT", start: anchorInitStart, end: anchorInitEnd, lines: []string{
			`if cfg.Mysql.DSN != "" {`,
			`	if err = mysql.Init(cfg.Mysql); err != nil {`,
			`		return nil, err`,
			`	}`,
			`}`,
		}},
		featureSection{name: "CLOSE", start: anchorCloseStart, end: anchorCloseEnd, lines: []string{
			`if a.cfg.Mysql.DSN != "" {`,
			`	_ = mysql.GetClient().Close()`,
			`}`,
		}},
	)},
	"redis": {bootstrapTarget(
		featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
			`iredis "github.com/liukunxin/go-infra/pkg/infra/redis"`,
		}},
		featureSection{name: "INIT", start: anchorInitStart, end: anchorInitEnd, lines: []string{
			`if len(cfg.Redis.Addresses) > 0 {`,
			`	if err = iredis.Init(&cfg.Redis); err != nil {`,
			`		return nil, err`,
			`	}`,
			`}`,
		}},
		featureSection{name: "CLOSE", start: anchorCloseStart, end: anchorCloseEnd, lines: []string{
			`if len(a.cfg.Redis.Addresses) > 0 {`,
			`	_ = iredis.GetClient().Close()`,
			`}`,
		}},
	)},
	"http-client": {bootstrapTarget(
		featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
			`httpclient "github.com/liukunxin/go-infra/pkg/infra/http_client"`,
		}},
		featureSection{name: "INIT", start: anchorInitStart, end: anchorInitEnd, lines: []string{
			`httpclient.Init(cfg.HTTP)`,
		}},
	)},
	"traffic": {bootstrapTarget(
		featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
			`kitraffic "github.com/liukunxin/go-infra/pkg/infra/traffic"`,
			`"golang.org/x/time/rate"`,
		}},
		featureSection{name: "INIT", start: anchorInitStart, end: anchorInitEnd, lines: []string{
			`limit := cfg.Traffic.RateLimitQPS`,
			`if limit <= 0 {`,
			`	limit = 200`,
			`}`,
			`burst := cfg.Traffic.RateLimitBurst`,
			`if burst <= 0 {`,
			`	burst = 50`,
			`}`,
			`controller := kitraffic.NewRateLimitController(rate.Limit(limit), burst)`,
			`if err = kitraffic.Init(kitraffic.WithController(controller)); err != nil {`,
			`	return nil, err`,
			`}`,
		}},
	)},
	"pprof": {bootstrapTarget(
		featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
			`"github.com/liukunxin/go-infra/pkg/infra/pprof"`,
		}},
		featureSection{name: "INIT", start: anchorInitStart, end: anchorInitEnd, lines: []string{
			`pprof.Start()`,
		}},
	)},
	"metrics": {bootstrapTarget(
		featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
			`"github.com/liukunxin/go-infra/pkg/infra/metrics"`,
		}},
		featureSection{name: "ROUTER", start: anchorRouterStart, end: anchorRouterEnd, lines: []string{
			`metrics.InitMetrics(cfg.AppName, router)`,
		}},
	)},
}

// llmFeatureTargets 覆盖 `init --features llm`：bootstrap 里初始化 LLM 客户端，
// 路由里挂上 setupLLM（实现由 _features/llm 覆盖文件提供）。
// LLM 业务代码不进模板主线，避免默认项目里带上用不上的依赖。
//
// 注入发生在模板渲染之后，所以这里的 import 必须直接用目标模块路径；
// 写成模板名会残留 `"single-starter/internal/infra/ai"`，生成物直接编译失败。
func llmFeatureTargets(modulePath string) []featureTarget {
	return []featureTarget{
		bootstrapTarget(
			featureSection{name: "IMPORTS", start: anchorImportsStart, end: anchorImportsEnd, lines: []string{
				`"` + modulePath + `/internal/infra/ai"`,
			}},
			featureSection{name: "INIT", start: anchorInitStart, end: anchorInitEnd, lines: []string{
				`if err = ai.InitLLM(cfg); err != nil {`,
				`	return nil, err`,
				`}`,
			}},
		),
		routeTarget(
			featureSection{name: "ROUTES", start: anchorRoutesStart, end: anchorRoutesEnd, lines: []string{
				`setupLLM(api)`,
			}},
		),
	}
}

func syncConfigFeatureArtifacts(projectDir, feature string, enable bool) (changed bool, err error) {
	targets, ok := configFeatureSpecs[feature]
	if !ok {
		return false, fmt.Errorf("unsupported feature spec: %s", feature)
	}
	return syncFeatureArtifacts(projectDir, feature, targets, enable)
}

func syncLLMFeatureArtifacts(projectDir string, enable bool) (bool, error) {
	modulePath, err := readModulePath(projectDir)
	if err != nil {
		return false, err
	}
	return syncFeatureArtifacts(projectDir, "llm", llmFeatureTargets(modulePath), enable)
}

// syncStructuralFeature 处理覆盖文件型能力：add 与 remove 必须对称，
// 否则会出现「装得上、卸不掉」的能力。
func syncStructuralFeature(projectDir, feature string, enable bool) (bool, error) {
	switch feature {
	case "llm":
		return syncLLMFeature(projectDir, enable)
	default:
		return false, fmt.Errorf("unsupported structural feature: %s", feature)
	}
}

func syncLLMFeature(projectDir string, enable bool) (bool, error) {
	if enable {
		// 覆盖文件自带模板模块路径，copyEmbeddedLLMOverlay 写盘时已渲染成目标模块。
		if err := copyEmbeddedLLMOverlay(projectDir); err != nil {
			return false, err
		}
		return syncLLMFeatureArtifacts(projectDir, true)
	}
	// 先拆接线再回收文件：顺序反了的话，拆接线失败会留下"文件在、没人引用"的中间态。
	anchorsChanged, err := syncLLMFeatureArtifacts(projectDir, false)
	if err != nil {
		return anchorsChanged, err
	}
	filesChanged, err := removeEmbeddedLLMOverlay(projectDir)
	return anchorsChanged || filesChanged, err
}

func syncFeatureArtifacts(projectDir, feature string, targets []featureTarget, enable bool) (bool, error) {
	anyChanged := false
	for _, target := range targets {
		path := filepath.Join(projectDir, filepath.FromSlash(target.file))
		data, err := os.ReadFile(path)
		if err != nil {
			return false, fmt.Errorf("read %s: %w", target.file, err)
		}
		content := string(data)
		changed := false
		for _, sec := range target.sections {
			if enable {
				content, changed, err = insertFeatureSection(content, feature, sec, changed)
			} else {
				content, changed = removeFeatureSection(content, feature, sec.name, changed)
			}
			if err != nil {
				return false, fmt.Errorf("%s: %w", target.file, err)
			}
		}
		if !changed {
			continue
		}
		if err = os.WriteFile(path, []byte(content), 0o644); err != nil {
			return false, fmt.Errorf("write %s: %w", target.file, err)
		}
		anyChanged = true
	}
	return anyChanged, nil
}

// isFeatureInstalled 判定能力是否已接线：所有能力的注入块都带
// // FEATURE:<name>:<SEG>:START 标记，命中任一即视为已安装。
func isFeatureInstalled(projectDir, feature string) bool {
	data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(bootstrapFilePath)))
	if err != nil {
		return false
	}
	content := string(data)
	for _, sec := range featureSectionNames {
		if strings.Contains(content, featureSectionStartMarker(feature, sec)) {
			return true
		}
	}
	return false
}

// featureSectionNames 用于"是否已安装"判定，与各 spec 的段名保持一致。
var featureSectionNames = []string{"IMPORTS", "INIT", "ROUTER", "CLOSE", "ROUTES"}

// syncAllFeatures 按 init 的 --features/--scenes 结果注入或移除能力。
func syncAllFeatures(projectDir string, flags map[string]bool, withLLM bool) error {
	for feature, enabled := range flags {
		changed, err := syncConfigFeatureArtifacts(projectDir, feature, enabled)
		if err != nil {
			return err
		}
		if changed {
			fmt.Printf("feature %s: %s\n", feature, featureState(enabled))
		}
	}
	if !withLLM {
		return nil
	}
	changed, err := syncLLMFeatureArtifacts(projectDir, true)
	if err != nil {
		return err
	}
	if changed {
		fmt.Printf("feature llm: %s\n", featureState(true))
	}
	return nil
}

func featureState(enabled bool) string {
	if enabled {
		return "installed"
	}
	return "removed"
}

func featureSectionStartMarker(feature, section string) string {
	return fmt.Sprintf("// FEATURE:%s:%s:START", feature, section)
}

func featureSectionEndMarker(feature, section string) string {
	return fmt.Sprintf("// FEATURE:%s:%s:END", feature, section)
}

func insertFeatureSection(content, feature string, sec featureSection, changed bool) (string, bool, error) {
	if len(sec.lines) == 0 {
		return content, changed, nil
	}
	startMarker := featureSectionStartMarker(feature, sec.name)
	endMarker := featureSectionEndMarker(feature, sec.name)
	if strings.Contains(content, startMarker) {
		return content, changed, nil
	}
	startIdx := strings.Index(content, sec.start)
	endIdx := strings.Index(content, sec.end)
	if startIdx < 0 || endIdx < 0 || endIdx < startIdx {
		return content, changed, fmt.Errorf("missing anchor %s/%s", sec.start, sec.end)
	}
	// 插到锚点所在行的行首，而不是锚点的前导缩进之后，否则第一行会多一层缩进。
	insertAt := endIdx
	if lineStart := strings.LastIndex(content[:endIdx], "\n"); lineStart >= 0 {
		insertAt = lineStart + 1
	}
	var block strings.Builder
	block.WriteString("\t" + startMarker + "\n")
	for _, line := range sec.lines {
		block.WriteString("\t" + line + "\n")
	}
	block.WriteString("\t" + endMarker + "\n")
	return content[:insertAt] + block.String() + content[insertAt:], true, nil
}

func removeFeatureSection(content, feature, section string, changed bool) (string, bool) {
	startMarker := featureSectionStartMarker(feature, section)
	endMarker := featureSectionEndMarker(feature, section)
	startIdx := strings.Index(content, startMarker)
	if startIdx < 0 {
		return content, changed
	}
	lineStart := strings.LastIndex(content[:startIdx], "\n")
	if lineStart >= 0 {
		startIdx = lineStart + 1
	}
	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx < 0 {
		return content, changed
	}
	endIdx = startIdx + endIdx
	lineEnd := strings.Index(content[endIdx:], "\n")
	if lineEnd >= 0 {
		endIdx += lineEnd + 1
	} else {
		endIdx = len(content)
	}
	return content[:startIdx] + content[endIdx:], true
}
