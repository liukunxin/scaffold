# packages/go/

Go 共享包。每个子目录是一个**独立的 Go module**，开发期被根 `go.work` 串起来。

```text
packages/go/
└── contracts/     # 跨 Project 契约的 Go 绑定
```

## 为什么一个包一个 module

仓库根没有 `go.mod`（它是纯 `go.work` workspace），所以共享包不能靠「同一个模块里的相对目录」
被引用。独立成 module 之后：依赖显式写在 `go.mod` 里、可单独 `go test ./...`、可单独发版。

Project 引用它时，要**同时**写 `require` 和相对路径 `replace`：

```go.mod
require <模块根>/packages/go/contracts v0.0.0

replace <模块根>/packages/go/contracts => ../../packages/go/contracts
```

`replace` 是必需的：CI 里可能没有 `go.work`（单模块构建、`go mod tidy`），
没有它就会去公网拉一个不存在的版本然后失败。`go-infra-cli mono add` 生成的项目已包含这一对。

## 加一个共享包

```bash
mkdir -p packages/go/<name>
cd packages/go/<name>
go mod init <模块根>/packages/go/<name>
```

再把 `./packages/go/<name>` 加进根 `go.work` 的 `use (...)`。

> 手写即可。`mono add` 是给 `apps/` 与 `services/` 用的；共享包通常很小，而且需要人来判断
> 它的边界与命名，不适合自动生成。

## 现有内容

| 包 | 说明 |
|---|---|
| `contracts` | 事件信封（`fevents.Envelope`）的 Go 绑定。见 `contracts/README.md` |

## 约束

- 包名用**小写单词**，不用下划线、不用 `go-utils` 这类泛名。
- **禁止** import `apps/*` 或 `services/*` —— 依赖方向只能是 Project → packages。
- 导出 API 克制：实现细节放 `internal/`，只导出真正需要共享的部分。
- 改完在仓库根跑 `make go-tidy && make go-test`。
