# gateway

示例 Go Project，演示 Monorepo 里一个 Project 应该长什么样。

**按 `single-starter` 规范组织**：`cmd` 按运行形态划分，`internal/app` 按业务域垂直切片。

## 目录

```text
services/gateway/
├── cmd/
│   ├── http/main.go           # HTTP 入口（只做装配与生命周期）
│   └── grpc/main.go           # gRPC 入口
├── configs/
│   ├── config.yml             # 基础配置
│   └── config.local.yml       # 环境覆盖（env=local）
├── internal/
│   ├── app/                   # 业务域（垂直切片）
│   │   ├── demo/              #   三层示例：controller -> service
│   │   │   ├── controller/  service/  dto/  vo/
│   │   │   └── grpc/          #   gRPC 适配
│   │   └── realtime/          #   WebSocket demo（/ws）
│   ├── bootstrap/             # 启动编排：base.go + app.go / grpc.go
│   ├── infra/                 # 技术适配层（配置加载）
│   └── route/                 # 传输层注册：HTTP 路由 + gRPC 服务挂载
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

## 跨 Project 的部分

这里有一个刻意的示范：`internal/app/demo/vo/ping.go` 直接复用 `packages/go/contracts/events.Envelope`，
它的定义在 `contracts/events/envelope.schema.json`。

```text
gateway ──► packages/go/contracts ◄── contracts/events/
```

Project 只能依赖 `packages/` 与 `contracts/`，**不能** import 另一个 Project 的 `internal/`。
需要和别的 Project 通信时走 HTTP / gRPC / Event，接口定义放 `contracts/`。

## 运行

```bash
# 从仓库根
go work sync
cd services/gateway

go run ./cmd/http      # :8080
go run ./cmd/grpc      # :9091
```

示例接口：

- `GET /health` —— 探活，刻意返回裸 JSON
- `GET /api/demo/ping?name=go-infra`
- `GET /ws` —— WebSocket 广播 demo
- gRPC `demo.DemoService/Ping`

```bash
curl "http://127.0.0.1:8080/api/demo/ping?name=go-infra"
grpcurl -plaintext 127.0.0.1:9091 demo.DemoService/Ping
```

## 能力增删

本 Project 的结构与 `single-starter` 一致，因此 `go-infra-cli add/remove` 可以直接作用在它上面：

```bash
cd services/gateway
go-infra-cli add redis          # 在 internal/bootstrap/app.go 的 FEATURE_* 锚点内注入
go-infra-cli remove redis
```

## 配置

环境覆盖靠 `env` 变量（与 single-starter 相同）：

```bash
env=local go run ./cmd/http     # 叠加 configs/config.local.yml
go run ./cmd/http               # 不设 env，只用 config.yml
```

## 与 `go-infra-cli mono add service` 的关系

本目录是**手写的最小示例**，用来演示规范；`mono add service <name>` 生成的是 `single-starter` 的**完整骨架**
（含 `internal/app/user` 四层示例、`internal/infra/notifier`、以及四个环境覆盖配置文件）。

两者的**目录形状与分层完全一致**，区别只在示例代码的多少。

## 部署

`Dockerfile` 的构建上下文是**仓库根目录**：

```bash
docker build -f services/gateway/Dockerfile -t gateway:dev .
```
