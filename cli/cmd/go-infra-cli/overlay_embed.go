package main

import (
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed all:_features/llm/internal
var llmOverlayFS embed.FS

const llmOverlayBase = "_features/llm/internal"

// copyEmbeddedLLMOverlay 把覆盖文件落地到项目。写盘前先做与其它生成文件同样的
// 两步变换（模板模块名 → 目标模块名、gofmt），这样 remove 时才有确定的比对基准，
// 不必依赖调用方"记得"再跑一遍 rewrite/format。
func copyEmbeddedLLMOverlay(projectDir string) error {
	modulePath, err := readModulePath(projectDir)
	if err != nil {
		return err
	}
	return fs.WalkDir(llmOverlayFS, llmOverlayBase, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, llmOverlayBase), "/")
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
		return os.WriteFile(dstPath, renderOverlayFile(data, modulePath), 0o644)
	})
}

// renderOverlayFile 把内嵌原件渲染成项目里的最终形态。
// 模板渲染阶段（模块名还是模板名）调用时替换是恒等变换，由 renderTemplate 兜底完成，
// 因此两条路径（init / add）结果一致。
func renderOverlayFile(data []byte, modulePath string) []byte {
	content := replaceImportPrefix(string(data), singleStarterName, modulePath)
	if formatted, err := format.Source([]byte(content)); err == nil {
		content = string(formatted)
	}
	return []byte(content)
}

// removeEmbeddedLLMOverlay 是 copyEmbeddedLLMOverlay 的逆操作。
//
// 只删「与内嵌原件一致」的文件：用户自己改过（甚至写了业务逻辑）的文件保留并提示。
// 卸载一个能力不该顺手删掉别人的代码；不一致就交回人工判断。
func removeEmbeddedLLMOverlay(projectDir string) (bool, error) {
	modulePath, err := readModulePath(projectDir)
	if err != nil {
		return false, err
	}

	var files, dirs []string
	err = fs.WalkDir(llmOverlayFS, llmOverlayBase, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, llmOverlayBase), "/")
		if rel == "" {
			return nil
		}
		if d.IsDir() {
			dirs = append(dirs, rel)
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return false, err
	}

	changed := false
	var kept []string
	for _, rel := range files {
		dstPath := filepath.Join(projectDir, "internal", filepath.FromSlash(rel))
		info, statErr := os.Stat(dstPath)
		if statErr != nil || !info.Mode().IsRegular() {
			continue
		}
		same, err := matchesEmbeddedOverlay(dstPath, rel, modulePath)
		if err != nil {
			return changed, err
		}
		if !same {
			kept = append(kept, filepath.ToSlash(filepath.Join("internal", rel)))
			continue
		}
		if err = os.Remove(dstPath); err != nil {
			return changed, err
		}
		changed = true
	}

	removeEmptyOverlayDirs(projectDir, dirs)

	if len(kept) > 0 {
		fmt.Printf("llm: kept %d file(s) that differ from the original overlay — removal is incomplete, clean them up manually: %s\n",
			len(kept), strings.Join(kept, ", "))
	}
	return changed, nil
}

// removeEmptyOverlayDirs 回收覆盖能力留下的空目录（子目录先于父目录）。
//
// 必须自己判空后再删，不能依赖 os.Remove 的"只删空目录"语义：实测在启用了
// POSIX 删除语义的 NTFS 卷上，os.Remove 对非空目录同样返回成功，并把整棵子树删掉。
// 而 internal/app、internal/infra、internal/route 都是与模板共用的目录，
// 一旦误删就把项目里其它代码一起带走。
func removeEmptyOverlayDirs(projectDir string, dirs []string) {
	for _, rel := range deepestFirst(dirs) {
		dir := filepath.Join(projectDir, "internal", filepath.FromSlash(rel))
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			continue
		}
		_ = os.Remove(dir)
	}
}

// matchesEmbeddedOverlay 判断项目里的覆盖文件是否仍是「原件」。
// 比较基准与写盘时完全一致（同一个 renderOverlayFile），因此只对"用户改过"敏感。
func matchesEmbeddedOverlay(dstPath, rel, modulePath string) (bool, error) {
	data, err := fs.ReadFile(llmOverlayFS, path.Join(llmOverlayBase, filepath.ToSlash(rel)))
	if err != nil {
		return false, fmt.Errorf("read embedded llm overlay file %s: %w", rel, err)
	}
	got, err := os.ReadFile(dstPath)
	if err != nil {
		return false, err
	}
	// 两边都过一遍 render：用户只做了 gofmt 不算改动。
	return string(renderOverlayFile(got, modulePath)) == string(renderOverlayFile(data, modulePath)), nil
}

// deepestFirst 让子目录排在父目录前面，回收目录时不会因父目录非空而提前失败。
func deepestFirst(dirs []string) []string {
	out := append([]string(nil), dirs...)
	sort.Slice(out, func(i, j int) bool {
		di, dj := strings.Count(out[i], "/"), strings.Count(out[j], "/")
		if di != dj {
			return di > dj
		}
		return len(out[i]) > len(out[j])
	})
	return out
}
