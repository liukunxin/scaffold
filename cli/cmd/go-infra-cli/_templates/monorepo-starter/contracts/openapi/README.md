# contracts/openapi/

HTTP 接口契约。一个服务一份 `<服务>.openapi.yaml`。

```text
openapi/
└── gateway.openapi.yaml    # 示范：与 services/gateway/internal/route/init.go、apps/web/src/api.ts 同源
```

## 该放什么

- 对外 HTTP 接口的路径、方法、请求/响应体、状态码。
- 前端（`apps/*`）与后端服务之间的接口约定 —— 前端照这份文件写自己的 `src/api.ts`。

## 不该放什么

- 健康检查、`/metrics` 这类不构成对外约定的端点，简单提一句即可。
- 服务之间的 RPC —— 那是 `proto/`。

## 统一响应壳

本仓库的 HTTP 响应统一用 go-infra 的响应结构：

```json
{ "code": 0, "msg": "ok", "data": {}, "trace_id": "..." }
```

`code` 非 0 表示业务错误，`data` 可缺省，`trace_id` 用于串联日志。契约里每个接口的 `responses`
只描述 `data` 的形状，不用把壳重写一遍。

## 已知缺口

契约、后端路由、前端调用三者**目前没有自动一致性校验**，靠人守。改接口时三处一起改：

```text
openapi/gateway.openapi.yaml
services/gateway/internal/route/init.go
apps/web/src/api.ts
```

接口数量上去之后，值得在 `tools/codegen/` 里加一个「按契约生成前端 client」的脚本，
或者至少在 CI 里校验前端请求的路径都出现在契约中。
