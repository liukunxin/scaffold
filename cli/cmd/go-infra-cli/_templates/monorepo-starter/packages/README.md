# packages/

**跨 Project 复用**的代码，按语言分组。

```text
packages/
├── go/          # Go 模块（独立 module，被 apps/services 以 replace 引用）
├── typescript/  # 前端包（npm package，用 file: 引用）
└── python/      # Python 包（可安装 distribution）
```

## 准入门槛（三条同时满足才放这里）

1. **被 2 个以上 Project 真实使用。** 只有一个使用者时，它就该待在那个 Project 的 `internal/`。
2. **不承载业务语义。** 一旦出现「订单」「支付」这类词，说明它属于某个业务 Project。
3. **边界稳定。** 接口三个月内不会因为某个业务需求而反复改。

反例（不要做）：

- `packages/common` / `packages/utils` / `packages/shared` —— 名字越泛，越会变成什么都往里塞的垃圾桶。
- 把某个 Project 的业务代码「上提」到 `packages/` 让另一个 Project 复用 —— 这只是把耦合换个地方藏。

## 提升路径

需要复用时按这个顺序走，别跳步：

```text
Project A 的 internal/app/<域>/     ← 先只在一个 Project 里长出来
        ↓ 第二个 Project 也要用，且语义一致
packages/<lang>/<name>/             ← 提为共享包，两个 Project 都改成引用它
        ↓ 若涉及跨语言的接口 / 数据形状
contracts/                          ← 抽成语言中立契约，各语言再出绑定
```

## 命名

| 位置 | 规则 | 例 |
|---|---|---|
| `packages/go/<name>` | 小写单词，不用连字符、不用泛名 | `contracts` |
| `packages/typescript/<name>` | npm 包名 `@repo/<name>` | `@repo/contracts` |
| `packages/python/<name>` | 发行名 `repo-<name>`，导入名 `<name>` | `repo-contracts` |

## 约束

- 每个包**自带清单文件**（`go.mod` / `package.json` / `pyproject.toml`），是独立单元，可单独测试与发版。
- 共享包**禁止反向依赖**任何 Project（`packages/` 里不许 import `apps/*` 或 `services/*`）。
- `packages/go/*` 之间可以互相依赖，但不要成环。
- 新增后跑 `make structure`，它会检查清单文件是否齐全。
