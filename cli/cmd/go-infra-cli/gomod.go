package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const goInfraModulePath = "github.com/liukunxin/go-infra"

func readModulePath(projectDir string) (string, error) {
	return readModulePathFromFile(filepath.Join(projectDir, "go.mod"))
}

func readModulePathFromFile(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found in %s", goModPath)
}

// findLocalGoInfraDir walks upward from start looking for go-infra/go.mod
// whose module path is github.com/liukunxin/go-infra.
func findLocalGoInfraDir(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		candidate := filepath.Join(dir, "go-infra")
		modFile := filepath.Join(candidate, "go.mod")
		if fileExists(modFile) {
			data, err := os.ReadFile(modFile)
			if err == nil && strings.Contains(string(data), "module "+goInfraModulePath) {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", nil
}

// ensureLocalGoInfraReplace 只在显式 --use-local-sdk 时调用：
// 把 go-infra 指到旁边的本地 checkout，方便 SDK 未发版时联调。
// 默认不调用——模板 require 的是发布版，任何机器 clone 下来都能解析依赖。
func ensureLocalGoInfraReplace(moduleDir string) error {
	goInfraDir, err := findLocalGoInfraDir(moduleDir)
	if err != nil {
		return err
	}
	if goInfraDir == "" {
		fmt.Println("--use-local-sdk: no local go-infra checkout found nearby, keeping the released dependency")
		return nil
	}
	rel, err := filepath.Rel(moduleDir, goInfraDir)
	if err != nil {
		return fmt.Errorf("rel path to go-infra: %w", err)
	}
	rel = filepath.ToSlash(rel)

	drop := exec.Command("go", "mod", "edit", "-dropreplace="+goInfraModulePath)
	drop.Dir = moduleDir
	_ = drop.Run() // ignore: may not exist yet

	req := exec.Command("go", "mod", "edit", "-require="+goInfraModulePath+"@v0.0.0")
	req.Dir = moduleDir
	req.Stdout = os.Stdout
	req.Stderr = os.Stderr
	if err := req.Run(); err != nil {
		return fmt.Errorf("go mod edit -require go-infra: %w", err)
	}

	rep := exec.Command("go", "mod", "edit", "-replace="+goInfraModulePath+"="+rel)
	rep.Dir = moduleDir
	rep.Stdout = os.Stdout
	rep.Stderr = os.Stderr
	if err := rep.Run(); err != nil {
		return fmt.Errorf("go mod edit -replace go-infra: %w", err)
	}
	fmt.Printf("go-infra: using local module via replace => %s\n", rel)
	return nil
}
