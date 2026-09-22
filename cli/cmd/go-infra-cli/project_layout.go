package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 文件定位与 feature 锚点集中在这里，add/remove/init 与布局校验共用同一份定义。
const (
	bootstrapFilePath = "internal/bootstrap/app.go"
	routeFilePath     = "internal/route/init.go"
)

const (
	anchorImportsStart = "// FEATURE_IMPORTS_START"
	anchorImportsEnd   = "// FEATURE_IMPORTS_END"
	anchorInitStart    = "// FEATURE_INIT_START"
	anchorInitEnd      = "// FEATURE_INIT_END"
	anchorRouterStart  = "// FEATURE_ROUTER_START"
	anchorRouterEnd    = "// FEATURE_ROUTER_END"
	anchorCloseStart   = "// FEATURE_CLOSE_START"
	anchorCloseEnd     = "// FEATURE_CLOSE_END"
	anchorRoutesStart  = "// FEATURE_ROUTES_START"
	anchorRoutesEnd    = "// FEATURE_ROUTES_END"
)

// starterAnchors 是 add/remove 依赖的最小骨架锚点。
// 只校验关键文件，不绑业务模块路径——用户删掉 internal/app/user 之后 add redis 仍应可用。
var starterAnchors = []string{
	"go.mod",
	"configs/config.yml",
	bootstrapFilePath,
}

// anchorGroup 声明注入能力所依赖的锚点，缺任一即无法安全注入。
type anchorGroup struct {
	file    string
	anchors []string
}

var requiredFeatureAnchors = []anchorGroup{
	{
		file:    bootstrapFilePath,
		anchors: []string{anchorImportsStart, anchorInitStart, anchorRouterStart, anchorCloseStart},
	},
	{
		file:    routeFilePath,
		anchors: []string{anchorRoutesStart},
	},
}

func validateStarterLayout(projectDir string) error {
	var missing []string
	for _, rel := range starterAnchors {
		if !fileExists(filepath.Join(projectDir, filepath.FromSlash(rel))) {
			missing = append(missing, rel)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"not a go-infra-cli project (missing: %s); run `go-infra-cli init` first or pass --dir",
			strings.Join(missing, ", "),
		)
	}

	for _, group := range requiredFeatureAnchors {
		path := filepath.Join(projectDir, filepath.FromSlash(group.file))
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", group.file, err)
		}
		content := string(data)
		for _, anchor := range group.anchors {
			if !strings.Contains(content, anchor) {
				return fmt.Errorf("%s is missing feature anchor %q; regenerate the project with the current CLI", group.file, anchor)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "configs", "config.yml"))
	if err != nil {
		return fmt.Errorf("read configs/config.yml: %w", err)
	}
	if strings.Contains(string(data), "\nfeatures:") || strings.HasPrefix(string(data), "features:") {
		return fmt.Errorf("configs/config.yml still has legacy features: section; remove it or regenerate the project")
	}
	return nil
}
