# monorepo-starter

跨语言、多 Project 的 Monorepo 标准骨架。

## 1. 什么是这个 Monorepo

**一个 Git Repository 管理多个相对独立的 Project。**

- **Monorepo 是 Repository 层级的概念。** 它只规定「多个 Project 住在同一个仓库里」，不规定它们用什么语言、什么构建系统、怎么发版。
- **Project 是项目边界。** `go.mod` / `package.json` / `pyproject.toml` 属于 Project 内部，不属于 Monorepo。
- Monorepo **不要求**所有 Project 同语言，**不要求**共用一套构建系统，也**不要求**把多个 Project 耦合成一个。

判断一个东西该不该是独立 Project，只看一条：**它是否需要被仓库外消费，或需要独立运行/独立发版。**

> 反面示范：把整个仓库设计成「一个大 Go 项目」——根目录放 `controller/ service/ dao/ model/`。
> 那等于把多个 Project 的内部实现混在一起，边界立刻消失。本项目**不采用**这种形态。

## 2. 目录职责

| 目录 | 是什么 | 放什么 | 不放什么 |
|---|---|---|---|
| `apps/` | 面向用户/客户端的应用 Project | `web`、`admin`、`mini-program`、BFF | 后端通用能力 |
| `services/` | 后端服务 Project | `gateway`、`search`、`recommendation`、`sdk-proxy` | 前端页面 |
| `packages/` | 跨 Project 复用的公共代码 | 按语言分组：`go/`、`typescript/`、`python/` | 业务代码、没有边界的大杂烩 |
| `contracts/` | 跨 Project 的通信契约（语言中立） | `proto/`、`openapi/`、`json-schema/`、`events/` | 某一语言的实现代码 |
| `tools/` | Monorepo 自身的工具链 | `codegen/`、`lint/`、`release/` | 业务逻辑 |
| `deploy/` | 跨 Project 共用的部署体系 | `docker/`、`k8s/`、`helm/`、`terraform/` | 单个 Project 专属的 Dockerfile |
| `docs/` | 跨 Project 的架构与约定文档 | 架构说明、开发手册、约定 | 单个 Project 的 README |

**Project 专属的东西留在 Project 内部**：`Dockerfile`、配置、迁移脚本、以及它的 `go.mod` / `package.json`。

`apps/` 与 `services/` 的区别是**语义**，不是结构：

- `apps/` 面向用户或客户端（页面、小程序、网关 BFF）。
- `services/` 是后端能力（搜索、推荐、代理）。
- 两者都是独立 Project，只是服务对象不同。不要为了"形式统一"把所有东西都塞进 `services/`。

## 3. Project 标准

### Go Project 遵循 `single-starter`

`apps/*` 与 `services/*` 下的 Go Project **统一按 `single-starter` 组织**，不要在 Monorepo 里另发明一套 Service 架构：

```text
services/gateway/
├── cmd/            # 按「运行形态」划分：http / grpc / worker ...
│   ├── http/
│   └── grpc/
├── configs/        # 配置 + 环境覆盖
├── internal/
│   ├── app/        # 按「业务域」垂直切片
│   │   └── demo/   #   controller / service / dao / dto / vo ...
│   ├── bootstrap/  # 启动编排与依赖装配
│   ├── infra/      # 技术基础设施适配（配置、外部依赖端口）
│   └── route/      # HTTP / gRPC / WS 协议层
├── Dockerfile      # 按需：取决于怎么部署，参考 services/gateway/Dockerfile
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

两条划分原则：

- `cmd` 按**运行形态**划分
- `internal/app` 按**业务域**划分（垂直切片）

而不是横向切：❌ `controller/ service/ dao/ model/`（那是把多个域的实现摊平，域边界会消失）。

能力基线（日志 / 链路 / 错误 / 配置）统一复用 `github.com/liukunxin/go-infra`，不重复造轮子。

> `go-infra-cli mono add` 生成的骨架**不含** `Dockerfile`：它取决于你打算怎么部署。
> 要加就照 `services/gateway/Dockerfile` 写，并且注意一个重要差异 ——
> 在 Monorepo 里**构建上下文是仓库根**（因为该 Project 的 `go.mod` 用相对路径
> `replace` 引用 `packages/`），不是在 Project 目录里执行 `docker build`。

### 非 Go Project

`apps/` 下的 React / Next / Taro，`services/` 下的 Python / Java —— 用**各自的官方脚手架**创建后放进对应目录。Monorepo 只约束两件事：**放在哪个目录**、**怎么被别的 Project 引用**。

```bash
npm create vite@latest apps/console -- --template react-ts   # 前端
uv init services/search-py                 # Python
```

## 4. Project 之间如何通信

```text
✅ 允许
Project → packages/            复用跨 Project 公共代码
Project → contracts/           共享通信契约
Project A ↔ contracts ↔ Project B

❌ 禁止
Project A → Project B/internal/xxx       直接 import 另一个 Project 的内部实现
```

Project 之间只通过 **HTTP / gRPC / Message-Event / Contract** 通信，不共享进程内代码。

共享的接口定义放 `contracts/`，它是**语言中立**的；各语言绑定放 `packages/<lang>/`：

```text
contracts/events/envelope.schema.json                       # 定义（语言中立）
packages/go/contracts/events/envelope.go                    # Go 绑定
packages/typescript/contracts/src/events.ts                 # TS 绑定
packages/python/contracts/src/contracts/events.py           # Python 绑定
```

## 5. 什么情况下创建新的 Project / packages

**创建 Project**（`apps/<name>` 或 `services/<name>`）：有一个能独立运行、独立部署、独立演进的东西。

```bash
# Go Project：按 single-starter 生成（自动写入 go.work）
go-infra-cli mono add service search
go-infra-cli mono add app bff

# 非 Go Project：用各自官方脚手架生成后放进 apps/ 或 services/
```

**创建 `packages/`**：只有当同一份代码**确实被 2 个以上 Project 复用、且边界稳定**时才提升上去。

> 默认策略：Project 内部代码留在 Project 内部。

- ❌ `packages/common/`、`packages/utils/`、`packages/business/` —— 没有明确边界的大杂烩。
- ❌ 把业务代码抽成 `packages/go/search/`、`packages/go/order/` —— 除非该业务能力确实是多个 Project 共享的独立领域能力。

## 6. 为什么 Go Project 遵循 `single-starter`

因为 Project 的**内部架构**与它**放在哪个仓库**是正交的两件事。

`single-starter` 已经定义了一套可用的单服务标准（`cmd` 按运行形态、`internal/app` 按业务域、bootstrap 显式装配、能力走 SDK）。Monorepo 只改变「多个 Project 共处一仓」，不该顺手把每个 Project 的内部结构也换一遍 —— 那会让同一家公司的 Go 服务出现两套写法。

所以：**仓库层用 Monorepo 规范，Project 层用 single-starter 规范。**

## 7. `go.work` 的作用

`go.work` 只负责**开发环境**下 Go 模块之间的联调：让本仓库的 Go Project 直接互相引用，不必先发版。

- **它不是 Monorepo 的定义。** 删掉 `go.work`，这个仓库依然是 Monorepo。
- 只列 **Go** 模块，不列前端 / Python Project。
- 全是 Go 可以用它；Go + Python + TS 混在一起也照样是 Monorepo。**不要因为有 `go.work` 就把整个仓库设计成 Go 专属结构。**

同仓 Go 模块之间另外用**相对路径 `replace`** 互相引用（见 `services/gateway/go.mod`）：

```text
replace <module-root>/packages/go/contracts => ../../packages/go/contracts
```

这样在**没有** `go.work` 的场景（CI 单模块构建、`go mod tidy`）也能离线解析，不必去公网拉一个不存在的伪版本。

`go.work` 里还有一行 `replace google.golang.org/genproto => ...`，那是历史包袱的兜底：
`go-infra v1.0.2` 还间接依赖拆分前的单体 `genproto`，它和 grpc 要的拆分模块同时进入 workspace 构建列表会
报 `ambiguous import`（单模块工程会被 `go mod tidy` 剪掉，workspace 不会剪）。`go.work` 的改动不会被
`go mod tidy` 覆盖，所以放这里最稳。go-infra 摘掉该依赖后这行即可删除。

## 快速开始

```bash
go work sync        # 收敛 workspace
make list           # 列出本仓库所有 Project
make check          # fmt + go-tidy + go-build + go-test + structure
```

`go.work` 的 workspace 根目录**不是** Go 模块，所以不能在根目录直接 `go build ./...`（会报 `directory prefix . does not contain modules`）。`make` 的目标会逐个进入模块目录执行。

跑示例服务：

```bash
cd services/gateway
go run ./cmd/http      # :8080
go run ./cmd/grpc      # :9091
```

```bash
curl "http://127.0.0.1:8080/api/demo/ping?name=go-infra"
grpcurl -plaintext 127.0.0.1:9091 demo.DemoService/Ping
```

跑示例前端（它会调上面这个接口）：

```bash
cd apps/web
npm install
npm run dev            # http://127.0.0.1:5173
```

不想装本地 Go / Node，也可以用容器把后端和共享依赖一起起：

```bash
docker compose -f deploy/docker/docker-compose.dev.yml up --build
```

完整走查（含「加一个新 Project」「改一次契约」）见 `docs/development/getting-started.md`。

## 目录

```text
monorepo-starter/
├── apps/
│   └── web/                 # 示范：Vite + TS 前端 Project
├── services/
│   └── gateway/             # 示范：Go Project，按 single-starter 组织
├── packages/
│   ├── go/contracts/        # 契约的 Go 绑定（独立 Go 模块）
│   ├── typescript/contracts/  # 契约的 TS 绑定
│   └── python/contracts/    # 契约的 Python 绑定
├── contracts/
│   ├── proto/demo/v1/       # gRPC / IDL（demo.proto）
│   ├── openapi/             # HTTP 契约（gateway.openapi.yaml）
│   ├── json-schema/         # 通用数据结构（按需）
│   └── events/              # 事件契约（envelope.schema.json）
├── tools/
│   ├── codegen/             # 契约代码生成（按需）
│   ├── lint/                # check-structure.py：目录约定检查
│   └── release/             # changelog / 版本（按需）
├── deploy/
│   ├── docker/              # docker-compose.dev.yml
│   ├── k8s/ helm/ terraform/   # 按需
├── docs/
│   ├── architecture/        # 按需
│   ├── development/         # getting-started.md
│   └── conventions/         # 按需
├── .github/workflows/       # CI
├── .cursor/rules/           # AI 协作规则
├── AGENTS.md
├── .gitignore
├── go.work
├── go.work.sum
├── Makefile
└── README.md
```

## 仓库里带了哪些示范

骨架不放假的业务代码，只放能跑通的**最小示范**，用来说明每个位置该怎么用：

| 示范 | 位置 | 说明 |
|---|---|---|
| Go 服务 Project | `services/gateway/` | single-starter 布局；`internal/app/demo` 演示 `controller -> service` 与 `Project -> packages/go/contracts` 这条允许的依赖；`internal/app/realtime` 演示 WebSocket |
| 前端 Project | `apps/web/` | Vite + TS；用 `file:` 引用 `packages/typescript/contracts`，通过 HTTP 调 gateway，**不 import 对方的内部代码** |
| 多语言契约绑定 | `contracts/events/envelope.schema.json` + `packages/{go,typescript,python}/contracts/` | 同一份事件信封在三种语言里的绑定，演示 `Project A ↔ contracts ↔ Project B` |
| HTTP 契约 | `contracts/openapi/gateway.openapi.yaml` | 与 `services/gateway/internal/route/init.go`、`apps/web/src/api.ts` 同源 |
| gRPC 契约 | `contracts/proto/demo/v1/demo.proto` | 表达接入 codegen 后的契约形态（当前 gateway 的 gRPC 示例是手工注册） |
| 共享依赖编排 | `deploy/docker/docker-compose.dev.yml` | gateway + redis，一条命令起本地环境 |
| 目录约定检查 | `tools/lint/check-structure.py` | 把「目录约定」变成 CI 能拦住的错误 |
| 本地走查 | `docs/development/getting-started.md` | 跑起来 → 加 Project → 改契约 → 提交前检查 |

## 相关文档

- `AGENTS.md` —— Agent 协作约定
- `.cursor/rules/` —— 架构边界、SDK 优先、HTTP 路由、契约规则的强制约束
- `docs/development/getting-started.md` —— 本地开发走查
- `services/gateway/README.md` —— 示例 Go Project 的说明
