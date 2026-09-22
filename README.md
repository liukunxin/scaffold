# scaffold

go-infra 官方脚手架仓库：一条命令生成按 **single-starter**（单服务）或 **monorepo-starter**（跨语言、多 Project）
规范组织的项目骨架。生成物自带可跑的示范代码，拿到的不是空壳。

```text
scaffold/
├── cli/                                  # 脚手架工具（Go module: github.com/liukunxin/scaffold/cli）
│   ├── cmd/go-infra-cli/
│   │   ├── _templates/                   # 随 CLI 编译进二进制的项目模板
│   │   │   ├── single-starter/           # 单服务项目 = Project 层的架构标准
│   │   │   └── monorepo-starter/         # 跨语言多 Project = Repository 层的标准
│   │   └── _features/                    # 可选能力覆盖文件（add <feature> 时注入）
│   ├── Makefile                          # cli-build / cli-install / cli-version
│   └── README.md                         # CLI 用法（安装、init、add、mono）
├── scripts/smoke.sh                      # 渲染冒烟（CI 与本地跑同一份）
├── .github/workflows/ci.yml              # CI
├── AGENTS.md                             # 改本仓库的约束（人与 AI 都看这份）
├── LICENSE                               # Apache-2.0
├── .editorconfig                         # 编辑器侧固化编码 / 换行
├── .gitattributes                        # 固化换行（模板必须 LF）
└── README.md                             # 本文件
```

## 快速开始

```bash
git clone https://github.com/liukunxin/scaffold.git
cd scaffold/cli
make cli-build        # 产物：cli/bin/go-infra-cli
# 或安装到 $GOBIN
make cli-install
```

```bash
# 单服务项目
go-infra-cli init myapp --module github.com/acme/myapp --layout single

# 跨语言、多 Project 的 Monorepo
go-infra-cli init collab-platform --module github.com/acme/collab-platform --layout monorepo
cd collab-platform
go-infra-cli mono add service order
go-infra-cli mono add app bff
```

参数、能力增删（`--features` / `--scenes`）与各命令的完整说明见 `cli/README.md`。

## 两个模板的关系

- **Repository 层**（仓库怎么分类、多个 Project 怎么共处、怎么互相引用）→ 由 `monorepo-starter` 定义。
- **Project 层**（一个服务内部怎么分层）→ 只有 `single-starter` 一个来源。

`monorepo-starter` 不在内部另发明一套服务架构：`apps/*`、`services/*` 下的 Go Project 一律按
`single-starter` 组织，`mono add app|service` 就是**直接拷 `single-starter` 再改模块路径**。
所以 Project 层的改动只需要改 `single-starter` 一处。

完整规则（含「7 个问题」的回答）见 `cli/cmd/go-infra-cli/_templates/monorepo-starter/README.md`。
模板内自带一套可跑示范（Go 服务 / Vite 前端 / TS 与 Python 契约绑定 / 结构检查 / docker-compose），
清单见该 README 的「示范清单」一节。

## 模板为什么放在 `_templates/` 下

不是随手取的名，是被 Go 工具链的三条规则逼出来的：

1. 模板里全是 `.go` 文件，目录名必须以 `_` 开头（`_templates`、`_features`），否则 `go build ./...`
   会把模板当源码编译并直接报错（实测 `templates/` → `package nonexistent-module/internal/route is not in std`）。
2. 模板自带 `go.mod` 会让它变成嵌套 module，使 `//go:embed` 报 `contains no embeddable files`，
   所以模块元数据必须写成 `*.tmpl`（`go.mod.tmpl` / `go.sum.tmpl` / `go.work.tmpl`），落地时再去掉后缀。
3. `//go:embed` 只能嵌**包目录以内**的文件 —— 模板因此必须待在 `cmd/go-infra-cli/` 下，
   挪不到仓库根，也挪不到与 `cli/` 同级。

> 原先放在 `testdata/templates/`。改成 `_templates/` 是为了与 `_features/` 对称，
> 且不再借 `testdata/` 的「测试数据」语义去表达「构建期资产」。

细节与踩坑记录见 `cli/cmd/go-infra-cli/templates_embed.go` 顶部注释，改动前先读。

## go-infra SDK 参考放在哪

go-infra 的包地图与用法速查**不单独维护成一份 skill**，而是作为模板规则下发：

```text
cli/cmd/go-infra-cli/_templates/single-starter/.cursor/rules/11-go-infra-api.mdc
cli/cmd/go-infra-cli/_templates/monorepo-starter/.cursor/rules/11-go-infra-api.mdc
```

理由：这份内容只在写 Go 代码时有意义，跟着模板生成进项目、随项目走，比在仓库外另存一份更可靠
—— 放外面不会自动同步，改了也白改。两份模板里的这一文件内容必须保持一致（改一处就改两处）。
新增 SDK 包时的同步要求见 `go-infra/README.md` 的「新增包后的同步 Checklist」。

## 改模板时的三条注意事项

- **换行必须是 LF。** 模板会被 `//go:embed` 原样写进用户项目；`Makefile` 一旦带 CRLF，
  在 Linux/macOS/WSL 下 `make` 会直接报 `missing separator`。仓库根的 `.gitattributes` 负责固化，
  CI 另有一道 `_templates` 无 CRLF 的检查。
- **别在模板里留模板名。** 渲染规则见 `cli/cmd/go-infra-cli/render.go`：`.go` 只改 import 前缀，
  `.mod` / `.work` 只改模块引用，其余文本会把模板名**整体换成项目名**。
- **改完跑一遍验证**：`cd cli && go build ./... && go vet ./... && go test ./...`，再跑
  `sh scripts/smoke.sh`（不联网的渲染冒烟，与 CI 是同一份脚本；Windows 下用 WSL 跑）。

## CI

`.github/workflows/ci.yml` 两个 job：

| job | 内容 |
|---|---|
| `cli` | gofmt 检查 + `go build` / `go vet` / `go test` |
| `templates` | `_templates` 不得含 CRLF；跑 `scripts/smoke.sh`（生成两种布局并做结构检查，全程不联网） |

## 许可证

Apache License 2.0，见 [LICENSE](LICENSE)。Copyright 2026 liukunxin。

选 Apache-2.0 而不是 MIT：第 3 条给出**显式专利授权**（MIT 全文不涉及专利），并规定
一旦就本作品发起专利诉讼则授权自动终止。对要长期开源的 SDK / 脚手架，这比 MIT 的
一句话免责更周全，也与 `go-infra` 的 LICENSE 文件保持一致。
