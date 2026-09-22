# deploy/docker/

本地与联调用的 docker compose。

```text
docker/
└── docker-compose.dev.yml    # gateway + redis
```

## 怎么跑

```bash
docker compose -f deploy/docker/docker-compose.dev.yml up --build
docker compose -f deploy/docker/docker-compose.dev.yml down -v    # 停掉并清数据
```

## 三个容易踩的点

- **构建上下文是仓库根**（`context: ../..`），不是服务目录。服务的 `go.mod` 用相对路径
  `replace` 引用了 `packages/go/*`，只拷服务目录在 `go mod download` 阶段就解不开。
- 服务在容器里跑 `env=docker`，对应 `services/<名>/configs/config.docker.yml`。
- 容器里的中间件地址必须用 **compose 服务名**（`redis:6379`），不能用 `127.0.0.1`
  —— 容器里的 localhost 是容器自己。

## 新增一个服务

在 `services:` 下加一段：

```yaml
  order:
    build:
      context: ../..
      dockerfile: services/order/Dockerfile
    environment:
      - env=docker
    ports:
      - "8081:8080"
    depends_on:
      redis:
        condition: service_healthy
```

端口别和已有服务撞（宿主侧错开，容器内固定 8080 / 9091）。

## 不放这里的东西

| 想放的东西 | 应该去哪 |
|---|---|
| 生产编排 | `k8s/` 或 `helm/` |
| 单个服务的 `Dockerfile` | Project 内部（`services/<名>/Dockerfile`） |
