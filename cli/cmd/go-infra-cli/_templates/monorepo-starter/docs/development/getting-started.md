# 本地开发走查

从零到「跑起来 + 加一个 Project + 改一次契约」的完整路径。
命令前的注释标明了**在哪个目录执行**，这是 Monorepo 里最容易踩的坑。

## 0. 前置

| 需要 | 用在哪 |
|---|---|
| Go 1.25+ | `apps/*`、`services/*`、`packages/go/*` |
| Node 20+ / npm | `apps/web` |
| Python 3.11+ | `packages/python/*`、`tools/lint/*` |
| GNU make（可选） | 根目录的编排目标（`make check` 等） |

## 1. 先看清仓库里有什么

```bash
# 仓库根
make list          # 列出所有 Go Project（读 go.work 覆盖不到的目录，靠通配）
ls apps services packages contracts
```

没有 `make` 的话直接用等价命令：

```bash
ls apps/*/go.mod services/*/go.mod packages/*/*/go.mod
```

## 2. 收敛 Go 依赖

```bash
# 仓库根
go work sync
```

`go.work` 只服务于**开发期**的模块联调。根目录**不是** Go 模块，所以：

```bash
go build ./...      # ❌ 报 directory prefix . does not contain modules
```

要构建就进到具体模块，或用 `make go-build`。

## 3. 跑起来

```bash
# 终端 1：起后端
cd services/gateway
go run ./cmd/http          # :8080
# go run ./cmd/grpc        # :9091

curl "http://127.0.0.1:8080/health"
curl "http://127.0.0.1:8080/api/demo/ping?name=go-infra"
```

```bash
# 终端 2：起前端
cd apps/web
npm install
npm run dev                # http://127.0.0.1:5173
```

前端页面上点按钮，会走 `/api/demo/ping`（由 vite 反代到 :8080）。
这条链路的含义是：**前端 → HTTP → gateway → 契约类型**，中间没有任何跨 Project 的代码 import。

也可以用容器把后端和依赖一起起：

```bash
# 仓库根
docker compose -f deploy/docker/docker-compose.dev.yml up --build
```

## 4. 加一个新的 Go Project

```bash
# 仓库根
go-infra-cli mono add service order
go-infra-cli mono add app bff
```

生成结果：

- 目录按 `single-starter` 布局（`cmd/`、`internal/{app,bootstrap,infra,route}`、`configs/`）。
- 模块路径自动改成 `<你的模块根>/services/order`。
- 自动追加到 `go.work` 的 `use (...)`。
- **不含 `Dockerfile`**：它取决于你怎么部署。要加就照 `services/gateway/Dockerfile` 写，
  并注意在 Monorepo 里**构建上下文是仓库根**（该 Project 的 `go.mod` 用相对路径 `replace` 引用 `packages/`）：

  ```bash
  # 仓库根
  docker build -f services/order/Dockerfile -t order:dev .
  ```

加完接着做：

```bash
go work sync
make list                  # 应该能看到新 Project
make check                 # 或者只在它目录里 go build ./... && go test ./...
```

非 Go 的 Project 用各自脚手架，然后放进 `apps/` 或 `services/`：

```bash
npm create vite@latest apps/console -- --template react-ts
uv init services/search-py
```

## 5. 改一次契约（最重要的一条路径）

以 `contracts/events/envelope.schema.json` 为例，正确顺序是：

1. 改定义：`contracts/events/envelope.schema.json`（**只允许向后兼容**：加字段可以，改语义、删字段不行）
2. 改各语言绑定，**同一次提交**：
   - `packages/go/contracts/events/envelope.go`
   - `packages/typescript/contracts/src/events.ts`
   - `packages/python/contracts/src/contracts/events.py`
3. 改用到它的实现（例如 `services/gateway/internal/app/demo/vo/ping.go`）
4. 跑 `make check`（含 `tools/lint/check-structure.py`）

HTTP 契约同理，定义在 `contracts/openapi/gateway.openapi.yaml`，实现里的路由在
`services/gateway/internal/route/init.go`，前端调用在 `apps/web/src/api.ts`。

> 为什么要手工同步三份绑定？因为骨架里没有引入 codegen 依赖，
> 先用「契约 + 手写绑定」把边界立住。接入 `buf` / `openapi-generator` 之后，
> 第 2 步应变成 `make -C tools/codegen generate`，见 `tools/README.md`。

## 6. 提交前

```bash
# 仓库根
make check
```

它等价于 `fmt + go-tidy + go-build + go-test + structure`。
没有 `make` 时，最低要求是：每个改动过的 Go 模块内 `go build ./... && go test ./... && gofmt -l .`；
另外跑一次 `python tools/lint/check-structure.py`。

## 常见问题

**根目录 `go build ./...` 报错**
正常。workspace 根不是模块，见第 2 节。

**`ambiguous import: google.golang.org/genproto/...`**
`go.work` 里有一行 `replace google.golang.org/genproto => ...`，是给 `go-infra v1.0.2` 的间接依赖兜底
（它引了拆分前的单体 genproto，与 grpc 要的 `genproto/googleapis/rpc` 冲突，单模块工程会被 `go mod tidy`
剪掉、workspace 不会）。不要删这行。go-infra 摘掉该依赖后即可移除。详见根 `README.md` 第 7 节。

**`make: command not found`**
Windows 上装 GNU make（或用 WSL / Git Bash）。上面的每个目标都有等价的手工命令。

**`npm install` 之后 `@repo/contracts` 找不到类型**
它是以 `file:` 依赖软链进来的，确认 `packages/typescript/contracts/src/index.ts` 存在，
以及 `apps/web/package.json` 里的相对路径没被改错。

**新增 Project 后 CI 提示结构不合格**
`tools/lint/check-structure.py` 会检查 Go Project 是否是 single-starter 布局（`cmd/`、`internal/*`、
`configs/`）。按提示补齐，或者用 `go-infra-cli mono add` 重新生成。
缺 `Dockerfile` 只会提示、不会失败。
