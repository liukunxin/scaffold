# contracts/

**语言中立的接口定义**。跨 Project、跨语言的「约定」放这里；各语言的实现放 `packages/<lang>/`。

```text
contracts/
├── proto/          # gRPC / IDL（*.proto）
├── openapi/        # HTTP 接口（*.openapi.yaml）
├── events/         # 事件载荷的 JSON Schema
└── json-schema/    # 其它通用数据结构
```

## 三条边界

1. **这里只放「形状」，不放实现。** `*.proto` / `*.yaml` / `*.schema.json` 是定义；
   对应语言的绑定（struct / interface / dataclass）在 `packages/<lang>/*`。
2. **契约先改，代码后改。** 顺序永远是：改 `contracts/` → 各语言绑定跟着改 → 使用方最后改。
   反过来做，一定会出现「服务已经按新字段发消息、调用方还在按旧字段解析」的中间态。
3. **不 import 任何 Project 的代码。** 与 `packages/` 一样，只能被 Project 引用，不能反向依赖。

## 现有契约与它们的绑定

| 契约 | 语言中立定义 | Go 绑定 | TS 绑定 | Python 绑定 |
|---|---|---|---|---|
| 事件信封 | `events/envelope.schema.json` | `packages/go/contracts/events` | `packages/typescript/contracts` | `packages/python/contracts` |
| Demo HTTP 接口 | `openapi/gateway.openapi.yaml` | `services/gateway/internal/route/init.go` | `apps/web/src/api.ts` | — |
| Demo gRPC 接口 | `proto/demo/v1/demo.proto` | `services/gateway/internal/app/demo/grpc` | — | — |

> **改一份契约，同一行里的所有绑定一起改。** 每一行都是「同一份约定的多个投影」，
> 只改一边就会漂移。哪个文件与哪个实现同源，各文件头部都写了。

## 事件的命名与版本

事件的 `eventType` 必须满足：

```text
^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)*\.v[0-9]+$
```

即 `<域>.<动作>.v<主版本>`：

```text
demo.ping.completed.v1
order.created.v1
order.payment.failed.v2
```

- **主版本进名字**，不做「同名兼容」的隐式升级 —— 订阅方一眼看出该按哪版解析。
- 加**可选**字段不升版本；删字段、改字段语义、改必填性 = 升主版本。
- 三个绑定都带校验，写错在测试阶段就炸：Go 见 `events.Envelope`、TS 用 `isEnvelope()`、
  Python 用 `EVENT_TYPE_PATTERN`。

## 加一份新契约

1. 在对应子目录放定义文件（`proto/<域>/v1/<名>.proto`、`openapi/<服务>.openapi.yaml`、
   `events/<名>.schema.json`）。
2. 文件头部写清**与哪个实现 / 哪个调用方同源**（照现有文件的注释格式写）。
3. 在 `packages/<lang>/` 下补绑定，并回到本文件的表格里登记一行。
4. 使用方改成引用绑定，不要在 Project 里再手抄一份结构体。

## 代码生成

`proto/` 与 `openapi/` 的桩代码生成脚本将来放 `tools/codegen/`（目前为空）。
本仓库的示范契约规模很小，各语言绑定都是**手写**的；契约一多就该走生成，别再手抄。
