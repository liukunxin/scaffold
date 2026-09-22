package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInitScenesArg_DefaultAllOn(t *testing.T) {
	scenes, err := parseInitScenesArg("")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range sceneNames {
		if !scenes[name] {
			t.Fatalf("scene %s should default to on", name)
		}
	}
}

func TestParseInitScenesArg_ExplicitKeepsHTTP(t *testing.T) {
	scenes, err := parseInitScenesArg("grpc")
	if err != nil {
		t.Fatal(err)
	}
	if !scenes["http"] || !scenes["grpc"] || scenes["ws"] {
		t.Fatalf("unexpected scenes: %+v", scenes)
	}
}

func TestParseInitScenesArg_RejectsUnknown(t *testing.T) {
	if _, err := parseInitScenesArg("sse"); err == nil {
		t.Fatal("expected error for unsupported scene")
	}
}

func TestApplySceneSelection_RemovesGRPCAndStripsMarkers(t *testing.T) {
	dir := t.TempDir()
	for _, rel := range []string{"cmd/grpc", "internal/app/demo/grpc/view"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel), "main.go"), []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "internal/bootstrap"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal/bootstrap/grpc.go"), []byte("package bootstrap\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "internal/route"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal/route/grpc.go"), []byte("package route\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	routeSrc := "package route\n\nfunc Setup(router *gin.Engine) {\n\t// SCENE_WS_START\n\trouter.GET(\"/ws\", h)\n\t// SCENE_WS_END\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "internal/route/init.go"), []byte(routeSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	goMod := "module example.com/app\n\nrequire (\n\tgoogle.golang.org/grpc v1.74.2\n\tgoogle.golang.org/protobuf v1.36.9\n\tgithub.com/gin-gonic/gin v1.11.0\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	scenes, err := parseInitScenesArg("http")
	if err != nil {
		t.Fatal(err)
	}
	if err = applySceneSelection(dir, scenes); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"cmd/grpc", "internal/bootstrap/grpc.go", "internal/route/grpc.go", "internal/app/demo/grpc"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			t.Fatalf("%s should be removed when the grpc scene is off", rel)
		}
	}

	routeData, err := os.ReadFile(filepath.Join(dir, "internal/route/init.go"))
	if err != nil {
		t.Fatal(err)
	}
	routeContent := string(routeData)
	if strings.Contains(routeContent, "SCENE_WS") || strings.Contains(routeContent, "/ws") {
		t.Fatalf("ws markers and body should be stripped, got:\n%s", routeContent)
	}

	goModData, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	goModContent := string(goModData)
	if strings.Contains(goModContent, "google.golang.org/grpc") || strings.Contains(goModContent, "protobuf") {
		t.Fatalf("grpc requires should be dropped, got:\n%s", goModContent)
	}
	if !strings.Contains(goModContent, "github.com/gin-gonic/gin") {
		t.Fatalf("unrelated requires must survive, got:\n%s", goModContent)
	}
}

func TestApplySceneSelection_KeepsWSBodyButStripsMarkers(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "internal/route"), 0o755); err != nil {
		t.Fatal(err)
	}
	routeSrc := "package route\n\nfunc Setup() {\n\t// SCENE_WS_START\n\twired := 1\n\t// SCENE_WS_END\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "internal/route/init.go"), []byte(routeSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	scenes, err := parseInitScenesArg("")
	if err != nil {
		t.Fatal(err)
	}
	if err = applySceneSelection(dir, scenes); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "internal/route/init.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "SCENE_WS") {
		t.Fatalf("markers must not ship, got:\n%s", content)
	}
	if !strings.Contains(content, "wired := 1") {
		t.Fatalf("ws body must survive when the scene is on, got:\n%s", content)
	}
}
