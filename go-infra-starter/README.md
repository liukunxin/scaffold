# go-infra-starter

一个可单独提取的新项目骨架，目标是：

- 保留 `controller -> logic -> service -> dao` 四层风格
- 增加 `internal/infra` 技术适配层，隔离外部组件细节
- 使用 `go-infra` 完成 config/log/trace/mysql/redis/metrics 初始化
- 通过 `bootstrap` 统一启动编排，保证入口清晰

## 目录

```text
go-infra-starter/
├─ cmd/http/main.go
├─ configs/
├─ internal/
│  ├─ app/                          # 业务域（垂直切片）
│  │  └─ user/
│  │     ├─ controller/
│  │     ├─ logic/
│  │     ├─ service/
│  │     ├─ dao/
│  │     ├─ model/
│  │     ├─ ro/
│  │     ├─ dto/
│  │     ├─ vo/
│  │     └─ convert/
│  ├─ bootstrap/
│  ├─ infra/                        # 技术适配层
│  │  ├─ config/
│  │  ├─ observability/
│  │  ├─ persistence/
│  │  └─ runtime/
│  └─ route/
└─ go.mod
```

## 分层职责

- `controller`: 参数绑定、鉴权上下文读取、统一响应
- `logic`: 用例编排、跨 service 聚合、事务边界入口
- `service`: 业务规则、状态流转、外部能力调用
- `dao`: 数据访问与查询封装，不承载业务决策
- `infra`: 技术细节适配（配置、日志、追踪、存储、运行时）

## 快速运行

1. 复制该目录到你的新仓库根目录
2. 在新仓库执行：

```bash
go mod tidy
go run ./cmd/http
```

默认监听 `:8080`，示例接口：

- `GET /health`
- `POST /api/v1/users`
- `GET /api/v1/users/:id`

> 说明：示例 `user dao` 采用内存实现，便于开箱即跑。真实项目中将 `dao` 替换为 MySQL/GORM 实现即可。

