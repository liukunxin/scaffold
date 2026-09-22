# tools/

**跨 Project** 的工程脚本：目录约定检查、契约代码生成、发版。

```text
tools/
├── lint/           # 跨 Project 静态检查（现在有 check-structure.py）
├── codegen/        # 契约 → 各语言桩代码
└── release/        # changelog / 版本
```

## 什么进 tools/

- 至少服务 **2 个 Project** 的脚本。
- 只用标准库或内置依赖就能跑起来的脚本 —— 工程脚本本身不该有复杂的安装步骤。

## 什么不进 tools/

| 想放的东西 | 应该去哪 |
|---|---|
| 某个 Project 自己的构建 / 运行脚本 | 那个 Project 的 `Makefile`（如 `services/gateway/Makefile`） |
| 部署编排 | `deploy/` |

## 现有内容

| 脚本 | 用途 | 怎么跑 |
|---|---|---|
| `lint/check-structure.py` | 校验目录约定：仓库根骨架是否齐全、每个 `apps/*` `services/*` 是不是合规 Project | `make structure`（已纳入 `make check`） |

## 写新脚本的约定

- **退出码即结论**：0 = 通过，非 0 = 不合格，CI 直接拿它卡门禁。不要只打印警告然后退 0。
- **可指定根目录**：支持 `python tools/lint/xxx.py /path/to/repo`，默认取脚本所在的仓库。
- **输出要能一眼看出哪条不合格**：一行一条，带相对路径与缺失项，不要输出一大段散文。
