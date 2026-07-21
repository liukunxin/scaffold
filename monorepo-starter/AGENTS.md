# Agent Working Agreement (Monorepo)

This project uses a monorepo + event-driven + runtime-core architecture.

## Core Rules

1. Runtime is the single source of truth for state.
2. Domains are stateless plugins and communicate via event contracts.
3. Apps are transport entry layers only.
4. Reuse `github.com/liukunxin/go-infra` for shared infra capabilities whenever available.

> go-infra SDK 完整包清单与用法速查见 `.cursor/rules/11-go-infra-api.mdc`（编写 Go 文件时自动注入）。

## Layer Boundaries

- `apps/*`: HTTP/WebSocket/gRPC entry, request parsing, invoking runtime, response rendering.
- Gateway HTTP 约定见 `.cursor/rules/12-http-routing.mdc`：默认 `/api`（无版本）；跨模块演进走 event 契约，不靠 URL `/vN`。
- `internal/*`: runtime core (`event/session/snapshot/replay/wsproto/runtime`).
- `domains/*`: business plugins, no runtime implementation dependency.
- `pkg|packages/*`: infrastructure adapters only; no business decisions.

## Event Contract

All cross-module interactions must use:

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

