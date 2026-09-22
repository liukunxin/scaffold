# packages/python/

Python 共享包。每个子目录是一个可安装 distribution：发行名 `repo-<name>`、导入名 `<name>`。

```text
python/
└── contracts/     # repo-contracts：契约的 Python 绑定
```

## 怎么被引用

不发布到 PyPI，使用方按相对路径可编辑安装：

```bash
pip install -e ../../packages/python/contracts
# 或
uv pip install -e ../../packages/python/contracts
```

## 加一个共享包

```text
packages/python/<name>/
├── pyproject.toml              # [project] name = "repo-<name>"
└── src/<name>/__init__.py      # [tool.hatch.build.targets.wheel] packages = ["src/<name>"]
```

照 `contracts/` 抄结构即可。`src/` 布局是为了让「可安装」和「能直接跑测试」两件事不冲突。

## 现有内容

| 包 | 说明 |
|---|---|
| `repo-contracts` | 事件信封的 Python 绑定（`Envelope` / `new_envelope` / `to_wire` / `from_wire` / `EVENT_TYPE_PATTERN`），**零第三方依赖**。见 `contracts/README.md` |

## 怎么自检

根 `Makefile` 的 `py-check` 会把 `src/` 加进 `sys.path` 后 import 并构造一个信封：

```bash
make py-check
```

包本身只用标准库，所以没有虚拟环境也能跑。

## 约束

- **禁止** import `apps/*` 或 `services/*`。
- 依赖尽量少：标准库能写完的不要引第三方（这个包一个第三方依赖都没有）。
- Python 侧字段名用 `snake_case`，但**序列化到 wire 上必须是 `camelCase`**，
  由 `to_wire()` / `from_wire()` 负责转换 —— wire 上的名字三个语言必须一致，否则跨语言订阅解不开。
