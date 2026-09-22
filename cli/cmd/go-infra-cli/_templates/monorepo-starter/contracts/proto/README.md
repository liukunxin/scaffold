# contracts/proto/

gRPC / IDL 定义。按 `<域>/v<版本>/<名>.proto` 组织。

```text
proto/
└── demo/v1/demo.proto     # 示范：与 services/gateway/internal/app/demo/grpc 同源
```

## 该放什么

- 服务之间的 **gRPC 接口**（service / rpc / message）。
- 需要跨语言共享、且希望用 protobuf 编码的结构。

## 不该放什么

- 只在一个 Project 内部用的结构 —— 直接写 Go struct。
- HTTP 的请求 / 响应形状 —— 那是 `openapi/` 的事。

## 版本与兼容

- 目录带 `v1` / `v2`，与文件里的 `package demo.v1;` 保持一致。
- **不兼容改动必须开新目录**，不要原地改字段号 —— 线上还有旧客户端在按老定义解析。
- 字段号一旦发布就不复用；删字段用 `reserved` 占位。
- 纯新增字段给默认值即可，不用升版本。

## 生成

生成脚本将来放 `tools/codegen/`。目前示范契约只有一个方法，Go 侧是**手写 ServiceDesc**
（见 `services/gateway/internal/bootstrap/grpc.go`）；接口一多就换成 `protoc` 生成，别继续手写。
