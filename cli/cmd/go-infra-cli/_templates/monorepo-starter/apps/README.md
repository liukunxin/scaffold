# apps/

面向用户或客户端的应用 Project。

```text
apps/
├── web/            # 浏览器端（示范：Vite + TS）
├── console/        # 管理后台（按需）
└── mini-program/   # 小程序 / Taro（按需）
```

## 什么该放这里

用户直接看到、直接使用的入口。与 `services/` 的区别是**服务对象**，不是结构 —— 两者都是独立
Project，只是 `apps/` 面向用户，`services/` 提供后端能力。

判断标准：**这个目录如果明天独立部署、独立发版，说得通吗？** 说得通才建。说不通就别建 ——
多一个 Project 的代价是独立的依赖、构建、版本和发布流水线。

## 怎么创建

- **Go Project**（BFF）：

  ```bash
  go-infra-cli mono add app bff
  ```

  按 `single-starter` 生成（`cmd/` + `internal/{app,bootstrap,infra,route}` + `configs/` + `Dockerfile`），
  并自动写入 `go.work`。

- **前端 / 小程序**：用各自的官方脚手架生成后放进来。

  ```bash
  npm create vite@latest apps/console -- --template react-ts
  npx @tarojs/cli init apps/mini-program
  ```

  > 本仓库自带的 `apps/web` 是**手写的示范**（Vite + TS），不是 CLI 生成的 —— 目的就是让人知道
  > 一个前端 Project 在这个仓库里长什么样。

## 现有 Project

| Project | 技术栈 | 说明 |
|---|---|---|
| `web` | Vite + TypeScript | **示范**：怎么用 `file:` 引用 `packages/typescript/contracts`、怎么通过 HTTP 调 `services/gateway`。见 `apps/web/README.md` |

## 约束

- 每个子目录是一个**独立 Project**，自带清单文件（`package.json` / `go.mod` / `pyproject.toml`）
  与自己的构建命令，**不在根目录统一构建**。
- 禁止 import 另一个 Project 的 `internal/`；跨 Project 走 `contracts/`（见根 README 第 4 节）。
- 前端之间共享的 UI / 工具代码提升到 `packages/typescript/`，不要在 `apps/` 里放公共代码。
- 目录名用 `kebab-case`，与 Project 名一致。
- 新增后跑一次 `make structure`，它会检查新目录是否符合约定。
