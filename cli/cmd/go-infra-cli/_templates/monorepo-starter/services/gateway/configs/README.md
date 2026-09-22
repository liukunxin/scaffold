# configs/

gateway 的配置，**分层叠加**：`config.yml` 打底，`config.<env>.yml` 覆盖同名键。

```text
configs/
├── config.yml          # 基础配置（server / grpc / log / trace / mysql / redis / http_client / traffic / llm）
├── config.local.yml    # local：本机开发
├── config.prod.yml     # prod：生产
└── config.docker.yml   # docker：容器内运行（Dockerfile 默认选它）
```

## 选环境

环境名从环境变量 `env` 读（见 `internal/infra/config/loader.go` 的 `WithEnvFrom("env")`）：

| `env` | 叠加文件 | 当前是否存在 |
|---|---|---|
| 不设置 / `local` / `dev` / `develop` | `config.local.yml` | 有 |
| `test` / `testing` | `config.test.yml` | 没有 |
| `gray` / `staging` | `config.gray.yml` | 没有 |
| `prod` / `production` / `release` | `config.prod.yml` | 有 |
| `docker` | `config.docker.yml` | 有 |

```bash
env=local go run ./cmd/http
docker run --rm -p 8080:8080 -e env=docker app:dev
```

> 环境名不在表里（比如写了 `test` 但没有 `config.test.yml`）会**静默退化**成 `local`，
> 不报错。加新环境时先确认这里认得它，并且对应的文件真的在。

## 现有覆盖项

| 文件 | 覆盖了什么 |
|---|---|
| `config.local.yml` | 日志改成 text 格式、trace 全采样（本机看日志方便） |
| `config.prod.yml` | 日志级别调低、trace 采样率 0.1 |
| `config.docker.yml` | redis 指向 compose 服务名 `redis:6379` |

## 约定

- **只写覆盖项。** 叠加语义下重复写没改动的键，会让「基础配置改了、环境文件没跟上」变成静默不一致。
- 密钥不进仓库，用环境变量或部署侧 secret 注入。
- 新增配置项要先加到 `internal/infra/config/app.go` 的结构体，否则 yaml 里的新字段会被忽略。
- 容器内的中间件地址用 compose 服务名，不能用 `127.0.0.1`。
