# scaffold

go-infra 官方脚手架：一条命令生成可运行的项目骨架。

| 布局 | 模板 | 用途 |
|---|---|---|
| `single` | `single-starter` | 单服务；也是 **Project 层**架构标准 |
| `monorepo` | `monorepo-starter` | 跨语言多 Project；定义 **Repository 层**规范 |

`apps/*` / `services/*` 下的 Go Project 一律复用 `single-starter`，不另造一套服务架构。
模板通过 `//go:embed` 编进 CLI，安装一个二进制即可用。

## 快速开始

```bash
git clone https://github.com/liukunxin/scaffold.git
cd scaffold/cli
make cli-build        # 产物：cli/bin/go-infra-cli
# 或：make cli-install
```

```bash
# 单服务
go-infra-cli init myapp --module github.com/acme/myapp --layout single

# Monorepo
go-infra-cli init collab-platform --module github.com/acme/collab-platform --layout monorepo
cd collab-platform
go-infra-cli mono add service order
go-infra-cli mono add app bff
```

命令、`--features` / `--scenes`、能力增删等完整说明见 [`cli/README.md`](cli/README.md)。

## 仓库结构

```text
scaffold/
├── cli/                    # Go module + 内嵌模板（_templates / _features）
├── scripts/smoke.sh        # 渲染冒烟（与 CI 共用）
├── AGENTS.md               # 改本仓库的约束
└── .github/workflows/      # CI：cli 测试 + templates 冒烟
```

## 相关文档

| 文档 | 内容 |
|---|---|
| [`cli/README.md`](cli/README.md) | CLI 安装与用法 |
| [`AGENTS.md`](AGENTS.md) | 改模板 / CLI 时必须遵守的约定 |
| [monorepo 模板 README](cli/cmd/go-infra-cli/_templates/monorepo-starter/README.md) | 目录职责、示范清单、「7 个问题」 |

改完请跑：`cd cli && go test ./...`，再 `sh scripts/smoke.sh`（Windows 用 WSL）。

## 许可证

[Apache License 2.0](LICENSE)。Copyright 2026 liukunxin。
选用 Apache-2.0（含显式专利授权），与 `go-infra` 保持一致。
