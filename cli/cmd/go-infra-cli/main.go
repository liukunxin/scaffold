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
	fmt.Println("  go-infra-cli mono add app|service <name> [--dir <monorepo-root>]")
	fmt.Println("  go-infra-cli keygen")
	fmt.Println("  go-infra-cli encrypt --value=<plaintext> [--key=<hex>|--key-env=<ENV>]")
	fmt.Println("  go-infra-cli decrypt --value=<ENC(...)> [--key=<hex>|--key-env=<ENV>]")
	fmt.Println("  go-infra-cli version")
	fmt.Println()
	fmt.Println("common flags:")
	fmt.Println("  init: --layout single|monorepo --module --app-name --features --scenes --output --force --skip-tidy --use-local-sdk")
	fmt.Println("  add/remove/mono add: --dir")
	fmt.Println("  features: mysql,redis,metrics,pprof,http-client,traffic,llm")
	fmt.Println("            (llm ships extra files: add/remove installs or reclaims them too)")
	fmt.Println("  scenes: http,grpc,ws (default: all; http is always kept)")
	fmt.Println("  encrypt/decrypt: --key (hex) or --key-env (default: CONFIG_ENCRYPT_KEY)")
	fmt.Println()
	fmt.Println("layouts:")
	fmt.Println("  single   -> single-starter: one Go project (cmd/http + cmd/grpc + internal/{app,bootstrap,infra,route})")
	fmt.Println("  monorepo -> monorepo-starter: apps/ services/ packages/ contracts/ tools/ deploy/ docs/")
	fmt.Println("              Go projects under apps|services follow the single-starter layout;")
	fmt.Println("              add more with `mono add app|service`.")
}

func runInit(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	moduleName := flags.String("module", "", "go module name")
	appName := flags.String("app-name", "", "app_name in config")
	layout := flags.String("layout", "single", "project layout: single|monorepo")
	output := flags.String("output", ".", "output directory")
	template := flags.String("template", "", "override template directory (default: embedded templates)")
	force := flags.Bool("force", false, "overwrite existing directory")
	features := flags.String("features", "", "comma-separated install list: mysql,redis,metrics,pprof,http-client,traffic,llm")
	scenes := flags.String("scenes", "", "comma-separated runtime scenes: http,grpc,ws (default: all)")
	skipTidy := flags.Bool("skip-tidy", false, "skip go mod tidy and the generated-project build check")
	useLocalSDK := flags.Bool("use-local-sdk", false, "point go-infra at a nearby local checkout (dev only)")

	var projectName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		projectName = strings.TrimSpace(args[0])
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
	} else {
		if err := flags.Parse(args); err != nil {
			return err
		}
		if flags.NArg() < 1 {
			return errors.New("project-name is required")
		}
		projectName = strings.TrimSpace(flags.Arg(0))
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

	rawFeatures := strings.TrimSpace(*features)
	rawScenes := strings.TrimSpace(*scenes)
	if layoutName == "monorepo" {
		// monorepo 是「各 Project 一个 module + go.work」的布局，没有单一
		// internal/bootstrap/app.go 可注入；能力增删针对具体 Project 执行。
		// --scenes 同理只服务 single：mono add 没有场景选择，Project 恒为全场景。
		if rawFeatures != "" {
			return errors.New("--features is not supported for --layout monorepo: run `go-infra-cli add <features> --dir services/<name>` on the target project instead")
		}
		if rawScenes != "" {
			return errors.New("--scenes is not supported for --layout monorepo: scene trimming is only available for --layout single; monorepo projects are always generated with all scenes (http,grpc,ws)")
		}
	}

	configFlags, resolvedLLM, err := parseInitFeaturesArg(rawFeatures)
	if err != nil {
		return err
	}
	sceneFlags, err := parseInitScenesArg(rawScenes)
	if err != nil {
		return err
	}

	src, err := resolveTemplate(*template, layoutName)
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

	if err = copyTree(src, targetDir); err != nil {
		return err
	}

	if layoutName == "single" {
		// llm 覆盖文件里含模板模块路径，必须在渲染前拷进来。
		if err = applyLLMOverlay(targetDir, resolvedLLM); err != nil {
			return err
		}
		if err = applySceneSelection(targetDir, sceneFlags); err != nil {
			return err
		}
	}

	// 渲染分两步：模块路径（go.mod/import）与展示名（文档/config）分开处理。
	if err = renderTemplate(targetDir, src.name, *moduleName, projectName); err != nil {
		return err
	}

	if layoutName == "single" {
		// monorepo 下每个 Project 有自己的 configs/，--app-name 映射不到唯一目标，
		// 因此这条只对 single 生效（monorepo 的展示名已由 renderTemplate 处理）。
		if err = updateConfigYAML(filepath.Join(targetDir, "configs"), *appName); err != nil {
			return err
		}
		if err = syncAllFeatures(targetDir, configFlags, resolvedLLM); err != nil {
			return err
		}
	}

	if err = formatGoSources(targetDir); err != nil {
		return err
	}

	if *useLocalSDK {
		if err = ensureLocalGoInfraReplace(targetDir); err != nil {
			return err
		}
	}

	if *skipTidy {
		fmt.Println("skip-tidy: skipped go mod tidy and build verification; run 'go mod tidy && go build ./...' yourself")
	} else {
		if err = runGoModTidy(targetDir); err != nil {
			return rollbackTarget(targetDir, err)
		}
		if err = verifyGeneratedProject(targetDir); err != nil {
			return rollbackTarget(targetDir, err)
		}
	}

	fmt.Printf("project generated: %s\n", targetDir)
	fmt.Println("next steps:")
	fmt.Printf("  cd %s\n", targetDir)
	if layoutName == "single" {
		fmt.Println("  go run ./cmd/http")
		if sceneFlags["grpc"] {
			fmt.Println("  go run ./cmd/grpc")
		}
		// 启用了「不填参数就用不起来」的能力时提醒一句：否则用户会以为装完就能跑。
		if needConfig := configHintList(configFlags, resolvedLLM); len(needConfig) > 0 {
			fmt.Printf("  fill configs/config.yml to enable: %s\n", strings.Join(needConfig, ", "))
		}
	} else {
		fmt.Println("  cd services/gateway && go run ./cmd/http   # example Project")
		fmt.Println("  go-infra-cli mono add service <name>        # add another Go Project")
		fmt.Println("  make check                                  # fmt + tidy + build + test")
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

func prepareTargetDir(target string, force bool) error {
	if !fileOrDirExists(target) {
		return nil
	}
	if !force {
		return fmt.Errorf("target directory already exists: %s (use --force to overwrite)", target)
	}
	return os.RemoveAll(target)
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

// runGoModTidy 只做依赖整理，不再改写 go.mod 的 replace。
// 模板本身 require 发布版 go-infra，任何机器都能解析；本地 SDK 需显式 --use-local-sdk。
func runGoModTidy(targetDir string) error {
	if fileExists(filepath.Join(targetDir, "go.work")) {
		if err := runGo(targetDir, "work", "sync"); err != nil {
			return fmt.Errorf("run go work sync: %w", err)
		}
	}

	modDirs, err := collectModuleDirs(targetDir)
	if err != nil {
		return err
	}
	if len(modDirs) == 0 {
		return fmt.Errorf("no go.mod found under %s", targetDir)
	}
	for _, dir := range modDirs {
		if err = runGo(dir, "mod", "tidy"); err != nil {
			return fmt.Errorf("run go mod tidy in %s: %w", dir, err)
		}
	}
	return nil
}

// verifyGeneratedProject 自检生成物能编译。模板或渲染链路一坏就地失败，
// 不留「生成即坏」的项目——这正是之前 replace 写法埋下的坑。
func verifyGeneratedProject(targetDir string) error {
	modDirs, err := collectModuleDirs(targetDir)
	if err != nil {
		return err
	}
	for _, dir := range modDirs {
		if err = runGo(dir, "build", "./..."); err != nil {
			return fmt.Errorf("generated project does not build (%s): %w", dir, err)
		}
	}
	fmt.Println("verify: go build ./... ok")
	return nil
}

func rollbackTarget(targetDir string, cause error) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("%w (also failed to remove %s: %v)", cause, targetDir, err)
	}
	return fmt.Errorf("%w (broken output removed: %s)", cause, targetDir)
}

func runGo(dir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
