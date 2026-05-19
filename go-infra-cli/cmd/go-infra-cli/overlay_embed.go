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

//go:embed _features/llm/internal/**
var llmOverlayFS embed.FS

func copyEmbeddedLLMOverlay(projectDir string) error {
	base := "_features/llm/internal"
	return fs.WalkDir(llmOverlayFS, base, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(p, base)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		dstPath := filepath.Join(projectDir, "internal", filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}
		data, err := fs.ReadFile(llmOverlayFS, path.Clean(p))
		if err != nil {
			return fmt.Errorf("read embedded llm overlay file %s: %w", p, err)
		}
		if err = os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	})
}
