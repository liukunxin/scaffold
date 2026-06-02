# Agent Working Agreement (Monorepo)

This project uses a monorepo + event-driven + runtime-core architecture.

## Core Rules

1. Runtime is the single source of truth for state.
2. Domains are stateless plugins and communicate via event contracts.
3. Apps are transport entry layers only.
4. Reuse `github.com/liukunxin/go-infra` for shared infra capabilities whenever available.

## Layer Boundaries

- `apps/*`: HTTP/WebSocket/gRPC entry, request parsing, invoking runtime, response rendering.
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

