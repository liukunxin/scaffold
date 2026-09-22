package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateStarterLayout_MissingAnchors(t *testing.T) {
	if err := validateStarterLayout(t.TempDir()); err == nil {
		t.Fatal("expected layout validation error for an empty directory")
	}
}

func TestValidateStarterLayout_MissingFeatureAnchors(t *testing.T) {
	dir := writeLayoutFixture(t, "package bootstrap\n")
	if err := validateStarterLayout(dir); err == nil {
		t.Fatal("expected error when feature anchors are absent")
	}
}

func TestValidateStarterLayout_RejectsLegacyFeatures(t *testing.T) {
	dir := writeLayoutFixture(t, featureAnchorSource())
	legacy := "app_name: test\nfeatures:\n  mysql: true\n"
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yml"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateStarterLayout(dir); err == nil {
		t.Fatal("expected legacy features rejection")
	}
}

// 业务模块被删掉之后 add/remove 仍应可用：校验只看关键锚点，不绑业务路径。
func TestValidateStarterLayout_OK(t *testing.T) {
	dir := writeLayoutFixture(t, featureAnchorSource())
	if err := validateStarterLayout(dir); err != nil {
		t.Fatalf("expected layout ok, got %v", err)
	}
}

// 随 CLI 分发的模板本体必须满足 add/remove 的锚点契约，否则生成物一 add 就炸。
// 校验对象是「恢复真实文件名之后」的落地产物，所以先 copyTree 再校验。
func TestValidateStarterLayout_EmbeddedTemplate(t *testing.T) {
	src, err := resolveTemplate("", "single")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = copyTree(src, dir); err != nil {
		t.Fatal(err)
	}
	if err = validateStarterLayout(dir); err != nil {
		t.Fatalf("embedded single-starter violates the layout contract: %v", err)
	}
}

func writeLayoutFixture(t *testing.T, bootstrapSource string) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"configs", "internal/bootstrap", "internal/route"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(sub)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"go.mod":             "module example.com/app\n",
		"configs/config.yml": "app_name: test\n",
		bootstrapFilePath:    bootstrapSource,
		routeFilePath:        routeAnchorSource(),
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func featureAnchorSource() string {
	return "package bootstrap\n\nimport (\n" +
		"\t" + anchorImportsStart + "\n\t" + anchorImportsEnd + "\n)\n\n" +
		"func New() (*App, error) {\n" +
		"\t" + anchorInitStart + "\n\t" + anchorInitEnd + "\n" +
		"\t" + anchorRouterStart + "\n\t" + anchorRouterEnd + "\n" +
		"\treturn nil, nil\n}\n\n" +
		"func (a *App) Close() {\n" +
		"\t" + anchorCloseStart + "\n\t" + anchorCloseEnd + "\n}\n"
}

func routeAnchorSource() string {
	return "package route\n\nfunc Setup() {\n" +
		"\t" + anchorRoutesStart + "\n\t" + anchorRoutesEnd + "\n}\n"
}
