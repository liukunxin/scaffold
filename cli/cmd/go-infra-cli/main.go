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
	case "mono":
		if err := runMono(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "mono failed: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if err := runAdd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "add failed: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if err := runRemove(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "remove failed: %v\n", err)
			os.Exit(1)
		}
	case "keygen":
		if err := runKeygen(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "keygen failed: %v\n", err)
			os.Exit(1)
		}
	case "encrypt":
		if err := runEncrypt(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "encrypt failed: %v\n", err)
			os.Exit(1)
		}
	case "decrypt":
		if err := runDecrypt(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "decrypt failed: %v\n", err)
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
	fmt.Println("  go-infra-cli add <features> [flags]")
	fmt.Println("  go-infra-cli remove <features> [flags]")
	fmt.Println("  go-infra-cli mono add app <name> [--dir <project-root>]")
	fmt.Println("  go-infra-cli mono add domain <name> [--dir <project-root>]")
	fmt.Println("  go-infra-cli keygen")
	fmt.Println("  go-infra-cli encrypt --value=<plaintext> [--key=<hex>|--key-env=<ENV>]")
	fmt.Println("  go-infra-cli decrypt --value=<ENC(...)> [--key=<hex>|--key-env=<ENV>]")
	fmt.Println("  go-infra-cli version")
	fmt.Println()
	fmt.Println("common flags:")
	fmt.Println("  init: --layout single|monorepo --module --app-name --features --scenes --output --force --skip-tidy")
	fmt.Println("  add/remove: --dir")
	fmt.Println("  features: mysql,redis,metrics,pprof,http-client,traffic (llm via init --features llm)")
	fmt.Println("  encrypt/decrypt: --key (hex) or --key-env (default: CONFIG_ENCRYPT_KEY)")
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	moduleName := fs.String("module", "", "go module name")
	appName := fs.String("app-name", "", "app_name in config")
	layout := fs.String("layout", "single", "project layout: single|monorepo")
	output := fs.String("output", ".", "output directory")
	template := fs.String("template", "", "template directory path")
	force := fs.Bool("force", false, "overwrite existing directory")
	features := fs.String("features", "", "comma-separated install list: mysql,redis,metrics,pprof,http-client,traffic,llm")
	scenes := fs.String("scenes", "", "comma-separated runtime scenes: http,grpc,ws")
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

	layoutName := strings.ToLower(strings.TrimSpace(*layout))
	if layoutName == "" {
		layoutName = "single"
	}
	if layoutName != "single" && layoutName != "monorepo" {
		return fmt.Errorf("unsupported layout %q: only single|monorepo are allowed", *layout)
	}

	configFlags, resolvedLLM, err := parseInitFeaturesArg(*features)
	if err != nil {
		return err
	}
	sceneFlags, err := parseInitScenesArg(*scenes)
	if err != nil {
		return err
	}

	templateDir, err := resolveTemplateDir(*template, layoutName)
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
	if layoutName == "single" {
		if err = applyLLMOverlay(targetDir, resolvedLLM); err != nil {
			return err
		}
		if err = syncAllConfigFeatures(targetDir, configFlags); err != nil {
			return err
		}
		if err = applySceneSelection(targetDir, sceneFlags); err != nil {
			return err
		}
	} else {
		// Monorepo starter currently has fixed runtime skeleton and does not use
		// single-repo feature/scenes mutation logic.
		if strings.TrimSpace(*features) != "" {
			return errors.New("--features is not supported for --layout monorepo in MVP")
		}
		if strings.TrimSpace(*scenes) != "" {
			return errors.New("--scenes is not supported for --layout monorepo in MVP")
		}
	}

	if err = replaceStarterStrings(targetDir, *moduleName); err != nil {
		return err
	}
	if err = updateConfigYAML(filepath.Join(targetDir, "configs", "config.yml"), *appName); err != nil {
		return err
	}
	if layoutName == "single" {
		if err = applyInitMySQLDSN(targetDir, configFlags["mysql"]); err != nil {
			return err
		}
	}

	if !*skipTidy {
		if err = runGoModTidy(targetDir); err != nil {
			return err
		}
	}

	fmt.Printf("project generated: %s\n", targetDir)
	fmt.Printf("next steps:\n  cd %s\n", targetDir)
	if layoutName == "single" {
		fmt.Println("  go run ./cmd/http")
		if sceneFlags["grpc"] {
			fmt.Println("  go run ./cmd/grpc")
		}
	} else {
		fmt.Println("  go run ./apps/gateway/cmd")
	}
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

func resolveTemplateDir(explicit, layout string) (string, error) {
	if explicit != "" {
		return validateTemplateDir(explicit)
	}
	templateName := "single-starter"
	if layout == "monorepo" {
		templateName = "monorepo-starter"
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidates := []string{
			filepath.Join(wd, templateName),
			filepath.Join(wd, "scaffold", templateName),
			filepath.Join(wd, "..", templateName),
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
	if !fileExists(filepath.Join(abs, "go.mod")) && !fileExists(filepath.Join(abs, "go.work")) {
		return "", fmt.Errorf("template directory %s missing go.mod/go.work", abs)
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
		if rel == "_features" || strings.HasPrefix(rel, "_features"+string(filepath.Separator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
		content := string(data)
		content = strings.ReplaceAll(content, "single-starter", moduleName)
		content = strings.ReplaceAll(content, "monorepo-starter", moduleName)
		// Legacy placeholders still replaced for historical templates.
		content = strings.ReplaceAll(content, "go-infra-starter", moduleName)
		content = strings.ReplaceAll(content, "go-infra-monorepo-starter", moduleName)
		return os.WriteFile(path, []byte(content), 0o644)
	})
}

func updateConfigYAML(path, appName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	content = replaceLineByPrefix(content, "app_name:", fmt.Sprintf("app_name: %s", appName))
	content = replaceLineByPrefix(content, "service_name:", fmt.Sprintf("service_name: %s", appName))
	return os.WriteFile(path, []byte(content), 0o644)
}

func applyLLMOverlay(targetDir string, withLLM bool) error {
	if !withLLM {
		return nil
	}
	if err := copyEmbeddedLLMOverlay(targetDir); err != nil {
		return fmt.Errorf("apply llm feature overlay: %w", err)
	}
	return nil
}

func applyInitMySQLDSN(targetDir string, withMySQL bool) error {
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

func parseInitScenesArg(raw string) (map[string]bool, error) {
	// Default scene is always HTTP.
	scenes := map[string]bool{
		"http": true,
		"grpc": false,
		"ws":   false,
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return scenes, nil
	}
	for _, part := range strings.Split(raw, ",") {
		scene := strings.TrimSpace(strings.ToLower(part))
		if scene == "" {
			continue
		}
		switch scene {
		case "http", "grpc", "ws":
			scenes[scene] = true
		default:
			return nil, fmt.Errorf("unsupported scene %q: only http,grpc,ws are allowed", scene)
		}
	}
	// ws runs on http upgrade, so http must exist.
	if scenes["ws"] {
		scenes["http"] = true
	}
	return scenes, nil
}

func applySceneSelection(projectDir string, scenes map[string]bool) error {
	if !scenes["ws"] {
		_ = os.RemoveAll(filepath.Join(projectDir, "internal", "app", "realtime"))
	}

	if scenes["grpc"] {
		return keepSceneMarkers(filepath.Join(projectDir, "internal", "route", "init.go"), "SCENE_WS", scenes["ws"])
	}

	// grpc scene disabled: remove grpc demo command and bootstrap wiring.
	_ = os.RemoveAll(filepath.Join(projectDir, "cmd", "grpc"))
	_ = os.Remove(filepath.Join(projectDir, "internal", "bootstrap", "grpc.go"))
	_ = os.RemoveAll(filepath.Join(projectDir, "internal", "app", "demo", "grpc"))

	return keepSceneMarkers(filepath.Join(projectDir, "internal", "route", "init.go"), "SCENE_WS", scenes["ws"])
}

func keepSceneMarkers(filePath, sceneKey string, keep bool) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(data)
	startMarker := "// " + sceneKey + "_START"
	endMarker := "// " + sceneKey + "_END"

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == startMarker {
			inBlock = true
			if keep {
				continue
			}
			continue
		}
		if trimmed == endMarker {
			inBlock = false
			continue
		}
		if inBlock && !keep {
			continue
		}
		out = append(out, line)
	}

	return os.WriteFile(filePath, []byte(strings.Join(out, "\n")), 0o644)
}

func runGoModTidy(targetDir string) error {
	// Monorepo: sync workspace then tidy each app module.
	if fileExists(filepath.Join(targetDir, "go.work")) {
		syncCmd := exec.Command("go", "work", "sync")
		syncCmd.Dir = targetDir
		syncCmd.Stdout = os.Stdout
		syncCmd.Stderr = os.Stderr
		if err := syncCmd.Run(); err != nil {
			return fmt.Errorf("run go work sync: %w", err)
		}

		modDirs, err := collectModuleDirs(targetDir)
		if err != nil {
			return err
		}
		for _, dir := range modDirs {
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("run go mod tidy in %s: %w", dir, err)
			}
		}
		return nil
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run go mod tidy: %w", err)
	}
	return nil
}

func collectModuleDirs(root string) ([]string, error) {
	out := make([]string, 0, 4)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			out = append(out, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover go modules: %w", err)
	}
	return out, nil
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
