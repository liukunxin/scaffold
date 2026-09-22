# services/

后端服务 Project。

```text
services/
├── gateway/         # 示范：Go Project，按 single-starter 组织
├── order/           # 按需
└── search/          # 按需
```

## 什么该放这里

提供后端能力的独立服务：网关、检索、推荐、代理、任务消费者、定时任务。

判断标准：**有独立的部署单元与伸缩诉求吗？** 如果两个功能永远一起部署、一起扩容、共享同一个
数据库事务边界，那它们应该是**一个** Project 里的两个业务域（`internal/app/<域>/`），而不是两个
`services/` —— 拆错的代价是分布式事务和跨服务联调，比目录不整齐大得多。

## 怎么创建

```bash
# Go Project：按 single-starter 生成 + 自动写入 go.work
go-infra-cli mono add service order

# 非 Go Project：用官方脚手架生成后放进来
uv init services/search-py
```

Go Project 生成的目录：

```text
services/order/
├── cmd/                 # 按运行形态划分（http / grpc），一个形态一个 main
├── configs/             # config.yml 打底 + config.<env>.yml 覆盖
├── internal/
│   ├── app/             # 按业务域垂直切片（不是 controller/service/dao 横切）
│   ├── bootstrap/       # 启动编排
│   ├── infra/           # 技术适配（config、中间件、外部客户端）
│   └── route/           # 协议层：HTTP 路由、gRPC 注册
├── Dockerfile           # 构建上下文 = 仓库根，见文件内注释
├── Makefile
├── go.mod
└── go.sum
```

## 本地怎么跑

```bash
cd services/order
make run            # HTTP，:8080
make run-grpc       # gRPC，:9091
make test
```

出镜像（注意 `-f` 指到具体服务，但**上下文是仓库根**）：

```bash
docker build -f services/order/Dockerfile -t order:dev .
docker run --rm -p 8080:8080 -e env=docker order:dev
```

## 约束

- 每个子目录是一个**独立 Project**：独立构建、独立部署、独立发版。
- 禁止 `services/a` 直接 import `services/b/internal/...`；跨 Project 走 HTTP / gRPC /
  Message-Event，接口定义放 `contracts/`。
- 能力基线统一复用 `github.com/liukunxin/go-infra`，不各自造轮子。
- Project 专用的 `Dockerfile`、配置、部署细节留在 Project 内部；跨 Project 共用的放 `deploy/`。
- 服务之间不要共享数据库表；需要共享的数据通过契约交换。

> 只有 `gateway/` 是脚手架自带的示范，用来说明规范；其余目录按需创建，不预置假业务。
