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
	data, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
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

// ensureLocalGoInfraReplace points the module at a nearby local go-infra checkout
// so generated projects use the latest SDK sources instead of a stale proxy version.
// No-op when no local go-infra is found (keeps go.mod require as-is for tidy).
func ensureLocalGoInfraReplace(moduleDir string) error {
	goInfraDir, err := findLocalGoInfraDir(moduleDir)
	if err != nil {
		return err
	}
	if goInfraDir == "" {
		fmt.Println("go-infra: local checkout not found; keeping go.mod require (run go get after publishing a new version)")
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
