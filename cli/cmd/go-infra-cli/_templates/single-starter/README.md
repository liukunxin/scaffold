# single-starter

单服务 Go 项目骨架。分层、配置、启动编排都按 `go-infra` SDK 的既有能力来，不重复造轮子。

## 目录

```text
single-starter/
├─ cmd/http/main.go                 HTTP 入口（只做装配与生命周期）
├─ cmd/grpc/main.go                 gRPC 入口（--scenes 不含 grpc 时会被移除）
├─ configs/
│  ├─ config.yml                    基础配置
│  └─ config.{local,test,gray,prod}.yml   环境覆盖
├─ internal/
│  ├─ app/                          业务域（垂直切片）
│  │  ├─ user/                     四层示例：controller → logic → service → dao
│  │  │  ├─ controller/  logic/  service/  dao/
│  │  │  ├─ model/  ro/  dto/  vo/  convert/  codes/
│  │  ├─ demo/                     三层示例：controller → service
│  │  │  ├─ controller/  service/  grpc/  dto/  vo/
│  │  └─ realtime/                 WebSocket demo（/ws）
│  ├─ bootstrap/                   启动编排：base.go（基线）+ app.go / grpc.go（进程）
│  ├─ infra/                       技术适配层
│  │  ├─ config/                   配置加载与环境覆盖
│  │  └─ notifier/                 通知端口 + 默认日志实现
│  └─ route/                       传输层注册：HTTP 路由 + gRPC 服务挂载
└─ go.mod
```

## 分层职责

- `controller`：参数绑定、统一响应（embed `kcontroller.GinBase`）
- `logic`：**跨 service / 跨域的用例编排**；单 service 直通不要加这一层
- `service`：业务规则、状态流转、外部能力调用
- `dao`：数据访问与查询封装，不承载业务决策
- `infra`：技术细节适配（配置、通知等外部依赖的端口与实现）

骨架同时保留两种形态，避免新人写无意义的中间层：

- `user` 是四层示例。`logic.CreateUser` 做真实编排：先落库，再通过 `infra/notifier` 发通知（通知失败只记日志，不回滚）。
- `demo` 是三层示例。`controller -> service` 直通，没有 `logic`。

对象按模块组织（`ro/dto/vo/convert`），跨项目复用的能力放进 SDK，不在 app 层做全局对象池。

## 基础能力（默认启用，非 feature 开关）

| 能力 | SDK | 说明 |
|------|-----|------|
| 日志 | `pkg/base/log` | `bootstrap/base.go` 中 `log.Init`；业务用 `log.WithContext(ctx)` |
| 链路 | `pkg/base/trace` | 同上 `trace.Init` + `GinTraceMiddleware` 中间件 |
| 错误 | `pkg/base/errors` | 业务层 `kerr.WrapError(status, code, err)`；Controller 用 `GinBase.ErrorResponse` |
| 跨域 | `pkg/biz/middlewares` | `CorsMiddleware()`，默认放开所有来源，生产请传白名单 |

`user` 模块查询不存在的用户时返回 `kerr.WrapError(kerr.StatusNotFound, codes.UserNotFound, ...)`，由 Controller 统一输出 `{"code":40401,"msg":"..."}`。

可选能力（mysql / redis / metrics / pprof / http-client / traffic / llm）由 `go-infra-cli add|remove` 在 `internal/bootstrap/app.go` 的 `// FEATURE_*_START/END` 锚点内注入或移除，**不用 `features.*` 开关**。

## 配置

`internal/infra/config.App` 优先复用 SDK 的配置结构（`log` / `trace` / `mysql` / `redis` / `http_client` / `llm`），只额外定义编排层需要的 `app_name` / `server.address` / `grpc.address` / `traffic`。

环境覆盖靠 `env` 变量：

```bash
env=local go run ./cmd/http     # 叠加 configs/config.local.yml
env=prod  go run ./cmd/http     # 叠加 configs/config.prod.yml
go run ./cmd/http               # 不设 env，只用 config.yml
```

生产环境可对配置值做加密（`ENC(...)` + `CONFIG_ENCRYPT_KEY`）：

```bash
go-infra-cli keygen
go-infra-cli encrypt --key <hex> --value "mysecret"
```

约束：MySQL 表设计**不使用外键**，一致性由应用层保证；常规 JSON API 不负责前端跳转与 Cookie 写入（SSE / OAuth 等长连接与专用回调例外）。完整约束见 `.cursor/rules/` 与 `AGENTS.md`。

## 运行

```bash
go mod tidy
go run ./cmd/http   # :8080
go run ./cmd/grpc   # :9091
```

默认监听 `:8080`，示例接口：

- `GET /health` — 探活，刻意返回裸 JSON
- `POST /api/users`、`GET /api/users/:id`
- `GET /api/demo/ping?name=go-infra`
- `GET /ws` — WebSocket demo

```bash
curl "http://127.0.0.1:8080/api/demo/ping?name=go-infra"
grpcurl -plaintext 127.0.0.1:9091 demo.DemoService/Ping
```

`user` 的 `dao` 是内存实现，便于开箱即跑；真实项目替换为 MySQL/GORM 实现即可。

## Cursor 协作规范

`go-infra-cli init` 生成的项目默认包含：

- `AGENTS.md`：Agent 协作约定与分层边界
- `.cursor/rules/00-architecture.mdc`：分层架构、依赖方向与变更安全约束
- `.cursor/rules/10-go-sdk-first.mdc`：SDK 优先复用约束
- `.cursor/rules/11-go-infra-api.mdc`：go-infra 包速查（完整清单见 go-infra 仓库 `README.md` 的模块总表）
- `.cursor/rules/12-http-routing.mdc`：HTTP 路由约定

依赖统一走发布版 SDK（`github.com/liukunxin/go-infra`），生成的项目**不含任何 `replace`**，clone 到任意机器 `go build ./...` 即可。
