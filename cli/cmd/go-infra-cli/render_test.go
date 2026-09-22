package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceImportPrefix(t *testing.T) {
	in := "package main\n\nimport (\n\t\"single-starter/internal/route\"\n)\n"
	want := "package main\n\nimport (\n\t\"github.com/acme/demo/internal/route\"\n)\n"
	if got := replaceImportPrefix(in, singleStarterName, "github.com/acme/demo"); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// 多模块模板里，app 自己的 go.mod 会以「模板名 + 子路径」出现，
// 还有 require/replace 指向仓库根模块，三种形态都必须换掉。
func TestReplaceModModuleRefs(t *testing.T) {
	in := "module monorepo-starter/apps/gateway\n\ngo 1.25\n\nrequire (\n\tmonorepo-starter v0.0.0\n)\n\nreplace monorepo-starter => ../..\n"
	want := "module github.com/acme/mono/apps/gateway\n\ngo 1.25\n\nrequire (\n\tgithub.com/acme/mono v0.0.0\n)\n\nreplace github.com/acme/mono => ../..\n"
	if got := replaceModModuleRefs(in, monorepoStarterName, "github.com/acme/mono"); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// 根模块的 module 行没有子路径，走的是同一个分支。
func TestReplaceModModuleRefs_RootModule(t *testing.T) {
	in := "module monorepo-starter\n\ngo 1.25\n"
	if got := replaceModModuleRefs(in, monorepoStarterName, "github.com/acme/mono"); !strings.HasPrefix(got, "module github.com/acme/mono\n") {
		t.Fatalf("root module line not rewritten: %q", got)
	}
	// 不能误伤别的模块：只按行首 token 匹配。
	in2 := "require (\n\tgithub.com/other/monorepo-starter v1.0.0\n)\n"
	if got := replaceModModuleRefs(in2, monorepoStarterName, "github.com/acme/mono"); got != in2 {
		t.Fatalf("unrelated module was rewritten: %q", got)
	}
}

// 模块路径与展示名分开替换：文档标题要变成项目名，不能变成模块路径。
func TestRenderTemplate_SplitsModulePathAndDisplayName(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"go.mod":                 "module single-starter\n",
		"cmd/http/main.go":       "package main\n\nimport \"single-starter/internal/bootstrap\"\n",
		"README.md":              "# single-starter\n",
		".cursor/rules/12-x.mdc": "# HTTP Routing Rule（single-starter）\n",
		"configs/config.yml":     "app_name: single-starter\ntrace:\n  service_name: single-starter\n",
	})

	if err := renderTemplate(dir, singleStarterName, "github.com/acme/demo", "demo"); err != nil {
		t.Fatal(err)
	}

	doc := readFixture(t, dir, ".cursor/rules/12-x.mdc")
	if !strings.Contains(doc, "（demo）") {
		t.Fatalf("doc title should use the project name, got %q", doc)
	}
	if strings.Contains(doc, "github.com/acme/demo") {
		t.Fatalf("doc title must not become a module path, got %q", doc)
	}

	goSrc := readFixture(t, dir, "cmd/http/main.go")
	if !strings.Contains(goSrc, `"github.com/acme/demo/internal/bootstrap"`) {
		t.Fatalf("unexpected import line: %q", goSrc)
	}

	if got := strings.TrimSpace(readFixture(t, dir, "go.mod")); got != "module github.com/acme/demo" {
		t.Fatalf("go.mod = %q", got)
	}
}

func TestUpdateConfigYAML_CoversEnvOverrides(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"configs/config.yml":       "app_name: single-starter\ntrace:\n  service_name: single-starter\n",
		"configs/config.local.yml": "log:\n  formatter: text\n",
		"configs/config.prod.yml":  "app_name: single-starter\ntrace:\n  service_name: single-starter\n",
	})

	if err := updateConfigYAML(filepath.Join(dir, "configs"), "demo"); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"configs/config.yml", "configs/config.prod.yml"} {
		content := readFixture(t, dir, rel)
		if !strings.Contains(content, "app_name: demo") || !strings.Contains(content, "service_name: demo") {
			t.Fatalf("%s not updated:\n%s", rel, content)
		}
	}
}

func TestDropGoModRequires(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/app\n\nrequire (\n\tgoogle.golang.org/grpc v1.74.2\n\tgoogle.golang.org/protobuf v1.36.9 // indirect\n\tgithub.com/gin-gonic/gin v1.11.0\n)\n",
	})
	if err := dropGoModRequires(filepath.Join(dir, "go.mod"), []string{"google.golang.org/grpc", "google.golang.org/protobuf"}); err != nil {
		t.Fatal(err)
	}
	content := readFixture(t, dir, "go.mod")
	if strings.Contains(content, "grpc") || strings.Contains(content, "protobuf") {
		t.Fatalf("requires not dropped:\n%s", content)
	}
	if !strings.Contains(content, "github.com/gin-gonic/gin") {
		t.Fatalf("unrelated require dropped:\n%s", content)
	}
}

func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func readFixture(t *testing.T, dir, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
