# packages/typescript/contracts

`contracts/events/envelope.schema.json` 的 **TypeScript 绑定**（跨 Project 事件信封）。

三份绑定共用同一份契约：

| 语言 | 位置 |
|---|---|
| 定义（语言中立） | `contracts/events/envelope.schema.json` |
| Go | `packages/go/contracts/events/envelope.go` |
| TypeScript | 本包 |
| Python | `packages/python/contracts/src/contracts/events.py` |

## 用法

```ts
import { newEnvelope, isEnvelope, type Envelope } from "@repo/contracts";

const env: Envelope = newEnvelope("evt-1", "demo.ping.completed.v1", { message: "pong" });
if (isEnvelope(incoming)) {
  // incoming 已收窄为 Envelope
}
```

`eventType` 的格式约束在本包里做了一次校验（`EVENT_TYPE_PATTERN`），
和 Go 侧一样，属于**契约的本地防御**——不要在这里发明新的事件类型命名规则，规则在
`contracts/events/envelope.schema.json` 与 `.cursor/rules/20-contracts.mdc`。

## 怎么被别的 Project 引用

本包不发布到 registry。`apps/web` 用**相对路径**引用它：

```json
{
  "dependencies": {
    "@repo/contracts": "file:../../packages/typescript/contracts"
  }
}
```

不需要额外的 workspace 配置，`npm install` 会把它作为软链接装进 `node_modules`。

> 包名里的 `@repo` 是占位 scope。改成你们公司的 scope（比如 `@acme`）后，
> 记得同步改引用方的 `dependencies`。
