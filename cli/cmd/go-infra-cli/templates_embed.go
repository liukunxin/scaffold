package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 模板随 CLI 一起编译进二进制，因此在任意目录执行 init 都不需要 --template，
// 也不依赖 CLI 的安装位置。下面的目录布局是被 Go 的两条规则逼出来的，别随手改：
//
//  1. 模板目录名必须以 `_` 开头（这里是 _templates）。模板里全是 .go 文件，
//     目录名不带 `_` 就会被 CLI 自己的 go build ./... 当成普通包编译并直接报错
//     （实测 `templates/` → "package nonexistent-module/internal/route is not in std"）。
//     go 工具链会忽略 `_` / `.` 开头的目录与 testdata，三者都可用；这里选 `_templates`
//     是为了和同目录下已有的 _features/ 保持对称。
//  2. go.mod / go.sum / go.work 在模板里必须写成 *.tmpl。只要模板目录下存在 go.mod，
//     go 就把它视作嵌套 module，//go:embed 会直接报 "contains no embeddable files"。
//     copyTree 落地时再把 .tmpl 去掉。
//
// all: 前缀同样是必须的，否则 .gitignore / .cursor 这类点开头的文件会被默认排除
// （也是它让 `_templates` 能被 embed 读到；不带 all: 时 `_` 开头的目录会被跳过）。
//
//go:embed all:_templates
var templatesFS embed.FS

const (
	templatesRoot = "_templates"

	// templateFileSuffix 是模块元数据在模板里的后缀，落地时会被去掉。
	templateFileSuffix = ".tmpl"

	singleStarterName   = "single-starter"
	monorepoStarterName = "monorepo-starter"
)

// templateSource 描述模板的读取源：内嵌 FS，或 --template 指定的本地目录（开发时覆盖）。
type templateSource struct {
	fsys fs.FS
	root string
	// name 是模板目录名，同时决定渲染时要替换的模块根名与展示名。
	name string
}

func resolveTemplate(explicit, layout string) (*templateSource, error) {
	name := singleStarterName
	if layout == "monorepo" {
		name = monorepoStarterName
	}

	if strings.TrimSpace(explicit) != "" {
		abs, err := validateTemplateDir(explicit)
		if err != nil {
			return nil, err
		}
		return &templateSource{fsys: os.DirFS(abs), root: ".", name: name}, nil
	}

	root := path.Join(templatesRoot, name)
	if _, err := fs.Stat(templatesFS, root); err != nil {
		return nil, fmt.Errorf("embedded template %s not found: %w", root, err)
	}
	return &templateSource{fsys: templatesFS, root: root, name: name}, nil
}

func validateTemplateDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if !dirExists(abs) {
		return "", fmt.Errorf("template directory not found: %s", abs)
	}
	if !fileExists(filepath.Join(abs, "go.mod")) &&
		!fileExists(filepath.Join(abs, "go.mod"+templateFileSuffix)) &&
		!fileExists(filepath.Join(abs, "go.work")) &&
		!fileExists(filepath.Join(abs, "go.work"+templateFileSuffix)) {
		return "", fmt.Errorf("template directory %s missing go.mod/go.work", abs)
	}
	return abs, nil
}

// templateDestName 把模板文件还原成项目里的真实名字（go.mod.tmpl → go.mod）。
func templateDestName(rel string) string {
	return strings.TrimSuffix(rel, templateFileSuffix)
}

// copyTree 把模板树原样拷到目标目录。_features/ 是 CLI 的能力覆盖文件，不属于项目本身。
func copyTree(src *templateSource, dstDir string) error {
	return fs.WalkDir(src.fsys, src.root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := p
		if src.root != "." {
			rel = strings.TrimPrefix(strings.TrimPrefix(p, src.root), "/")
		}
		if rel == "." {
			return os.MkdirAll(dstDir, 0o755)
		}
		if rel == "_features" || strings.HasPrefix(rel, "_features/") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		dstPath := filepath.Join(dstDir, filepath.FromSlash(templateDestName(rel)))
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}
		data, err := fs.ReadFile(src.fsys, p)
		if err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	})
}
