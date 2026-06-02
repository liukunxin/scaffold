package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateStarterLayout_MissingPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configsDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configsDir, "config.yml"), []byte("app_name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateStarterLayout(dir); err == nil {
		t.Fatal("expected layout validation error")
	}
}

func TestValidateStarterLayout_RejectsLegacyFeatures(t *testing.T) {
	dir := t.TempDir()
	for _, rel := range starterLayoutPaths {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	configsDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configsDir, "config.yml"), []byte("app_name: test\nfeatures:\n  mysql: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateStarterLayout(dir); err == nil {
		t.Fatal("expected legacy features rejection")
	}
}

func TestValidateStarterLayout_OK(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	starter := filepath.Clean(filepath.Join(wd, "..", "..", "..", "single-starter"))
	if _, err := os.Stat(filepath.Join(starter, "go.mod")); err != nil {
		t.Skip("single-starter not found beside cli module")
	}
	if err := validateStarterLayout(starter); err != nil {
		t.Fatalf("expected starter layout ok: %v", err)
	}
}
