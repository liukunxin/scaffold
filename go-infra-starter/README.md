# go-infra-starter

一个可单独提取的新项目骨架，目标是：

- 同时提供两种分层示例：`controller -> logic -> service -> dao`（四层）与 `controller -> service`（三层）
- 增加 `internal/infra` 技术适配层，隔离外部组件细节
- 使用 `go-infra` 完成 config/log/trace 基线初始化，并按需安装 mysql/redis/metrics 等能力
- 通过 `bootstrap` 统一启动编排，保证入口清晰

## 目录

```text
go-infra-starter/
├─ cmd/http/main.go
├─ cmd/grpc/main.go
├─ configs/
├─ internal/
│  ├─ app/                          # 业务域（垂直切片）
│  │  ├─ user/
│  │  │  ├─ controller/
│  │  │  ├─ logic/
│  │  │  ├─ service/
│  │  │  ├─ dao/
│  │  │  ├─ model/
│  │  │  ├─ ro/
│  │  │  ├─ dto/
│  │  │  ├─ vo/
│  │  │  ├─ convert/
│  │  │  └─ codes/
│  │  ├─ demo/
│  │  │  ├─ controller/
│  │  │  ├─ grpc/
│  │  │  ├─ service/
│  │  │  ├─ dto/
│  │  │  └─ vo/
│  │  ├─ realtime/                  # WebSocket demo（/ws）
│  │  └─ llm/                       # 仅 --features llm 时由 CLI 注入
│  │     ├─ controller/
│  │     ├─ service/
│  │     ├─ ro/
│  │     ├─ dto/
│  │     └─ vo/
│  ├─ bootstrap/
│  ├─ infra/                        # 技术适配层
│  │  ├─ ai/                        # 仅 --features llm 时由 CLI 注入
│  │  ├─ config/
│  └─ route/
└─ go.mod
```

## 分层职责

- `controller`: 参数绑定、鉴权上下文读取、统一响应
- `logic`: 用例编排、跨 service 聚合、事务边界入口（按需使用）
- `service`: 业务规则、状态流转、外部能力调用
- `dao`: 数据访问与查询封装，不承载业务决策
- `infra`: 技术细节适配（配置、日志、追踪、存储、运行时）

示例约定：

- `user` 模块演示四层（`controller -> logic -> service -> dao`）
- `demo` 模块演示三层（`controller -> service`，无 `logic`）
- `llm` 模块（`--features llm`）演示 LLM 接入与三层调用链
- 对象按模块组织（`ro/dto/vo/convert`）；跨项目复用能力放入 SDK，不在 app 层做全局对象池

## 基础能力（默认启用）

以下能力**无需 add**，项目启动时默认接入，生成代码应直接复用：

| 能力 | SDK | 说明 |
|------|-----|------|
| 日志 | `pkg/base/log` | `bootstrap` 中直接 `log.Init` |
| 链路 | `pkg/base/trace` | `bootstrap` 中直接 `trace.Init` + 路由中间件 |
| 错误 | `pkg/base/errors` | 业务层 `WrapError`；Controller 继承 `GinBase.ErrorResponse` |

示例：`user` 模块查询不存在用户时返回 `errors.WrapError(StatusNotFound, ...)`，由 Controller 统一输出 JSON。

## 配置约定说明

`internal/infra/config/App` 遵循“优先复用 SDK 配置结构”的原则：

- **来自 go-infra SDK 的字段（直接复用）**
  - `log`: `pkg/base/log.Config`
  - `trace`: `pkg/base/trace.Config`
  - `mysql`: `pkg/infra/mysql.Config`
  - `redis`: `pkg/infra/redis/v8.Config`
  - `http_client`: `pkg/infra/http_client.Config`
  - `llm`: `pkg/infra/llm.Config`
- **项目自定义字段（编排层）**
  - `app_name`: 服务名（用于 trace/metrics 默认命名）
  - `server.address`: HTTP 监听地址
  - `traffic`: 流控策略参数（`add traffic` 后生效）

可选能力**不用 `features.*` 开关**：是否启用由 `go-infra-cli add/remove` 在 `internal/bootstrap/app.go` 注入或移除 SDK 初始化 wiring 决定。

说明：`configs/config.yml` 中可出现 mysql/redis/http_client/traffic/llm 等配置段，作为参数模板；仅在对应能力被安装后才会被实际初始化使用。配置型 feature 不会在项目内额外生成 `internal/infra/*` 实现文件，而是直接注入 SDK 初始化 wiring。

数据与接口约定（详见 `.cursor/rules` 与 `AGENTS.md`）：

- MySQL 表设计**不使用外键**，一致性由应用层保证。
- 常规 JSON API **不负责**前端页面跳转、Cookie 写入等；SSE/OAuth 等场景除外。

最小配置（无需 add 可选能力）：

```yaml
app_name: myapp
server:
  address: ":8080"
grpc:
  address: ":9091"

log:
  level: 1
trace:
  service_name: myapp
```

安装可选能力后，再补对应配置（示例）：

```yaml
mysql:
  dsn: "user:pass@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  mode: single
  addresses: ["127.0.0.1:6379"]
```

`init --features llm`（或后续接入 llm 实现）后，在 `llm.providers` 中配置 provider，例如：

```yaml
llm:
  default_provider: deepseek
  default_model: deepseek-chat
  providers:
    deepseek:
      type: openai_compatible
      base_url: https://api.deepseek.com/v1
      api_key_env: DEEPSEEK_API_KEY
```

## 快速运行

1. 复制该目录到你的新仓库根目录
2. 在新仓库执行：

```bash
go mod tidy
go run ./cmd/http
go run ./cmd/grpc
```

模板默认包含基础 `.gitignore`（构建产物、日志、环境文件、常见 IDE/OS 噪音文件）。

默认监听 `:8080`，示例接口：

- `GET /health`
- `POST /api/v1/users`
- `GET /api/v1/users/:id`
- `GET /api/v1/demo/ping?name=go-infra`
- `GET /ws`（WebSocket demo）

三层示例：

```bash
curl "http://127.0.0.1:8080/api/v1/demo/ping?name=go-infra"
```

gRPC 示例（无 proto 文件，使用内置消息类型）：

```bash
grpcurl -plaintext 127.0.0.1:9091 demo.DemoService/Ping
```

启用 `--features llm` 后额外提供 `POST /api/v1/llm/ping`（需配置 `llm.providers`），例如：

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/llm/ping" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"用一句话介绍 go-infra"}'
```

> 说明：示例 `user dao` 采用内存实现，便于开箱即跑。真实项目中将 `dao` 替换为 MySQL/GORM 实现即可。

## Cursor 协作规范

生成的业务项目用 `go-infra-cli add <feature>` 按需安装；不安装即不初始化对应 SDK。

通过 `go-infra-cli` 初始化时，项目会默认包含：

- `AGENTS.md`: Agent 协作约定与分层边界说明
- `.cursor/rules/00-architecture.mdc`: 分层架构、依赖方向与基础变更安全约束
- `.cursor/rules/10-go-sdk-first.mdc`: SDK 优先复用与 Go 实现一致性约束

