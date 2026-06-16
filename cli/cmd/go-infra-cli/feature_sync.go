package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type featureSpec struct {
	Imports []string
	Init    []string
	Router  []string
	Close   []string
}

var configFeatureSpecs = map[string]featureSpec{
	"mysql": {
		Imports: []string{`"github.com/liukunxin/go-infra/pkg/infra/mysql"`},
		Init: []string{
			`if cfg.Mysql.DSN != "" {`,
			`	if err = mysql.Init(cfg.Mysql); err != nil {`,
			`		return nil, err`,
			`	}`,
			`}`,
		},
		Close: []string{
			`if a.cfg.Mysql.DSN != "" {`,
			`	_ = mysql.GetClient().Close()`,
			`}`,
		},
	},
	"redis": {
		Imports: []string{`iredis "github.com/liukunxin/go-infra/pkg/infra/redis"`},
		Init: []string{
			`if len(cfg.Redis.Addresses) > 0 {`,
			`	if err = iredis.Init(&cfg.Redis); err != nil {`,
			`		return nil, err`,
			`	}`,
			`}`,
		},
		Close: []string{
			`if len(a.cfg.Redis.Addresses) > 0 {`,
			`	_ = iredis.GetClient().Close()`,
			`}`,
		},
	},
	"http-client": {
		Imports: []string{`httpclient "github.com/liukunxin/go-infra/pkg/infra/http_client"`},
		Init: []string{
			`httpclient.Init(cfg.HTTP)`,
		},
	},
	"traffic": {
		Imports: []string{
			`kitraffic "github.com/liukunxin/go-infra/pkg/infra/traffic"`,
			`"golang.org/x/time/rate"`,
		},
		Init: []string{
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
		},
	},
	"pprof": {
		Imports: []string{`"github.com/liukunxin/go-infra/pkg/infra/pprof"`},
		Init: []string{
			`pprof.Start()`,
		},
	},
	"metrics": {
		Imports: []string{
			`"github.com/liukunxin/go-infra/pkg/infra/metrics"`,
		},
		Router: []string{
			`metrics.InitMetrics(cfg.AppName, router)`,
		},
	},
}

func syncConfigFeatureArtifacts(projectDir, feature string, enable bool) (changed bool, err error) {
	spec, ok := configFeatureSpecs[feature]
	if !ok {
		return false, fmt.Errorf("unsupported feature spec: %s", feature)
	}
	bootstrapPath := filepath.Join(projectDir, "internal", "bootstrap", "app.go")
	data, err := os.ReadFile(bootstrapPath)
	if err != nil {
		return false, fmt.Errorf("read bootstrap/app.go: %w", err)
	}
	content := string(data)
	updated, changed, err := applyFeatureToBootstrap(content, feature, spec, enable)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	if err = os.WriteFile(bootstrapPath, []byte(updated), 0o644); err != nil {
		return false, fmt.Errorf("write bootstrap/app.go: %w", err)
	}
	return true, nil
}

func isConfigFeatureInstalled(projectDir, feature string) bool {
	bootstrapPath := filepath.Join(projectDir, "internal", "bootstrap", "app.go")
	data, err := os.ReadFile(bootstrapPath)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, featureSectionStartMarker(feature, "IMPORTS")) ||
		strings.Contains(content, featureSectionStartMarker(feature, "INIT")) ||
		strings.Contains(content, featureSectionStartMarker(feature, "ROUTER")) ||
		strings.Contains(content, featureSectionStartMarker(feature, "CLOSE"))
}

func syncAllConfigFeatures(projectDir string, flags map[string]bool) error {
	for feature, enabled := range flags {
		changed, err := syncConfigFeatureArtifacts(projectDir, feature, enabled)
		if err != nil {
			return err
		}
		if changed {
			state := "installed"
			if !enabled {
				state = "removed"
			}
			fmt.Printf("feature %s: %s\n", feature, state)
		}
	}
	return nil
}

func applyFeatureToBootstrap(content, feature string, spec featureSpec, enable bool) (string, bool, error) {
	type sectionSpec struct {
		name   string
		start  string
		end    string
		indent string
		lines  []string
	}
	sections := []sectionSpec{
		{name: "IMPORTS", start: "// FEATURE_IMPORTS_START", end: "// FEATURE_IMPORTS_END", indent: "\t", lines: spec.Imports},
		{name: "INIT", start: "// FEATURE_INIT_START", end: "// FEATURE_INIT_END", indent: "\t", lines: spec.Init},
		{name: "ROUTER", start: "// FEATURE_ROUTER_START", end: "// FEATURE_ROUTER_END", indent: "\t", lines: spec.Router},
		{name: "CLOSE", start: "// FEATURE_CLOSE_START", end: "// FEATURE_CLOSE_END", indent: "\t", lines: spec.Close},
	}
	changed := false
	var err error
	for _, sec := range sections {
		if len(sec.lines) == 0 {
			continue
		}
		if enable {
			content, changed, err = insertFeatureSection(content, feature, sec.name, sec.start, sec.end, sec.indent, sec.lines, changed)
			if err != nil {
				return "", false, err
			}
		} else {
			content, changed = removeFeatureSection(content, feature, sec.name, changed)
		}
	}
	return content, changed, nil
}

func featureSectionStartMarker(feature, section string) string {
	return fmt.Sprintf("// FEATURE:%s:%s:START", feature, section)
}
func featureSectionEndMarker(feature, section string) string {
	return fmt.Sprintf("// FEATURE:%s:%s:END", feature, section)
}

func insertFeatureSection(content, feature, section, anchorStart, anchorEnd, indent string, lines []string, changed bool) (string, bool, error) {
	startMarker := featureSectionStartMarker(feature, section)
	endMarker := featureSectionEndMarker(feature, section)
	if strings.Contains(content, startMarker) {
		return content, changed, nil
	}
	startIdx := strings.Index(content, anchorStart)
	endIdx := strings.Index(content, anchorEnd)
	if startIdx < 0 || endIdx < 0 || endIdx < startIdx {
		return content, changed, fmt.Errorf("bootstrap missing anchor %s/%s", anchorStart, anchorEnd)
	}
	insertPos := endIdx
	var block strings.Builder
	block.WriteString(indent + startMarker + "\n")
	for _, line := range lines {
		block.WriteString(indent + line + "\n")
	}
	block.WriteString(indent + endMarker + "\n")
	content = content[:insertPos] + block.String() + content[insertPos:]
	return content, true, nil
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
