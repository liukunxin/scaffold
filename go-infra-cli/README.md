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
  --features mysql,metrics,pprof,http-client,traffic,llm
```

## init 参数

- `--module`: Go module 名，默认项目名
- `--app-name`: 配置里的 `app_name`，默认项目名
- `--output`: 输出目录，默认当前目录
- `--template`: 模板目录；不传时自动查找 `go-infra-starter`（也兼容 `scaffold/go-infra-starter`）
- `--force`: 目标目录存在时覆盖
- `--features`: 功能开关集合（逗号分隔），如 `mysql,redis,metrics,pprof,http-client,traffic,llm`
- `--skip-tidy`: 跳过 `go mod tidy`

> 不传 `--features` 时默认值：`mysql=true`，`redis=false`，`metrics=true`，`pprof=true`，`http-client=true`，`traffic=true`，`llm=false`。

> `.cursor/rules` 与 `AGENTS.md` 会始终随模板生成，不再由 feature 控制。

## 基础能力（默认启用，非 feature）

以下在模板中**默认接入**，无需也不应通过 `--features` 开关：

- `pkg/base/log` — 启动时初始化
- `pkg/base/trace` — 启动时初始化 + HTTP 中间件
- `pkg/base/errors` — 业务层 `WrapError`，Controller 使用 `GinBase` 统一错误响应

## Feature 说明

- `mysql`: 启用 MySQL 初始化（`features.mysql=true`）
- `redis`: 启用 Redis 初始化（`features.redis=true`）
- `metrics`: 启用 `/metrics` 指标上报（默认开启）
- `pprof`: 启用 pprof runtime 诊断入口（默认开启）
- `http-client`: 启用 `go-infra/pkg/infra/http_client` 全局客户端初始化（默认开启）
- `traffic`: 启用 `go-infra/pkg/infra/traffic` 流控初始化（默认开启，内置限流控制器）
- `llm`: 启用 `go-infra/pkg/infra/llm` 初始化（当配置了 `llm.providers` 时生效）

## 版本注入

构建时可通过 `-ldflags` 注入版本号：

```bash
go build -ldflags "-X main.version=v0.1.0" -o bin/go-infra-cli ./cmd/go-infra-cli
```

