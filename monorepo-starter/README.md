# monorepo-starter

基于 `github.com/liukunxin/go-infra` 的 monorepo 最小骨架（MVP）。

## 目标

- 保留单仓多服务（apps）+ 领域插件（domains）+ 平台 runtime（internal）结构
- 不重复造基础能力轮子，日志/追踪/错误等统一复用 `go-infra`
- 用 `go.work` 管理多模块联调

## 边界约束（推荐）

- `apps/*`: 负责协议适配与编排，不承载领域规则
- `domains/*`: 负责领域规则，不依赖 `apps/*` 与具体传输实现
- `internal/*`: 负责 runtime/event/session/snapshot 等平台内核
- 依赖方向：`apps -> domains/api + internal`，`domains -> domains/api`，禁止反向依赖

## 架构落地规范

### 1) 分层职责

- `transport/*`：HTTP/gRPC/WS 协议层，只做入参解析、出参映射、状态码处理
- `internal/handler`：应用层 handler，不写领域规则，只调用 app service
- `internal/service`：应用编排层，负责 runtime 事件编排、聚合多个 domain
- `domains/*/service`：领域规则层，沉淀业务规则与校验
- `domains/*/runtime`：runtime 适配层，通过 `domains/*/api.Contract` 挂接 domain

### 2) 允许/禁止依赖

- 允许：`apps/gateway/internal/service -> domains/*/api + internal/runtime`
- 允许：`domains/*/runtime -> domains/*/api + internal/event`
- 禁止：`domains/*` 直接 import `apps/*`
- 禁止：`domains/*/service` 直接依赖 HTTP/gRPC/WS 框架
- 禁止：`internal/*` 反向依赖 `apps/*` 或某个具体 domain 实现

### 3) 与 single-starter 的关系

- 对齐：启动流程（env/log/trace）、中间件接入方式、错误处理风格
- 不强行对齐：目录组织和 feature 开关模型（monorepo 以边界治理优先）
- 原则：优先保证边界清晰和演进稳定，不为了“看起来一致”而牺牲可维护性

### 4) 开发检查清单

- 新增 API 前，先定义 `domains/<name>/api` 的请求/响应与事件名
- 先写 domain `service`，再写 `runtime handler`，最后接到 app service/transport
- transport 只做协议转换，不直接 new domain service
- 跨层传递统一走 `event.Envelope`，避免临时结构体横飞
- 基础能力统一复用 `github.com/liukunxin/go-infra`，不重复造轮子

## 目录

```text
monorepo-starter/
├─ apps/
│  └─ gateway/
├─ domains/
│  └─ domain-demo/
├─ internal/
│  ├─ event/
│  ├─ session/
│  ├─ snapshot/
│  ├─ replay/
│  ├─ wsproto/
│  └─ runtime/
├─ configs/
├─ deploy/
├─ scripts/
├─ docs/
├─ .cursor/rules/
├─ AGENTS.md
├─ go.work
└─ Makefile
```

## 快速开始

```bash
go work sync
cd apps/gateway
go mod tidy

# HTTP + WS
go run ./cmd

# gRPC
go run ./cmd/grpc
```

默认示例能力：

- HTTP: `GET /health`
- Runtime: `GET /api/v1/runtime/ping?name=go-infra`
- WebSocket: `GET /ws`（广播示例）
- gRPC: `gateway.RuntimeService/Ping`

## Event Envelope

跨模块通信统一使用：

```json
{
  "eventId": "string",
  "eventType": "string",
  "sessionId": "string",
  "seq": "int64",
  "timestamp": "int64",
  "payload": {}
}
```

