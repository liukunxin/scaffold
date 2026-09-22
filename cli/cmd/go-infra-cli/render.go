package main

import (
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// renderTemplate 把模板自带的标识替换成目标项目的值，规则分两类：
//
//   - .go / .mod：模板名只出现在 go.mod 的 module 行与 import 前缀里，
//     因此按「带引号的前缀」和「module 行」精确替换成模块路径；
//   - 其它文本（.md/.mdc/.yml/...）：模板名是项目展示名，整体替换成项目名。
//
// 之所以要分两类：上一版用模块名做全局替换，会把文档标题
// `# HTTP Routing Rule（single-starter）` 变成 `（github.com/acme/myapp）`。
func renderTemplate(targetDir, templateName, moduleName, projectName string) error {
	return filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !isTextTemplateFile(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		switch strings.ToLower(filepath.Ext(path)) {
		case ".go":
			content = replaceImportPrefix(content, templateName, moduleName)
		case ".mod", ".work":
			content = replaceModModuleRefs(content, templateName, moduleName)
		default:
			content = strings.ReplaceAll(content, templateName, projectName)
		}
		if content == string(data) {
			return nil
		}
		return os.WriteFile(path, []byte(content), 0o644)
	})
}

// replaceImportPrefix 只改 .go 里的 import 前缀：`"<模板名>/...` → `"<模块名>/...`。
func replaceImportPrefix(content, templateName, moduleName string) string {
	return strings.ReplaceAll(content, `"`+templateName+`/`, `"`+moduleName+`/`)
}

// replaceModModuleRefs 处理 go.mod / go.work 里所有以模板根名开头的模块引用。
// 下面这些形态在多模块模板里都会出现，漏一个生成物就带着模板名出门：
//
//	module <name>[/sub]              → module <module>[/sub]
//	<name>[/sub] v1.2.3              → require 同仓模块（含 packages/go/*）
//	replace <name>[/sub] => ../..    → 同仓模块的相对路径替换
//	use <name>[/sub]                 → go.work
//	exclude <name> fakesum
//
// 只改「行首（或关键字之后）的第一个 token」，不做全局替换——
// 全局替换会误伤 github.com/other/<name> 这类无关模块；展示名替换走另一条分支。
func replaceModModuleRefs(content, templateName, moduleName string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if rewritten, ok := rewriteLeadingModuleRef(line, templateName, moduleName); ok {
			lines[i] = rewritten
		}
	}
	return strings.Join(lines, "\n")
}

// moduleRefKeywords 是「后一个 token 才是模块引用」的 go.mod / go.work 关键字。
var moduleRefKeywords = map[string]bool{
	"module":  true,
	"require": true,
	"replace": true,
	"use":     true,
	"exclude": true,
}

// rewriteLeadingModuleRef 把一行里第一个指向模板根模块的 token 换成目标模块路径。
func rewriteLeadingModuleRef(line, templateName, moduleName string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "//") {
		return line, false
	}
	fields := strings.Fields(trimmed)

	idx := 0
	if moduleRefKeywords[fields[0]] {
		if len(fields) < 2 {
			return line, false
		}
		idx = 1
	}
	token := fields[idx]
	if token != templateName && !strings.HasPrefix(token, templateName+"/") {
		return line, false
	}
	// 在原文里替换该 token，保留行内其余空格与注释排版。
	replacement := moduleName + strings.TrimPrefix(token, templateName)
	return leadingSpace(line) + strings.Replace(trimmed, token, replacement, 1), true
}

// formatGoSources 统一格式化落地后的 .go 文件，保证产物 gofmt-clean。
// 两处会破坏格式，且都不是模板本身能控制的：
//   - 模板 import 块的排序按模板名成立，模块名一换（single-starter → github.com/acme/x）就不成立了；
//   - feature 注入的代码片段是拼出来的，缩进需要归一化。
func formatGoSources(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || strings.ToLower(filepath.Ext(path)) != ".go" {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := format.Source(src)
		if err != nil {
			// 语法有问题就保留原样，让后续 go build 报出真实错误。
			return nil
		}
		if string(formatted) == string(src) {
			return nil
		}
		return os.WriteFile(path, formatted, 0o644)
	})
}

// legacyStarterNames 是历史上模板用过的模块根名；生成物里若还残留，
// 说明有覆盖文件没被渲染过（add/remove 后兜底修正）。
var legacyStarterNames = []string{
	singleStarterName,
	monorepoStarterName,
	"go-infra-starter",
	"go-infra-monorepo-starter",
}

func rewriteStaleStarterImports(root string) error {
	modulePath, err := readModulePath(root)
	if err != nil {
		return nil
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || strings.ToLower(filepath.Ext(path)) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		updated := content
		for _, name := range legacyStarterNames {
			updated = strings.ReplaceAll(updated, `"`+name+`/`, `"`+modulePath+`/`)
		}
		if updated == content {
			return nil
		}
		return os.WriteFile(path, []byte(updated), 0o644)
	})
}

// updateConfigYAML 把 configs/config*.yml 里的 app_name / service_name 统一成目标值。
func updateConfigYAML(configDir, appName string) error {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "config") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		path := filepath.Join(configDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := replaceAllLinesByPrefix(string(data), "app_name:", "app_name: "+appName)
		content = replaceAllLinesByPrefix(content, "service_name:", "service_name: "+appName)
		if content == string(data) {
			continue
		}
		if err = os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// dropGoModRequires 从 go.mod 中删除指定依赖的所有 require 行（含 indirect）。
// 场景裁剪后逐个 go mod edit 太慢，这里直接按行删；后续 go mod tidy 会收敛 go.sum。
func dropGoModRequires(goModPath string, modules []string) error {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		drop := false
		for _, module := range modules {
			if strings.HasPrefix(trimmed, module+" ") {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, line)
		}
	}
	if len(out) == len(lines) {
		return nil
	}
	return os.WriteFile(goModPath, []byte(strings.Join(out, "\n")), 0o644)
}

// replaceAllLinesByPrefix 替换所有匹配行；没命中就原样返回。
// 注意它无法区分「没命中」与「替换后内容恰好相同」，调用方别用返回值判存在性。
func replaceAllLinesByPrefix(content, prefix, replacement string) string {
	lines := strings.Split(content, "\n")
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), strings.TrimSpace(prefix)) {
			lines[i] = leadingSpace(line) + strings.TrimSpace(replacement)
			changed = true
		}
	}
	if !changed {
		return content
	}
	return strings.Join(lines, "\n")
}

func leadingSpace(s string) string {
	for i, ch := range s {
		if ch != ' ' && ch != '\t' {
			return s[:i]
		}
	}
	return s
}

func isTextTemplateFile(path string) bool {
	switch strings.ToLower(filepath.Ext(templateDestName(filepath.Base(path)))) {
	case ".go", ".mod", ".sum", ".work", ".yml", ".yaml", ".md", ".mdc", ".txt", ".json":
		return true
	default:
		return false
	}
}
