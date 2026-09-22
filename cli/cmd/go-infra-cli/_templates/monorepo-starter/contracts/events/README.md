# contracts/events/

事件载荷的 JSON Schema。事件是跨 Project、跨语言最常用的一种契约。

```text
events/
└── envelope.schema.json    # 事件信封：所有事件共用的外层结构
```

## 信封 + 载荷

所有事件套同一个**信封**，业务数据放在 `payload` 里：

```json
{
  "eventId": "9f1c...",
  "eventType": "demo.ping.completed.v1",
  "sessionId": "s-123",
  "seq": 1,
  "timestamp": "2026-09-22T09:30:00Z",
  "payload": { "message": "pong" }
}
```

字段与三个语言绑定的对应关系：

| 信封字段 | 含义 | Go | TypeScript | Python |
|---|---|---|---|---|
| `eventId` | 事件唯一 ID，幂等去重用 | `EventID` | `eventId` | `event_id` |
| `eventType` | `<域>.<动作>.v<版本>` | `EventType` | `eventType` | `event_type` |
| `sessionId` | 会话 / 关联 ID，串联一次业务过程 | `SessionID` | `sessionId` | `session_id` |
| `seq` | 同一会话内的序号，定序与补洞用 | `Seq` | `seq` | `seq` |
| `timestamp` | 产生时间（UTC，RFC3339） | `Timestamp` | `timestamp` | `timestamp` |
| `payload` | 业务数据，形状由 `eventType` 决定 | `Payload` | `payload` | `payload` |

> Python 侧字段名是 `snake_case`，`to_wire()` / `from_wire()` 负责与 `camelCase` 互转 ——
> **wire 上的名字永远是 camelCase**，三个语言必须一致，否则跨语言订阅解不开。

## 加一个事件

1. 在 `events/` 放 `<域>.<动作>.v<版本>.schema.json`（事件少时也可以只在本文件登记 payload 形状）。
2. `eventType` 按 `<域>.<动作>.v<版本>` 命名，并在各语言绑定里加常量
   （Go / TS / Python 各自一个 `EVENT_*`）。
3. 三个绑定各加一份 payload 类型，字段名与大小写规则保持一致，并更新本文件的表格。

## 不做什么

- 不做事件溯源（event sourcing）的基础设施 —— 这里只是**载荷形状**，投递与重试由具体服务决定。
- 不定义 topic / 队列名 —— 那是部署配置的事，放 `deploy/` 或服务自己的 `configs/`。
