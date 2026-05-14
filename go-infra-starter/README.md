# go-infra-starter

一个可单独提取的新项目骨架，目标是：

- 保留 `controller -> logic -> service -> dao` 四层风格
- 增加 `internal/infra` 技术适配层，隔离外部组件细节
- 使用 `go-infra` 完成 config/log/trace/mysql/redis/metrics 初始化
- 通过 `bootstrap` 统一启动编排，保证入口清晰

## 目录

```text
go-infra-starter/
├─ cmd/http/main.go
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
│  │  │  └─ convert/
│  │  └─ llm/
│  │     ├─ controller/
│  │     ├─ logic/
│  │     ├─ service/
│  │     ├─ ro/
│  │     ├─ dto/
│  │     └─ vo/
│  ├─ bootstrap/
│  ├─ infra/                        # 技术适配层
│  │  ├─ ai/
│  │  ├─ config/
│  │  ├─ network/
│  │  ├─ observability/
│  │  ├─ persistence/
│  │  └─ runtime/
│  └─ route/
└─ go.mod
```

## 分层职责

- `controller`: 参数绑定、鉴权上下文读取、统一响应
- `logic`: 用例编排、跨 service 聚合、事务边界入口
- `service`: 业务规则、状态流转、外部能力调用
- `dao`: 数据访问与查询封装，不承载业务决策
- `infra`: 技术细节适配（配置、日志、追踪、存储、运行时）

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
  - `features`: 运行期开关（`redis` / `metrics` / `pprof` / `http_client` / `llm`）

示例：

```yaml
app_name: myapp
server:
  address: ":8080"

log:
  level: 1
trace:
  service_name: myapp
mysql:
  dsn: ""
redis:
  mode: single
  addresses: ["127.0.0.1:6379"]

features:
  redis: false
  metrics: true
  pprof: false
  http_client: false
  llm: false

http_client:
  timeout: 30s
  max_idle_conns: 100

llm:
  default_provider: ""
  default_model: ""
  providers: {}
```

当开启 `features.llm=true` 时，需至少配置一个 provider，例如：

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
```

模板默认包含基础 `.gitignore`（构建产物、日志、环境文件、常见 IDE/OS 噪音文件）。

默认监听 `:8080`，示例接口：

- `GET /health`
- `POST /api/v1/users`
- `GET /api/v1/users/:id`
- `POST /api/v1/llm/ping`

LLM 验证示例：

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/llm/ping" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"用一句话介绍 go-infra"}'
```

> 说明：示例 `user dao` 采用内存实现，便于开箱即跑。真实项目中将 `dao` 替换为 MySQL/GORM 实现即可。

## Cursor 协作规范（可选）

当通过 `go-infra-cli` 初始化并启用 `cursor-rules` 特性时，项目会包含：

- `AGENTS.md`: Agent 协作约定与分层边界说明
- `.cursor/rules/00-architecture.mdc`: 分层架构、依赖方向与基础变更安全约束
- `.cursor/rules/10-go-sdk-first.mdc`: SDK 优先复用与 Go 实现一致性约束

如果初始化时未启用 `cursor-rules`，这些文件不会被保留。

