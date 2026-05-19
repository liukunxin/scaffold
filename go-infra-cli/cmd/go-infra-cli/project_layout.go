package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var starterLayoutPaths = []string{
	"cmd/http/main.go",
	"internal/bootstrap/app.go",
	"internal/route/init.go",
	"internal/infra/config/app.go",
	"internal/infra/config/loader.go",
	"internal/app/user/controller/controller.go",
	"internal/app/user/service/service.go",
	"internal/app/user/dao/repo.go",
	"internal/app/user/dto/user.go",
	"internal/app/user/ro/user.go",
	"internal/app/user/vo/user.go",
	"internal/app/user/convert/user.go",
	"internal/app/user/codes/user.go",
	"internal/app/demo/controller/controller.go",
	"internal/app/demo/service/service.go",
}

func validateStarterLayout(projectDir string) error {
	var missingPaths []string
	for _, rel := range starterLayoutPaths {
		if !fileExists(filepath.Join(projectDir, filepath.FromSlash(rel))) {
			missingPaths = append(missingPaths, rel)
		}
	}
	if len(missingPaths) > 0 {
		return fmt.Errorf(
			"project layout does not match go-infra-starter (missing: %s); add/remove only apply to standard scaffold projects",
			strings.Join(missingPaths, ", "),
		)
	}

	configPath := filepath.Join(projectDir, "configs", "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read configs/config.yml: %w", err)
	}
	if strings.Contains(string(data), "\nfeatures:") || strings.HasPrefix(string(data), "features:") {
		return fmt.Errorf("configs/config.yml still has legacy features: section; remove it or regenerate the project")
	}
	return nil
}
