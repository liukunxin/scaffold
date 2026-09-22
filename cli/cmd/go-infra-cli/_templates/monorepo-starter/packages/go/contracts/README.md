# packages/go/contracts

跨 Project 契约的 **Go 绑定**。定义在 `contracts/`，这里只是它在 Go 侧的落地。

独立 Go 模块，模块路径 `<module-root>/packages/go/contracts`，由根 `go.work` 纳管。

## 包

| 包 | 对应契约 | 说明 |
|---|---|---|
| `events` | `contracts/events/envelope.schema.json` | 跨 Project 事件的统一信封 |

## 被 Project 引用

在 Project 的 `go.mod` 里同时写 `require` 与相对路径 `replace`，这样没有 `go.work` 时（CI 单模块构建、`go mod tidy`）也能离线解析：

```text
require <module-root>/packages/go/contracts v0.0.0
replace <module-root>/packages/go/contracts => ../../packages/go/contracts
```

示例见 `services/gateway/go.mod`。

## 约束

- **不要手写与 `contracts/` 不一致的结构。** 契约一变，这里必须跟着变；能生成就交给 `tools/codegen`。
- 不要在这里放业务逻辑；它只描述"跨 Project 长什么样"。
- 只加字段、不改语义、不删字段 —— 消费方可能还没升级。
