# tools/lint/

跨 Project 的静态检查。

```text
lint/
└── check-structure.py     # 目录约定检查
```

## check-structure.py

把「目录约定」变成 CI 拦得住的错误 —— 写在文档里的规则没人执行，写成脚本才有人执行。

```bash
make structure                                      # 检查本仓库（默认）
python tools/lint/check-structure.py /path/to/repo  # 检查指定仓库
```

检查三类东西：

| 检查项 | 不合格会怎样 |
|---|---|
| 仓库根骨架 | 缺 `go.work` / `apps` / `services` / `packages` / `contracts` / `README.md` → 失败 |
| `apps/*`、`services/*` 是不是 Project | 缺清单文件（`go.mod` / `package.json` / `pyproject.toml`）→ 失败 |
| Go Project 的布局 | 缺 `cmd` / `configs` / `internal/{app,bootstrap,infra,route}` / `Dockerfile` / `README.md` → 失败 |

只检查「文件 / 目录在不在」，不做代码解析 —— 所以没有依赖、跑得快，任何环境都能跑。
退出码 0 = 通过，1 = 有不合格项。

## 它检查不了什么

它**只**看结构，不看内容质量。下面这些它看不出来，靠 review 与 `.cursor/rules/`：

- `internal/` 里是不是真的按业务域垂直切片（有没有偷偷写成 `controller/service/dao` 横切）；
- 有没有跨 Project import `internal/`；
- 契约与实现是否一致。

真要拦住这些，得往里加解析逻辑 —— 加之前先想想值不值得，别把简单脚本写成半个编译器。
