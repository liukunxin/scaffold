# Agent Working Agreement

本仓库是分层 Go 服务骨架，能力基线来自共享 SDK。

## Core Principles

1. 先复用 SDK 已有能力，再考虑新增基础设施代码。
2. 业务逻辑留在 `internal/app`，不要与底层细节耦合。
3. 优先小接口 + bootstrap 层显式注入，不要全局状态。
4. 新代码风格与现有骨架保持一致。

## Layer Boundaries

- `cmd/*`：只做入口装配与生命周期。SDK 日志初始化**之前**允许用标准库 `log.Fatalf`，这是唯一例外（此时 `log.WithContext` 尚未可用）。
- `internal/bootstrap`：启动编排与依赖装配。进程共用的基线能力统一走 `base.go` 的 `loadBaseConfig`。
- `internal/app`：业务用例、service、controller、dao 契约；`ro/dto/vo/convert` 默认模块内私有。
- `internal/infra`：适配层（配置、外部依赖的端口与实现）。
- `internal/route`：传输层注册（HTTP 路由、gRPC 服务挂载），不写业务分支。
- `internal/app/*/logic`：**仅用于跨 service / 跨域的编排**；单 service 直通请用 `controller -> service`。
- 骨架同时保留两种形态：`user`（四层，logic 做真实编排）与 `demo`（三层，无 logic）。
- 不要直接 import 其它模块的 RO/DTO/VO；需要跨项目复用就抽到 SDK，而不是做 app 级公共对象包。

## SDK-First Policy

新增包或工具之前：

1. 先查 SDK 里有没有等价能力。
2. 有就直接用。
3. 确实缺，才在 `internal/infra` 加最小封装，并写明原因。

> SDK 包清单与用法速查见 `.cursor/rules/11-go-infra-api.mdc`（编写 Go 文件时自动注入）。

架构约束、MySQL 无外键、前后端边界、`Must Not` 清单统一维护在 `.cursor/rules/00-architecture.mdc`，此处不重复。

## Delivery Checklist

- 该用 SDK 的地方没有另起炉灶（尤其 log / errors）。
- 没有跨层捷径 import。
- 配置键与可观测性标签保持一致。
- 新文件遵循现有命名与目录约定。
- 接口方法有简洁中文注释；具体实现方法不加注释。
- MySQL DDL 无外键。
- HTTP 接口不驱动前端跳转、不写 Cookie（除已记录的长连接 / OAuth / 下载场景）。
