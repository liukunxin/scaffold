# packages/python/contracts

`contracts/events/envelope.schema.json` 的 **Python 绑定**（跨 Project 事件信封）。

三份绑定共用同一份契约：

| 语言 | 位置 |
|---|---|
| 定义（语言中立） | `contracts/events/envelope.schema.json` |
| Go | `packages/go/contracts/events/envelope.go` |
| TypeScript | `packages/typescript/contracts/src/events.ts` |
| Python | 本包 |

## 用法

```python
from contracts import Envelope, new_envelope

env = new_envelope("evt-1", "demo.ping.completed.v1", {"message": "pong"})
env.to_wire()                      # camelCase，与 Go / TS 一致
Envelope.from_wire(incoming_json)  # 忽略未知字段
```

## 怎么被别的 Project 引用

本包不发布到 PyPI。同仓 Project 用**相对路径**装它（`pyproject.toml`）：

```toml
[project]
dependencies = ["repo-contracts"]

[tool.uv.sources]
repo-contracts = { path = "../../packages/python/contracts", editable = true }
```

用 pip 的话：`pip install -e ../../packages/python/contracts`。

## 本地自检

没有额外依赖，直接跑：

```bash
python -c "import sys; sys.path.insert(0, 'src'); import contracts; print(contracts.new_envelope('e1', 'demo.ping.completed.v1').to_wire())"
```
