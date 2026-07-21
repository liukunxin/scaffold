# gateway

默认网关服务，提供可跑的参考实现：

- HTTP: `GET /health`、`GET /api/runtime/ping?name=go-infra`
- WebSocket: `GET /ws`（消息广播示例）
- gRPC: `gateway.RuntimeService/Ping`（protobuf-free 内置类型示例）

## 启动

```bash
# HTTP + WS
go run ./cmd

# gRPC
go run ./cmd/grpc
```

## 代码参考

- `internal/repository`: 配置型 repository 示例
- `internal/service`: runtime 编排与 event 调用示例
- `internal/handler`: Gin handler 示例
- `transport/http`: HTTP 路由与 middleware 集成
- `transport/ws`: go-infra websocket 集成
- `transport/grpc`: go-infra grpc 集成

## 设计说明

- handler 不直接写业务逻辑，只做请求转换和响应映射
- service 负责 app 编排（调用 runtime 与 domains），不做协议细节
- domain 通过 `api.Contract` 接入 runtime handler，避免 runtime 直接依赖 domain 实现细节

