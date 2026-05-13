# go-infra-cli

`go-infra-cli` 用于按 `go-infra-starter` 模板快速生成新项目。

## 安装

```bash
go install github.com/liukunxin/scaffold/go-infra-cli/cmd/go-infra-cli@latest
```

## 使用

```bash
go-infra-cli version

go-infra-cli init myapp \
  --module github.com/acme/myapp \
  --output . \
  --features redis,pprof,cursor-rules
```

## init 参数

- `--module`: Go module 名，默认项目名
- `--app-name`: 配置里的 `app_name`，默认项目名
- `--output`: 输出目录，默认当前目录
- `--template`: 模板目录；不传时自动查找 `go-infra-starter`（也兼容 `scaffold/go-infra-starter`）
- `--force`: 目标目录存在时覆盖
- `--features`: 功能开关集合（逗号分隔），如 `redis,metrics,pprof,cursor-rules`
- `--with-mysql`: 保留 mysql 集成骨架（默认 `true`）
- `--skip-tidy`: 跳过 `go mod tidy`

> 不传 `--features` 时默认值：`metrics=true`，`redis=false`，`pprof=false`，`cursor-rules=false`。

## Cursor 规范文件开关

`cursor-rules` 是可选特性：

- 选择 `--features` 包含 `cursor-rules`：生成 `.cursor/rules/*.mdc` 与 `AGENTS.md`
- 未选择 `cursor-rules`：不生成上述文件

示例（启用 Cursor 规范）：

```bash
go-infra-cli init myapp --output . --features metrics,cursor-rules
```

## 版本注入

构建时可通过 `-ldflags` 注入版本号：

```bash
go build -ldflags "-X main.version=v0.1.0" -o bin/go-infra-cli ./cmd/go-infra-cli
```

