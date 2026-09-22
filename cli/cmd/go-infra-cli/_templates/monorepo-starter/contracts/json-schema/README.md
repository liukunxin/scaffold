# contracts/json-schema/

通用 JSON 数据结构 —— 放**不属于 proto / openapi / events 任何一类**、又被多个 Project
或多种语言共享的形状。

```text
json-schema/
└── （当前为空）
```

## 什么该放这里

- 既不是 RPC、也不是 HTTP 接口、还不是事件的共享结构。例如：配置文件的 schema、
  导出报表的格式、共享的枚举字典。

## 什么不该放这里

| 想放的东西 | 应该去哪 |
|---|---|
| 事件载荷 | `events/` |
| HTTP 请求 / 响应 | `openapi/` |
| gRPC message | `proto/` |

> 目录空着是正常的。没有这种需求就别往里塞 —— 归错类比不归类更麻烦。
