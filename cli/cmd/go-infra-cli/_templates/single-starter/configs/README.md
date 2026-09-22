# configs/

配置**分层叠加**：`config.yml` 打底，`config.<env>.yml` 覆盖同名项。

```text
configs/
├── config.yml          # 基础配置，所有环境共用
├── config.local.yml    # local
├── config.test.yml     # test
├── config.gray.yml     # gray
├── config.prod.yml     # prod
└── config.docker.yml   # docker（容器内运行，被 Dockerfile 默认选中）
```

## 选环境

环境名从环境变量 `env` 读（见 `internal/infra/config/loader.go` 的 `WithEnvFrom("env")`），别名映射：

| `env` 取值 | 叠加的文件 |
|---|---|
| 不设置 / `local` / `dev` / `develop` / `development` | `config.local.yml` |
| `test` / `testing` | `config.test.yml` |
| `gray` / `staging` | `config.gray.yml` |
| `prod` / `production` / `release` | `config.prod.yml` |
| `docker` | `config.docker.yml` |

```bash
env=local go run ./cmd/http
env=prod  go run ./cmd/http
```

容器里由 `Dockerfile` 的 `ENV env=docker` 设定，也可以在启动时改：

```bash
docker run --rm -p 8080:8080 -e env=gray myapp:dev
```

## 加一个新环境

复制 `config.local.yml` 成 `config.<新环境>.yml`，再用 `env=<新环境>` 启动即可，不用改代码。
但新环境名要落在上面那张表里 —— 不认识的取值会退化成 `local`，**不报错**，容易误以为生效了。

## 约定

- **只写覆盖项。** 叠加语义下重复写没改动的项，会让「基础配置改了但环境文件没跟上」变成静默不一致。
- **密钥不进仓库。** 用环境变量或部署侧 secret 注入。
- **新增配置项先改结构体** `internal/infra/config/app.go`，否则 yaml 里的新字段不会生效。
- 配置校验是开着的（`WithValidate` / `WithTagValidation`），字段写错会在启动时直接报错，而不是静默用零值。
- `config.docker.yml` 里的服务地址用 **compose 服务名**（如 `redis:6379`），不能用 `127.0.0.1` —— 容器里的
  localhost 是容器自己，不是宿主。
