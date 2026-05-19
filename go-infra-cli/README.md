# go-infra-cli

`go-infra-cli` 用于按 `go-infra-starter` 模板快速生成新项目。

## 安装

```bash
go install github.com/liukunxin/scaffold/go-infra-cli/cmd/go-infra-cli@latest
```

## 使用

```bash
go-infra-cli version

# 最小项目（仅 log/trace/errors，不装 mysql/redis 等可选能力）
go-infra-cli init myapp --module github.com/acme/myapp --output .

# 按需安装能力（与 add 相同：注入 bootstrap 初始化 wiring）
go-infra-cli init myapp \
  --module github.com/acme/myapp \
  --features mysql,redis,metrics,pprof,http-client,traffic,llm

cd myapp
go-infra-cli add redis
go-infra-cli remove redis
```

## 设计原则

- **装什么有什么**：可选能力通过 `bootstrap` 是否注入对应 SDK wiring 决定，不用 `features.*` 配置开关。
- **配置只描述连接参数**：`mysql.dsn`、`redis.addresses` 等；未 `add` 的模块不会注入对应 SDK 初始化。
- `init --features llm` 使用内置（embed）`cmd/go-infra-cli/_features/llm` 叠加资产；配置型 `add/remove` 直接修改 `bootstrap` 初始化 wiring。

## add / remove

仅在标准 `go-infra-starter` 目录结构下生效；结构不一致时拒绝并提示手动处理。

- 支持：`mysql`、`redis`、`metrics`、`pprof`、`http-client`、`traffic`
- **不支持** `llm` 的 add/remove（用 `init --features llm`）
- **add**：在 `internal/bootstrap/app.go` 注入对应 SDK 初始化/关闭 wiring
- **remove**：移除对应 wiring（不再保留 no-op stub）
- **幂等**：已是目标状态时跳过

## init 参数

- `--module`: Go module 名，默认项目名
- `--app-name`: 配置里的 `app_name`，默认项目名
- `--output`: 输出目录，默认当前目录
- `--template`: 模板目录；不传时自动查找 `go-infra-starter`
- `--force`: 目标目录存在时覆盖
- `--features`: 逗号分隔，**仅列出的能力会安装**；不传则一个都不装（仅基线能力）
- `--skip-tidy`: 跳过 `go mod tidy`

> `.cursor/rules` 与 `AGENTS.md` 会始终随模板生成。

## 基线能力（无需 add）

- `pkg/base/log`、`pkg/base/trace`、`pkg/base/errors` — 始终接入，非可选模块。

## Feature 说明

| Feature | 安装后 |
|---------|--------|
| `mysql` | 注入 MySQL SDK 初始化/关闭 wiring，按 `mysql.dsn` 生效 |
| `redis` | 注入 Redis SDK 初始化/关闭 wiring，按 `redis.addresses` 生效 |
| `metrics` | 注入 metrics wiring，在路由注册 `/metrics` |
| `pprof` | 注入 pprof 启动 wiring |
| `http-client` | 注入 HTTP client SDK 初始化 wiring |
| `traffic` | 注入限流 controller + traffic 初始化 wiring |
| `llm` | 仅 `init --features llm`：注入 `app/llm`、`infra/ai` 与路由 |

## 版本注入

```bash
go build -ldflags "-X main.version=v0.1.0" -o bin/go-infra-cli ./cmd/go-infra-cli
```
