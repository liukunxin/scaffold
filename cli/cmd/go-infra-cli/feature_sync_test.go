package main

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// llm 的注入发生在模板渲染之后，所以导入路径必须取目标模块。
// 曾经写死成模板名，生成物直接编译失败（`package single-starter/... is not in std`）。
func TestSyncLLMFeatureArtifacts_UsesTargetModulePath(t *testing.T) {
	dir := writeLayoutFixture(t, featureAnchorSource())

	changed, err := syncLLMFeatureArtifacts(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected the llm sections to be injected")
	}

	bootstrap := readFixture(t, dir, bootstrapFilePath)
	if !strings.Contains(bootstrap, `"example.com/app/internal/infra/ai"`) {
		t.Fatalf("llm import must use the target module path:\n%s", bootstrap)
	}
	if strings.Contains(bootstrap, singleStarterName) {
		t.Fatalf("llm import leaked the template module name:\n%s", bootstrap)
	}

	route := readFixture(t, dir, routeFilePath)
	if !strings.Contains(route, "setupLLM(api)") {
		t.Fatalf("llm route registration missing:\n%s", route)
	}

	// 重复执行必须幂等，否则 init/add 会因为标记重复而互相打架。
	changed, err = syncLLMFeatureArtifacts(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second sync should be a no-op")
	}
}

// 覆盖型能力必须"装得上也卸得掉"，且 remove 要回收自己拷进来的文件、
// 不能碰模板自带的文件。曾经的实现只允许 init 安装，llm 就没有卸载路径。
func TestSyncLLMFeature_RoundTripCleansOverlayFiles(t *testing.T) {
	dir := writeLayoutFixture(t, featureAnchorSource())

	if _, err := syncLLMFeature(dir, true); err != nil {
		t.Fatal(err)
	}
	if !isFeatureInstalled(dir, "llm") {
		t.Fatal("llm should be reported as installed")
	}
	for _, rel := range []string{
		"internal/infra/ai/llm.go",
		"internal/route/llm.go",
		"internal/app/llm/controller/controller.go",
		"internal/app/llm/service/service.go",
		"internal/app/llm/dto/ping.go",
		"internal/app/llm/ro/ping.go",
		"internal/app/llm/vo/ping.go",
	} {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if !fileExists(path) {
			t.Fatalf("%s missing after install", rel)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"`+singleStarterName+`/`) {
			t.Fatalf("%s still points at the template module:\n%s", rel, data)
		}
	}

	changed, err := syncLLMFeature(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected the llm feature to be removed")
	}
	if isFeatureInstalled(dir, "llm") {
		t.Fatal("llm should be reported as removed")
	}
	for _, rel := range []string{
		"internal/infra/ai/llm.go",
		"internal/route/llm.go",
		"internal/app/llm/controller/controller.go",
	} {
		if fileExists(filepath.Join(dir, filepath.FromSlash(rel))) {
			t.Fatalf("%s left behind after removal", rel)
		}
	}
	if dirExists(filepath.Join(dir, "internal", "app", "llm")) {
		t.Fatal("empty internal/app/llm should be reclaimed")
	}
	if !fileExists(filepath.Join(dir, filepath.FromSlash(routeFilePath))) {
		t.Fatal("removal must not touch template files")
	}
	if strings.Contains(readFixture(t, dir, bootstrapFilePath), "InitLLM") {
		t.Fatal("llm wiring left behind after removal")
	}
}

// remove 只回收"未改动过的原件"：用户改过的文件必须保留，卸载能力不该删别人的代码。
func TestSyncLLMFeature_RemoveKeepsEditedOverlayFiles(t *testing.T) {
	dir := writeLayoutFixture(t, featureAnchorSource())
	if _, err := syncLLMFeature(dir, true); err != nil {
		t.Fatal(err)
	}

	edited := filepath.Join(dir, "internal", "app", "llm", "service", "service.go")
	if err := os.WriteFile(edited, []byte("package service\n\n// my own code\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := syncLLMFeature(dir, false); err != nil {
		t.Fatal(err)
	}
	if !fileExists(edited) {
		t.Fatal("a user-edited overlay file must survive remove")
	}
}

// 配置型能力注入的是 SDK 路径，与目标模块无关，但不能把锚点插坏。
func TestSyncConfigFeatureArtifacts_RedisRoundTrip(t *testing.T) {
	dir := writeLayoutFixture(t, featureAnchorSource())

	changed, err := syncConfigFeatureArtifacts(dir, "redis", true)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected redis to be installed")
	}
	if !isFeatureInstalled(dir, "redis") {
		t.Fatal("redis should be reported as installed")
	}

	changed, err = syncConfigFeatureArtifacts(dir, "redis", false)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected redis to be removed")
	}
	if isFeatureInstalled(dir, "redis") {
		t.Fatal("redis should be reported as removed")
	}
	if strings.Contains(readFixture(t, dir, bootstrapFilePath), "iredis") {
		t.Fatal("redis lines left behind after removal")
	}
}

// 注入块必须从行首开始、统一一级 tab；否则第一行会多一层缩进，
// 生成物虽然能编译但 gofmt 不干净。
func TestInsertFeatureSection_IndentationIsUniform(t *testing.T) {
	content := "func New() error {\n\t// FEATURE_IMPORTS_START\n\t// FEATURE_IMPORTS_END\n\treturn nil\n}\n"
	sec := featureSection{
		name:  "IMPORTS",
		start: anchorImportsStart,
		end:   anchorImportsEnd,
		lines: []string{`"example.com/app/internal/infra/ai"`},
	}
	out, changed, err := insertFeatureSection(content, "llm", sec, false)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected the section to be inserted")
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "FEATURE:llm:") || strings.Contains(line, `"example.com/app`) {
			if strings.HasPrefix(line, "\t\t") {
				t.Fatalf("line has double indentation: %q", line)
			}
			if !strings.HasPrefix(line, "\t") {
				t.Fatalf("line is not indented: %q", line)
			}
		}
	}
}

// 模板 import 的排序在模块名替换后就不再成立，落盘时必须过一遍 go/format。
func TestFormatGoSources_MakesGeneratedFilesGofmtClean(t *testing.T) {
	// 模板原样：第三方在前、模板模块在后；换成 github.com/acme/... 后顺序就反了。
	src := "package bootstrap\n\nimport (\n\t\"github.com/gin-gonic/gin\"\n\t\"github.com/acme/app/internal/route\"\n)\n\nfunc New() { gin.New() }\n"
	dir := writeFixture(t, map[string]string{"internal/bootstrap/app.go": src})

	if err := formatGoSources(dir); err != nil {
		t.Fatal(err)
	}

	out := readFixture(t, dir, "internal/bootstrap/app.go")
	formatted, err := format.Source([]byte(out))
	if err != nil {
		t.Fatalf("generated file does not parse: %v\n%s", err, out)
	}
	if string(formatted) != out {
		t.Fatalf("file is not gofmt-clean:\n%s", out)
	}
	if strings.Index(out, "github.com/acme/app") > strings.Index(out, "github.com/gin-gonic/gin") {
		t.Fatalf("imports should be sorted:\n%s", out)
	}
}
