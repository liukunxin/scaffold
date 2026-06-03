# go-infra-cli

`go-infra-cli` 用于生成两种项目布局：

- `single-starter`：单服务项目
- `monorepo-starter`：多应用 + 多领域的 monorepo 项目

## 安装

```bash
go install github.com/liukunxin/scaffold/cli/cmd/go-infra-cli@latest
```

## 通用参数（init）

- `--module`：Go module 名（默认=项目名）
- `--app-name`：配置中的 `app_name`（默认=项目名）
- `--layout`：`single|monorepo`（默认 `single`）
- `--output`：输出目录（默认当前目录）
- `--template`：模板目录（不传则按 `layout` 自动查找）
- `--force`：目标目录存在时覆盖
- `--skip-tidy`：跳过 `go mod tidy`

> `.cursor/rules` 与 `AGENTS.md` 会随模板一起生成。

## Single 项目（single-starter）

### 初始化

```bash
# 最小项目（默认仅 http 场景）
go-infra-cli init myapp --module github.com/acme/myapp --layout single

# 开启 grpc / ws 场景
go-infra-cli init myapp --module github.com/acme/myapp --scenes http,grpc,ws

# 初始化时按需安装能力
go-infra-cli init myapp \
  --module github.com/acme/myapp \
  --features mysql,redis,metrics,pprof,http-client,traffic,llm
```

### single 专属参数

- `--features`：仅作用于 `single`
- `--scenes`：仅作用于 `single`，支持 `http,grpc,ws`（默认 `http`）

### 能力增删（仅 single）

```bash
go-infra-cli add redis --dir ./myapp
go-infra-cli remove redis --dir ./myapp
```

- 支持：`mysql`、`redis`、`metrics`、`pprof`、`http-client`、`traffic`
- `llm` 不支持 `add/remove`，仅支持 `init --features llm`
- 原则：只通过 bootstrap wiring 决定是否启用能力，不使用 `features.*` 配置开关

## Monorepo 项目（monorepo-starter）

### 初始化

```bash
go-infra-cli init collab-platform \
  --module github.com/acme/collab-platform \
  --layout monorepo
```

### 增量扩展骨架

```bash
go-infra-cli mono add app collab-gateway --dir ./collab-platform
go-infra-cli mono add domain domain-ocr --dir ./collab-platform
```

- `mono add app <name>`：新增 `apps/<name>`，并自动更新 `go.work use`
- `mono add domain <name>`：新增 `domains/<name>`
- 命名约束：`^[a-z][a-z0-9-]*$`
- 仅做骨架编排，基础能力统一复用 `github.com/liukunxin/go-infra`

## 构建版本

```bash
go build -ldflags "-X main.version=v0.1.0" -o bin/go-infra-cli ./cmd/go-infra-cli
```
