package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// sceneSpec 声明一个运行时场景「拥有」哪些文件与标记。
// 场景关闭时：删掉 files、剥离 markers、从 go.mod 移除 dropRequires；
// 场景开启时只剥离 markers（标记不该出现在生成物里）。
type sceneSpec struct {
	name         string
	files        []string
	markers      []sceneMarker
	dropRequires []string
}

type sceneMarker struct {
	file string
	key  string
}

var sceneNames = []string{"http", "grpc", "ws"}

var sceneSpecs = []sceneSpec{
	{
		name: "grpc",
		files: []string{
			"cmd/grpc",
			"internal/bootstrap/grpc.go",
			"internal/route/grpc.go",
			"internal/app/demo/grpc",
		},
		dropRequires: []string{
			"google.golang.org/grpc",
			"google.golang.org/protobuf",
		},
	},
	{
		name: "ws",
		files: []string{
			"internal/app/realtime",
		},
		markers: []sceneMarker{
			{file: routeFilePath, key: "SCENE_WS"},
		},
	},
}

// parseInitScenesArg 解析 --scenes。默认全部开启；显式传入时以传入为准。
// http 始终保留：route/ 与 /health 挂在它下面，ws 也依赖它。
func parseInitScenesArg(raw string) (map[string]bool, error) {
	scenes := make(map[string]bool, len(sceneNames))
	raw = strings.TrimSpace(raw)

	if raw == "" {
		for _, name := range sceneNames {
			scenes[name] = true
		}
	} else {
		for _, part := range strings.Split(raw, ",") {
			scene := strings.ToLower(strings.TrimSpace(part))
			if scene == "" {
				continue
			}
			if !containsString(sceneNames, scene) {
				return nil, fmt.Errorf("unsupported scene %q: only %s are allowed", scene, strings.Join(sceneNames, ","))
			}
			scenes[scene] = true
		}
	}

	scenes["http"] = true
	return scenes, nil
}

func containsString(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// applySceneSelection 按场景清单裁剪生成物。
func applySceneSelection(projectDir string, scenes map[string]bool) error {
	for _, spec := range sceneSpecs {
		enabled := scenes[spec.name]

		if !enabled {
			for _, rel := range spec.files {
				if err := os.RemoveAll(filepath.Join(projectDir, filepath.FromSlash(rel))); err != nil {
					return fmt.Errorf("remove %s (scene %s): %w", rel, spec.name, err)
				}
			}
			if len(spec.dropRequires) > 0 {
				if err := dropGoModRequires(filepath.Join(projectDir, "go.mod"), spec.dropRequires); err != nil {
					return err
				}
			}
		}

		for _, marker := range spec.markers {
			path := filepath.Join(projectDir, filepath.FromSlash(marker.file))
			if err := keepSceneMarkers(path, marker.key, enabled); err != nil {
				return fmt.Errorf("apply %s markers in %s: %w", marker.key, marker.file, err)
			}
		}
	}
	return nil
}

// keepSceneMarkers 剥离 SCENE_* 标记本身；keep=false 时连同标记内的代码一起删除。
func keepSceneMarkers(filePath, sceneKey string, keep bool) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(data)
	startMarker := "// " + sceneKey + "_START"
	endMarker := "// " + sceneKey + "_END"

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == startMarker {
			inBlock = true
			continue
		}
		if trimmed == endMarker {
			inBlock = false
			continue
		}
		if inBlock && !keep {
			continue
		}
		out = append(out, line)
	}

	return os.WriteFile(filePath, []byte(strings.Join(out, "\n")), 0o644)
}
