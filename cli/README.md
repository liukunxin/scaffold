# go-infra-cli

`go-infra-cli` 用于生成两种项目布局：

- `single-starter`：单服务项目（同时也是**Project 层的架构标准**）
- `monorepo-starter`：跨语言、多 Project 的 monorepo 骨架（`apps/ services/ packages/ contracts/ tools/ deploy/ docs/`）

模板直接编译进二进制，任意目录执行 `init` 都可用，无需 `--template`，也不依赖 CLI 的安装位置。

## 安装

```bash
go install github.com/liukunxin/scaffold/cli/cmd/go-infra-cli@latest
```

本地开发（推荐，覆盖旧 exe）：

```bash
cd scaffold/cli
go install ./cmd/go-infra-cli
```

生成的项目默认 `require github.com/liukunxin/go-infra v1.0.2`（发布版），不写 `replace`，
换台机器也能直接 `go mod tidy`。若确实要指向本机 checkout 的 SDK，显式加 `--use-local-sdk`。

## 通用参数（init）

- `--module`：Go module 名（默认=项目名）
- `--app-name`：配置中的 `app_name`（默认=项目名）
- `--layout`：`single|monorepo`（默认 `single`）
- `--output`：输出目录（默认当前目录）
- `--template`：覆盖模板目录（默认用内嵌模板，调试模板时才需要）
- `--force`：目标目录存在时覆盖
- `--skip-tidy`：跳过 `go mod tidy` 与生成物自检
- `--use-local-sdk`：写入 `replace` 指向本地 go-infra checkout（仅开发用）
- `--features` / `--scenes`：**仅对 `--layout single` 有效**，monorepo 传了会直接报错（见下）

> `.cursor/rules` 与 `AGENTS.md` 会随模板一起生成。
> 不加 `--skip-tidy` 时，`init` 结束前会 `go build ./...` 自检；失败则删除输出目录，不留半成品。
> monorepo 布局下是「逐模块 tidy + 逐模块 build」。

## Single 项目（single-starter）

### 初始化

```bash
# 默认全场景（http + grpc + ws）
go-infra-cli init myapp --module github.com/acme/myapp --layout single

# 只要 http
go-infra-cli init myapp --module github.com/acme/myapp --scenes http

# 初始化时按需安装能力
go-infra-cli init myapp \
  --module github.com/acme/myapp \
  --features mysql,redis,metrics,pprof,http-client,traffic,llm
```

### single 专属参数（仅 `init --layout single` 可用）

- `--features`：初始化时按需安装能力，不传则一项都不装；monorepo 的 Project 用 `add --dir` 安装（见下）
- `--scenes`：场景裁剪，支持 `http,grpc,ws`（默认全开，`http` 始终保留）；monorepo 的 Project 恒为全场景

### 能力增删

作用对象是「单个 Go Project」，因此 monorepo 里同样可用，只需 `--dir` 指到那个 Project（见下）。

```bash
go-infra-cli add redis --dir ./myapp
go-infra-cli remove redis --dir ./myapp
```

- 支持：`mysql`、`redis`、`metrics`、`pprof`、`http-client`、`traffic`、`llm`
- `llm` 是"覆盖文件型"能力：`add llm` 除了接线，还会把 `_features/llm/` 下的业务代码
  （`internal/app/llm/`、`internal/infra/ai/`、`internal/route/llm.go`）拷进项目；
  `remove llm` 会拆掉接线并回收这些文件，**用户改过的文件保留并提示**，不会静默删除
- 原则：只通过 bootstrap wiring 决定是否启用能力，不使用 `features.*` 配置开关
- 有文件改动时会自动 `go mod tidy` + `go build ./...` 自检（新增能力常带新依赖，不同步就是坏的），
  失败会直接报错退出；若项目本身已编译不过，这里也会一并暴露出来

## Monorepo 项目（monorepo-starter）

### 初始化

```bash
go-infra-cli init collab-platform \
  --module github.com/acme/collab-platform \
  --layout monorepo
```

生成的是 **Repository 层**骨架：`apps/ services/ packages/ contracts/ tools/ deploy/ docs/ .github/workflows/`，
外加 `go.work`、`Makefile`、`README.md`、`.gitignore`、`.cursor/rules/`。

模板里**自带一套能跑通的最小示范**（Go 服务、前端 Project、三语言契约绑定、共享依赖 compose、
目录约定检查）：清单见生成物根 `README.md` 的「仓库里带了哪些示范」一节。

仓库根**没有** `go.mod`（它是纯 workspace，每个 Project 各自是模块），所以不要在根目录敲 `go build ./...`
（会报 `directory prefix . does not contain modules`）；用 `make check`，它会逐个进模块目录执行。

### 增量扩展 Project

```bash
go-infra-cli mono add service search --dir ./collab-platform   # → services/search
go-infra-cli mono add app bff --dir ./collab-platform          # → apps/bff
```

- `app` → `apps/<name>`，`service` → `services/<name>`；两者都是**独立 Project**，结构完全一致（只是服务对象不同）。
- 生成的 Project 就是 **`single-starter` 的完整骨架**（`cmd/` + `configs/` + `internal/{app,bootstrap,infra,route}`），
  不是另写一份 monorepo 专用 Service 架构 —— 只把模块路径改对，并自动往 `go.work` 追加 `use`。
- 仓库根统一提供的 `.cursor`、`AGENTS.md`、`.gitignore` 不会重复生成到 Project 里。
- 命名约束：`^[a-z][a-z0-9-]*$`
- 生成后立刻 `go work sync` + 逐模块 `go mod tidy` + `go build ./...` 自检，失败会删除产物并报错。
- 生成的 Project 自带 `Dockerfile`，内容照 `services/gateway/Dockerfile` 生成（只把服务目录换掉）。
  注意在 monorepo 里**构建上下文是仓库根**，否则容器里看不见 `go.work` 与 `packages/`：

  ```bash
  docker build -f services/search/Dockerfile -t search:dev .
  ```

非 Go Project（前端 / Python / Java）不用这个命令，用各自官方脚手架生成后放进 `apps/` 或 `services/`。

### monorepo 下给具体 Project 增删能力

`init --layout monorepo` 不接受 `--features` / `--scenes`：能力属于 Project，而 Project 是
`mono add` 之后才存在的，monorepo 里没有唯一的 `internal/bootstrap/app.go` 可注入；
场景裁剪只服务 `--layout single`，`mono add` 生成的 Project 恒为全场景（http+grpc+ws）。
要对具体 Project 增删能力，用 `--dir` 指到它：

```bash
go-infra-cli add redis,traffic --dir services/search
go-infra-cli remove redis --dir services/search
```

具体 Project 的 `README.md` 里也会把这两条写好，照着敲即可。

## 配置加密

提供密钥生成、加密、解密三个命令，用于管理 YAML 配置中的敏感值。

### 生成密钥

```bash
go-infra-cli keygen
# 输出:
#   Key (hex): 3f8b1a...（64字符）
#
#   Set as environment variable:
#     export CONFIG_ENCRYPT_KEY=3f8b1a...
```

将输出的密钥存入环境变量（推荐通过 K8s Secret 或 CI/CD Secret 注入）。

### 加密配置值

```bash
# 从环境变量读取密钥（默认 CONFIG_ENCRYPT_KEY）
go-infra-cli encrypt --value="MyP@ssw0rd"

# 指定环境变量名
go-infra-cli encrypt --key-env=MY_KEY --value="secret"

# 直接传 hex 密钥（调试用）
go-infra-cli encrypt --key=3f8b1a... --value="secret"
```

输出 `ENC(...)` 格式的密文，粘贴到 YAML 配置中即可：

```yaml
mysql:
  password: "ENC(xxxxxxx)"
```

### 解密验证

```bash
go-infra-cli decrypt --value="ENC(xxxxxxx)"
```

### 运行时自动解密

应用侧使用 go-infra SDK：

```go
cfg := config.MustLoad[App](
    config.WithDecrypt(config.AESKeyFromEnv("CONFIG_ENCRYPT_KEY")),
)
```

不传 `WithDecrypt` 时 `ENC(...)` 被当作普通字符串，对现有项目零影响。

## 构建版本

```bash
go build -ldflags "-X main.version=v0.1.0" -o bin/go-infra-cli ./cmd/go-infra-cli
```
