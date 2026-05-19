# Agent Working Agreement

This repository follows a layered Go architecture based on shared SDK capabilities.

## Core Principles

1. Reuse existing SDK building blocks before introducing new infrastructure code.
2. Keep business logic in `internal/app` and avoid coupling it to low-level details.
3. Prefer small, composable interfaces and dependency injection at the bootstrap layer.
4. Keep generated and handwritten code style consistent with existing starter layout.

## Baseline SDK（默认启用，非 feature 开关）

以下能力在脚手架中**默认已接入**，新代码必须沿用，不要重复造轮子：

| 能力 | SDK 包 | 约定 |
|------|--------|------|
| 日志 | `pkg/base/log` | bootstrap 初始化；业务日志用 `log.WithContext(ctx)` |
| 链路 | `pkg/base/trace` | bootstrap + 中间件；不要在业务层自建 tracer |
| 错误 | `pkg/base/errors` | 业务层 `WrapError`；Controller 用 `GinBase.ErrorResponse` 统一响应 |

`init --features` / `go-infra-cli add` 按需安装可选基础设施（mysql/redis/metrics 等）；**不包含** log/trace/errors，且无 `features.*` 配置开关。

## Layer Boundaries

- `cmd/*`: entrypoints only (wire bootstrap, config, lifecycle).
- `internal/bootstrap`: application assembly and dependency wiring.
- `internal/app`: business use-cases, services, controller, dao contracts; object packages stay module-local by default.
- `internal/infra`: adapters for config, persistence, observability, runtime integrations.
- `internal/route`: route registration and transport glue.
- `internal/app/*/logic`: use only for cross-service/domain orchestration.
- For simple single-service pass-through, prefer `controller -> service` directly.
- Keep both patterns visible in starter examples (one module with logic, one module without logic).
- Avoid importing another module's RO/DTO/VO directly; if cross-project reuse is required, extract it into SDK instead of app-global object package.

Business-facing packages should not directly depend on concrete external clients when an SDK abstraction already exists.

## MySQL 设计约束

- **禁止使用外键**（`FOREIGN KEY`）。
- 引用关系、级联删除、一致性由应用层（service/dao）保证。
- 索引只为查询性能服务，不替代业务校验。

## 前后端边界

后端只提供**数据与业务结果**（JSON + 统一错误码），不负责前端交互编排。

**禁止**（除非需求明确且单独说明）：

- 用 `3xx` 重定向驱动页面跳转；
- 通过 `Set-Cookie` 写前端登录态/页面状态；
- 返回用于控制前端路由的 HTML/脚本片段。

**例外**（需窄化使用）：

- SSE、WebSocket 等长连接场景；
- OAuth 回调、文件下载等必须用重定向或 `Content-Disposition` 的专用接口。

## SDK-First Policy

Before adding a new package or utility:

1. Search for existing SDK package that solves the same problem.
2. Use the SDK package unless there is a clear gap.
3. If there is a gap, document the reason and add a minimal wrapper in `internal/infra`.

## Delivery Checklist

- Reused SDK package where applicable (especially log / errors).
- No cross-layer shortcut imports.
- Config keys and observability tags stay consistent.
- New files follow current naming and folder conventions.
- Every interface method has concise Chinese comments.
- Avoid adding method comments on concrete implementation methods unless specifically required.
- MySQL DDL without foreign keys.
- HTTP APIs do not control frontend navigation or cookies (except documented streaming/OAuth cases).
