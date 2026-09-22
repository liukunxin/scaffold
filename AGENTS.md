# AGENTS.md — 改 scaffold 本仓库时的约束

> 这份是给**改 scaffold 本身**的人与 AI 看的。
> 模板里另有一份 `cli/cmd/go-infra-cli/_templates/*/AGENTS.md`，那份会被生成进用户项目，是给**用模板的人**看的，两者不要混。

## 这个仓库是什么

一条命令生成两类项目骨架：

- `single-starter` —— 单服务项目，定义 **Project 层**架构（一个服务内部分层）。
- `monorepo-starter` —— 跨语言、多 Project 仓库，定义 **Repository 层**规范；它内部不另发明一套服务架构，
  `apps/*` 与 `services/*` 下的 Go Project 一律复用 `single-starter`。

模板通过 `//go:embed` 编进 CLI 二进制，用户只需装 CLI 一个东西。CLI 与模板都在 `cli/` 下。

## 改之前必须知道的五条

1. **模板目录必须以 `_` 开头**（`_templates/`、`_features/`），且只能待在 `cli/cmd/go-infra-cli/` 内。
   原因有两个：`//go:embed` 只能嵌**包目录以内**；目录名不带 `_` 会被 CLI 自己的 `go build ./...`
   当成普通包编译并直接报错。完整解释见 `templates_embed.go` 顶部注释，改动前先读。
2. **模块元数据必须写成 `*.tmpl`**（`go.mod.tmpl` / `go.sum.tmpl` / `go.work.tmpl`）。
   模板目录里一旦出现 `go.mod`，Go 就把它视作嵌套 module，`//go:embed` 会报 `contains no embeddable files`。
   落地时 `copyTree` 会去掉 `.tmpl` 后缀。
3. **模板必须 LF 换行。** 模板会被原样写进用户项目，`Makefile` 带 CRLF 会让 `make` 在 Unix 上报
   `missing separator`。三道防线：根 `.gitattributes`、`.editorconfig`、CI 的 `_templates` 无 CRLF 检查。
4. **能力注入靠成对锚点。** 注入标记是 `// FEATURE:<cap>:<SEG>:START/END`，锚点声明在
   `project_layout.go` 的 `requiredFeatureAnchors`，缺任一个 `add` 会直接失败。
   新增一个能力要**同时改两处**：`feature_sync.go` 里加 spec + 模板 `internal/bootstrap/app.go`、
   `internal/route/init.go` 里有对应锚点。
5. **模板与 CLI 的注入逻辑是同一个原子变更集**，必须一起提交。拆开提交会出现
   "`add` 报 missing anchor" 的中间态。

## 改完怎么验证

```sh
cd cli && go build ./... && go vet ./... && go test ./... && gofmt -l .
sh scripts/smoke.sh          # 不联网的渲染冒烟：两种布局 + 生成物结构检查（Windows 用 WSL 跑）
```

`scripts/smoke.sh` 与 CI 的 `templates` job 是同一份脚本，本地跑通即 CI 能过。

## 别做的事

- 不要在仓库根加 `go.mod` —— 这里只有 `cli/` 一个 module。
- 不要在模板里留模板名（`single-starter` / `monorepo-starter`）。渲染规则见 `render.go`：
  `.go` 只改 import 前缀，`.mod` / `.work` 只改模块引用，其余文本会把模板名整体换成项目名。
- 不要为了让检查通过而放松 `tools/lint/check-structure.py`。模板产不出合规产物就**补模板**
  （`Dockerfile` 那次就是补模板，而不是放宽检查）。
- 控制台输出一律英文（CLI、`tools/lint/*.py`、`Makefile` 的 `help`）。中文提示在 Windows 的 GBK
  控制台下是乱码；代码注释与写进生成物的文档保持中文。
