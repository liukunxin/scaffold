package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := runInit(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "init failed: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println(version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("go-infra-cli usage:")
	fmt.Println("  go-infra-cli init <project-name> [flags]")
	fmt.Println("  go-infra-cli version")
	fmt.Println()
	fmt.Println("init flags:")
	fmt.Println("  --module        module name (default: project-name)")
	fmt.Println("  --app-name      app_name in configs/config.yml (default: project-name)")
	fmt.Println("  --output        output directory (default: current directory)")
	fmt.Println("  --template      starter template directory (default: auto-detect go-infra-starter)")
	fmt.Println("  --force         overwrite existing target directory")
	fmt.Println("  --features      comma-separated features, e.g. redis,metrics,pprof")
	fmt.Println("  --with-mysql    keep mysql integration scaffold (default: true)")
	fmt.Println("  --skip-tidy     skip running go mod tidy")
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	moduleName := fs.String("module", "", "go module name")
	appName := fs.String("app-name", "", "app_name in config")
	output := fs.String("output", ".", "output directory")
	template := fs.String("template", "", "template directory path")
	force := fs.Bool("force", false, "overwrite existing directory")
	features := fs.String("features", "", "comma-separated feature flags: redis,metrics,pprof")
	withMySQL := fs.Bool("with-mysql", true, "keep mysql integration scaffold")
	skipTidy := fs.Bool("skip-tidy", false, "skip go mod tidy")

	var projectName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		projectName = strings.TrimSpace(args[0])
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
	} else {
		if err := fs.Parse(args); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("project-name is required")
		}
		projectName = strings.TrimSpace(fs.Arg(0))
	}
	if projectName == "" {
		return errors.New("project-name cannot be empty")
	}
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	if *moduleName == "" {
		*moduleName = projectName
	}
	if *appName == "" {
		*appName = projectName
	}

	resolvedRedis, resolvedMetrics, resolvedPprof, err := parseFeaturesArg(*features)
	if err != nil {
		return err
	}

	templateDir, err := resolveTemplateDir(*template)
	if err != nil {
		return err
	}

	outputDir, err := filepath.Abs(*output)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	targetDir := filepath.Join(outputDir, projectName)

	if err = prepareTargetDir(targetDir, *force); err != nil {
		return err
	}

	if err = copyTree(templateDir, targetDir); err != nil {
		return err
	}

	if err = replaceStarterStrings(targetDir, *moduleName); err != nil {
		return err
	}
	if err = updateConfigYAML(filepath.Join(targetDir, "configs", "config.yml"), *appName, resolvedRedis, resolvedMetrics, resolvedPprof); err != nil {
		return err
	}
	if err = applyFeatureFlags(targetDir, *withMySQL); err != nil {
		return err
	}

	if !*skipTidy {
		if err = runGoModTidy(targetDir); err != nil {
			return err
		}
	}

	fmt.Printf("project generated: %s\n", targetDir)
	fmt.Printf("next steps:\n  cd %s\n  go run ./cmd/http\n", targetDir)
	return nil
}

func validateProjectName(name string) error {
	matched, err := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, name)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid project-name %q: only letters, numbers, dot, underscore, dash are allowed", name)
	}
	return nil
}

func resolveTemplateDir(explicit string) (string, error) {
	if explicit != "" {
		return validateTemplateDir(explicit)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidates := []string{
			filepath.Join(wd, "go-infra-starter"),
			filepath.Join(wd, "scaffold", "go-infra-starter"),
			filepath.Join(wd, "..", "go-infra-starter"),
		}
		for _, candidate := range candidates {
			if dirExists(candidate) {
				return candidate, nil
			}
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", errors.New("cannot find template dir; set --template explicitly")
}

func validateTemplateDir(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if !dirExists(abs) {
		return "", fmt.Errorf("template directory not found: %s", abs)
	}
	if !fileExists(filepath.Join(abs, "go.mod")) {
		return "", fmt.Errorf("template directory %s missing go.mod", abs)
	}
	return abs, nil
}

func prepareTargetDir(target string, force bool) error {
	if !fileOrDirExists(target) {
		return nil
	}
	if !force {
		return fmt.Errorf("target directory already exists: %s (use --force to overwrite)", target)
	}
	return os.RemoveAll(target)
}

func copyTree(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dstDir, 0o755)
		}
		dstPath := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	})
}

func replaceStarterStrings(targetDir, moduleName string) error {
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
		content := strings.ReplaceAll(string(data), "go-infra-starter", moduleName)
		return os.WriteFile(path, []byte(content), 0o644)
	})
}

func updateConfigYAML(path, appName string, withRedis, withMetrics, withPprof bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	content = replaceLineByPrefix(content, "app_name:", fmt.Sprintf("app_name: %s", appName))
	content = replaceLineByPrefix(content, "service_name:", fmt.Sprintf("service_name: %s", appName))
	metricsValue := "false"
	if withMetrics {
		metricsValue = "true"
	}
	redisValue := "false"
	if withRedis {
		redisValue = "true"
	}
	pprofValue := "false"
	if withPprof {
		pprofValue = "true"
	}
	content = replaceLineByPrefix(content, "metrics:", "metrics: "+metricsValue)
	content = replaceLineByPrefix(content, "redis:", "redis: "+redisValue)
	content = replaceLineByPrefix(content, "pprof:", "pprof: "+pprofValue)
	return os.WriteFile(path, []byte(content), 0o644)
}

func parseFeaturesArg(features string) (bool, bool, bool, error) {
	// 默认值与模板保持一致。
	enabled := map[string]bool{
		"redis":   false,
		"metrics": true,
		"pprof":   false,
	}
	if strings.TrimSpace(features) == "" {
		return enabled["redis"], enabled["metrics"], enabled["pprof"], nil
	}

	// 一旦显式指定 --features，则以显式列表为准。
	enabled["redis"], enabled["metrics"], enabled["pprof"] = false, false, false
	for _, item := range strings.Split(features, ",") {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		if _, ok := enabled[key]; !ok {
			return false, false, false, fmt.Errorf("unsupported feature %q, allowed: redis,metrics,pprof", key)
		}
		enabled[key] = true
	}
	return enabled["redis"], enabled["metrics"], enabled["pprof"], nil
}

func applyFeatureFlags(targetDir string, withMySQL bool) error {
	if withMySQL {
		return nil
	}
	configPath := filepath.Join(targetDir, "configs", "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	content := replaceLineByPrefix(string(data), "  dsn:", `  dsn: ""`)
	return os.WriteFile(configPath, []byte(content), 0o644)
}

func runGoModTidy(targetDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run go mod tidy: %w", err)
	}
	return nil
}

func replaceLineByPrefix(content, prefix, replacement string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), strings.TrimSpace(prefix)) {
			indent := leadingSpace(line)
			lines[i] = indent + strings.TrimSpace(replacement)
			return strings.Join(lines, "\n")
		}
	}
	return content
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
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".mod", ".sum", ".yml", ".yaml", ".md", ".txt", ".json":
		return true
	default:
		return false
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func fileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

