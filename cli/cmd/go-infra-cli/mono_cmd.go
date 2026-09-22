package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// monoTargets 把 mono add 的目标映射到分类目录。
//
// app 与 service 生成的都是 single-starter 规范的 Go Project ——
// 区别只在服务对象（面向用户 vs 后端能力），结构完全一致。
var monoTargets = map[string]string{
	"app":     "apps",
	"service": "services",
}

// 布局名。resolveTemplate 只认这两个值：传错（例如 "mono"）不会报错，
// 而是静默退回 single-starter —— 是个很难查的坑，所以这里不写裸字符串。
const (
	layoutSingle   = "single"
	layoutMonorepo = "monorepo"
)

const monoUsage = "usage: go-infra-cli mono add app|service <name> [--dir <monorepo-root>] [--force]"

// 移出 Project 的文件：这些由 Monorepo 根统一提供，Project 里不再重复一份。
//
// .dockerignore / .gitattributes 也在内：前者只有**构建上下文根**那份才被 docker 读到；
// 后者虽然能作用于子目录，但仓库根那份已经覆盖到，Project 里再来一份只是噪音。
var monorepoOwnedFiles = []string{".cursor", "AGENTS.md", ".gitignore", ".dockerignore", ".gitattributes"}

func runMono(args []string) error {
	if len(args) < 1 {
		return errors.New(monoUsage)
	}
	switch args[0] {
	case "add":
		return runMonoAdd(args[1:])
	default:
		return fmt.Errorf("unsupported mono subcommand %q", args[0])
	}
}

func runMonoAdd(args []string) error {
	fs := flag.NewFlagSet("mono add", flag.ContinueOnError)
	projectDir := fs.String("dir", "", "monorepo root directory (default: auto-detect from current directory)")
	force := fs.Bool("force", false, "overwrite the project directory if it exists")

	if len(args) < 2 {
		return errors.New(monoUsage)
	}
	kind := strings.TrimSpace(strings.ToLower(args[0]))
	name := strings.TrimSpace(args[1])
	if err := validateMonoName(name); err != nil {
		return err
	}
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	parent, ok := monoTargets[kind]
	if !ok {
		return fmt.Errorf("unsupported mono add target %q: only app|service are allowed", kind)
	}

	root, err := resolveMonorepoDir(*projectDir)
	if err != nil {
		return err
	}
	moduleRoot, err := readMonorepoModuleRoot(root)
	if err != nil {
		return err
	}

	target := filepath.Join(root, parent, name)
	if err = prepareTargetDir(target, *force); err != nil {
		return err
	}
	if err = scaffoldMonoProject(target, parent, name, moduleRoot); err != nil {
		return rollbackTarget(target, err)
	}

	// go.work 一写进 use 行就立刻影响整个 workspace：回滚时只删 Project 目录而不撤销这行，
	// 会留下一个指向不存在目录的 use —— 之后任何 go build / go mod tidy 都直接失败，
	// 且报错指向 go.work，用户很难联想到是"上一次 mono add 没干净退出"。所以先留底。
	workPath := filepath.Join(root, "go.work")
	workBackup, readErr := os.ReadFile(workPath)
	if readErr != nil {
		return rollbackTarget(target, readErr)
	}
	rollback := func(cause error) error {
		if werr := os.WriteFile(workPath, workBackup, 0o644); werr != nil {
			return fmt.Errorf("%w (also failed to restore go.work: %v)", cause, werr)
		}
		return rollbackTarget(target, cause)
	}
	if err = appendGoWorkUse(root, "./"+parent+"/"+name); err != nil {
		return rollback(err)
	}

	// 新 Project 自带 go.mod，依赖与 go.sum 都还没收敛；与 single 的 init 对齐，生成完就自检。
	if err = runGoModTidy(root); err != nil {
		return rollback(err)
	}
	if err = verifyGeneratedProject(root); err != nil {
		return rollback(err)
	}

	fmt.Printf("%s project added: %s/%s\n", kind, parent, name)
	fmt.Printf("  module: %s/%s/%s\n", moduleRoot, parent, name)
	fmt.Println("  go.work updated; go-infra-cli add/remove needs --dir " + filepath.ToSlash(filepath.Join(parent, name)))
	return nil
}

// scaffoldMonoProject 按 single-starter 模板生成一个 Project 骨架。
//
// Project 的内部架构规范只有一个来源：single-starter。
// 所以这里不做任何"重新发明"，只做两件事：去掉独立项目专属的文件、把模块路径改对。
func scaffoldMonoProject(target, parent, name, moduleRoot string) error {
	src, err := resolveTemplate("", layoutSingle)
	if err != nil {
		return err
	}
	if err = copyTree(src, target); err != nil {
		return err
	}
	for _, rel := range monorepoOwnedFiles {
		if err = os.RemoveAll(filepath.Join(target, rel)); err != nil {
			return err
		}
	}
	if err = renderMonoProjectDockerfile(target, parent, name); err != nil {
		return err
	}

	moduleName := moduleRoot + "/" + parent + "/" + name
	if err = renderTemplate(target, src.name, moduleName, name); err != nil {
		return err
	}
	if err = formatGoSources(target); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(target, "README.md"), []byte(projectReadme(name, parent, moduleName)), 0o644)
}

// monoProjectDockerfileSource 是 Monorepo 内 Go Project 的 Dockerfile 真源。
//
// 为什么不直接沿用 single-starter 拷过来的那份：两者的**构建上下文**不同——
// 单服务项目以项目根为上下文，Monorepo 里的 Project 必须以仓库根为上下文，
// 否则容器里看不见 go.work 与 packages/，引用了同仓模块的 Project 直接构建失败。
const monoProjectDockerfileSource = "services/gateway/Dockerfile"

// renderMonoProjectDockerfile 以 gateway 的 Dockerfile 为模板生成新 Project 的，
// 只把「模板服务所在目录」替换成新 Project 的路径，其余（入口、端口、env）保持一致。
func renderMonoProjectDockerfile(target, parent, name string) error {
	src, err := resolveTemplate("", layoutMonorepo)
	if err != nil {
		return err
	}
	rel := path.Join(src.root, monoProjectDockerfileSource)
	data, err := fs.ReadFile(src.fsys, rel)
	if err != nil {
		return fmt.Errorf("read monorepo dockerfile template %s: %w", rel, err)
	}
	content := strings.ReplaceAll(string(data), "services/gateway", parent+"/"+name)
	return os.WriteFile(filepath.Join(target, "Dockerfile"), []byte(content), 0o644)
}

func validateMonoName(name string) error {
	ok, err := regexp.MatchString(`^[a-z][a-z0-9-]*$`, name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid name %q: must match ^[a-z][a-z0-9-]*$", name)
	}
	return nil
}

func resolveMonorepoDir(explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		if err := validateMonorepoRoot(abs); err != nil {
			return "", err
		}
		return abs, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if err := validateMonorepoRoot(wd); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", errors.New("cannot find monorepo root (need go.work + apps + services + packages + contracts); use --dir")
}

// validateMonorepoRoot 只校验 Repository 层的骨架目录，不绑具体 Project：
// 一个 PROJECT 都没建、或者某个分类目录暂时为空，都不该让 mono add 失败。
func validateMonorepoRoot(dir string) error {
	required := []string{
		"go.work",
		"apps",
		"services",
		"packages",
		"contracts",
	}
	for _, rel := range required {
		if !fileOrDirExists(filepath.Join(dir, filepath.FromSlash(rel))) {
			return fmt.Errorf("invalid monorepo root: missing %s", rel)
		}
	}
	return nil
}

// readMonorepoModuleRoot 推导 Monorepo 的 Go 模块根（例如 github.com/acme/mono）。
//
// 仓库根目录**没有** go.mod——它是纯 go.work 的 workspace，每个 Project 与
// packages/go/* 各自是独立模块。因此模块根只能从已有模块反推：
// 取任一 <category>/<name>/go.mod 的 module 行，去掉 "/<category>/<name>" 后缀。
func readMonorepoModuleRoot(root string) (string, error) {
	patterns := []string{
		filepath.Join(root, "apps", "*", "go.mod"),
		filepath.Join(root, "services", "*", "go.mod"),
		filepath.Join(root, "packages", "go", "*", "go.mod"),
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, goModPath := range matches {
			modulePath, err := readModulePathFromFile(goModPath)
			if err != nil {
				continue
			}
			rel, err := filepath.Rel(root, filepath.Dir(goModPath))
			if err != nil {
				continue
			}
			if suffix := "/" + filepath.ToSlash(rel); strings.HasSuffix(modulePath, suffix) {
				return strings.TrimSuffix(modulePath, suffix), nil
			}
		}
	}
	return "", errors.New(
		"cannot infer the monorepo Go module root: no Go module found under apps/*, services/* or packages/go/*; " +
			"create one first (e.g. go-infra-cli mono add service <name>)",
	)
}

func projectReadme(name, parent, moduleName string) string {
	kind := "Go Project"
	switch parent {
	case "apps":
		kind = "面向用户/客户端的 Go Project"
	case "services":
		kind = "后端服务 Go Project"
	}
	return fmt.Sprintf(`# %s

%s，按 `+"`single-starter`"+` 规范组织。

- module: %s
- 分类目录: %s/（见仓库根 README 的「目录职责」）

## 目录

	%s/%s/
	├── cmd/                 # 按运行形态划分（http / grpc / ...）
	├── configs/
	├── internal/
	│   ├── app/             # 按业务域垂直切片
	│   ├── bootstrap/       # 启动编排
	│   ├── infra/           # 技术适配
	│   └── route/           # 协议层
	├── Dockerfile           # 构建上下文 = 仓库根，见文件内注释
	├── Makefile
	├── go.mod
	└── go.sum

## 运行

	make run          # HTTP
	make run-grpc     # gRPC
	make test

## 能力增删

	go-infra-cli add mysql,redis,llm --dir %s/%s
	go-infra-cli remove redis --dir %s/%s

## 约束

- 禁止 import 其它 Project 的 internal/；跨 Project 走 contracts/ + HTTP / gRPC / Event。
- 能力基线统一复用 github.com/liukunxin/go-infra，不重复造轮子。
- 详细分层与依赖规则见仓库根 .cursor/rules/00-architecture.mdc。
`, name, kind, moduleName, parent, parent, name, parent, name, parent, name)
}

func appendGoWorkUse(root, usePath string) error {
	path := filepath.Join(root, "go.work")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "\n\t"+usePath+"\n") || strings.Contains(content, "\n\t"+usePath+"\r\n") {
		return nil
	}

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+1)
	inUseBlock := false
	inserted := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "use (") {
			inUseBlock = true
			out = append(out, line)
			continue
		}
		if inUseBlock && trimmed == ")" && !inserted {
			out = append(out, "\t"+usePath)
			inserted = true
		}
		if inUseBlock && trimmed == ")" {
			inUseBlock = false
		}
		out = append(out, line)
	}
	if !inserted {
		out = append(out, "use (", "\t"+usePath, ")")
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644)
}
