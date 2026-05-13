# Agent Working Agreement

This repository follows a layered Go architecture based on shared SDK capabilities.

## Core Principles

1. Reuse existing SDK building blocks before introducing new infrastructure code.
2. Keep business logic in `internal/app` and avoid coupling it to low-level details.
3. Prefer small, composable interfaces and dependency injection at the bootstrap layer.
4. Keep generated and handwritten code style consistent with existing starter layout.

## Layer Boundaries

- `cmd/*`: entrypoints only (wire bootstrap, config, lifecycle).
- `internal/bootstrap`: application assembly and dependency wiring.
- `internal/app`: business use-cases, DTO/VO/RO, services, controller, dao contracts.
- `internal/infra`: adapters for config, persistence, observability, runtime integrations.
- `internal/route`: route registration and transport glue.

Business-facing packages should not directly depend on concrete external clients when an SDK abstraction already exists.

## SDK-First Policy

Before adding a new package or utility:

1. Search for existing SDK package that solves the same problem.
2. Use the SDK package unless there is a clear gap.
3. If there is a gap, document the reason and add a minimal wrapper in `internal/infra`.

## Delivery Checklist

- Reused SDK package where applicable.
- No cross-layer shortcut imports.
- Config keys and observability tags stay consistent.
- New files follow current naming and folder conventions.
