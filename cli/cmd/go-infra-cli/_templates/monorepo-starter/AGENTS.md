# Agent Working Agreement (Monorepo)

本仓库是一个**跨语言、多 Project 的 Monorepo**。能力基线来自共享 SDK `github.com/liukunxin/go-infra`。

## Core Principles

1. **先分清层级**：改动落在「Repository 层」还是「Project 层」？两者规范不同。
2. **先复用 SDK**，再考虑新增基础设施代码。
3. **不跨 Project 共享进程内代码**；跨 Project 走契约。
4. 新代码风格与所在 Project 的既有骨架保持一致。

## 两层规范

- **Repository 层（Monorepo 规范）**：`apps/` `services/` `packages/` `contracts/` `tools/` `deploy/` `docs/` 的职责划分与依赖方向。
- **Project 层（single-starter 规范）**：`cmd/` 按运行形态、`internal/app/` 按业务域、bootstrap 显式装配。

> 不要把这个仓库当成"一个大 Go 项目"。根目录不允许出现 `controller/` `service/` `dao/` `model/` 这类横向分层目录。

## 依赖边界

```text
允许：Project → packages/        Project → contracts/        Project A ↔ contracts ↔ Project B
禁止：Project A → Project B/internal/xxx
```

- 跨 Project 只通过 HTTP / gRPC / Message-Event / Contract 通信。
- `contracts/` 放定义（语言中立），`packages/<lang>/` 放各语言绑定。
- `packages/` 的准入门槛：**真的被 2 个以上 Project 复用且边界稳定**。默认把代码留在 Project 内部。

## Project 内部分层（Go Project）

- `cmd/*`：只做入口装配与生命周期。SDK 日志初始化**之前**允许标准库 `log.Fatalf`，这是唯一例外。
- `internal/bootstrap`：启动编排与依赖装配；进程共用的基线走 `base.go`。
- `internal/app`：业务用例；`ro/dto/vo/convert` 默认模块内私有。
- `internal/infra`：适配层（配置、外部依赖的端口与实现）。
- `internal/route`：传输层注册（HTTP 路由、gRPC 服务挂载），不写业务分支。
- `internal/app/*/logic`：**仅用于跨 service / 跨域编排**；单 service 直通用 `controller -> service`。

## SDK-First Policy

新增包或工具之前：先查 SDK 有没有 → 有就用 → 确实缺，才在 `internal/infra` 或 `packages/` 加最小封装并写明原因。

> SDK 包清单与用法速查见 `.cursor/rules/11-go-infra-api.mdc`（编写 Go 文件时自动注入）。

架构边界、契约演进、HTTP 约定统一维护在 `.cursor/rules/00-architecture.mdc`、`20-contracts.mdc`、`12-http-routing.mdc`，此处不重复。

## go.work

`go.work` 只管开发环境的 Go 模块联调，**不是** Monorepo 的定义。它的 workspace 根目录不是模块，
根目录不要直接 `go build ./...`，用 `make` 的目标（见 `Makefile`）。

## Delivery Checklist

- 改动放对了层级（Repository 还是 Project）？
- 该用 SDK 的地方没有另起炉灶（尤其 log / errors）？
- 没有跨 Project 的捷径 import（尤其对方的 `internal/`）？
- 契约改了的话，`packages/*` 的绑定同步更新了吗？
- 配置键与可观测性标签保持一致？
- 接口方法有简洁中文注释；具体实现方法不加注释。
- MySQL DDL 无外键。
- 新增/改动 Project 后，`go.work` 与 `Makefile` 的模块列表能覆盖到（`make list` 能看到）？
- 跑过 `make structure`（`tools/lint/check-structure.py`）？新 Project 必须符合本文件里的布局约定。
- 契约变了的话，`contracts/` 的定义与三份绑定（Go / TS / Python）是否在同一次提交里同步了？
