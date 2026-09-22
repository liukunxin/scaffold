# docs/conventions/

命名、目录、提交、评审约定。

```text
conventions/
└── （当前为空）
```

## 为什么现在是空的

大部分约定已经写成**可执行的东西**，那比文档可靠：

| 约定 | 落在哪 |
|---|---|
| 目录结构与 Project 布局 | `tools/lint/check-structure.py`（CI 会跑） |
| 分层、依赖方向、跨 Project 边界 | `.cursor/rules/00-architecture.mdc` |
| Go 代码风格与 SDK 用法 | `.cursor/rules/10-go-sdk-first.mdc`、`11-go-infra-api.mdc` |
| 路由与统一响应壳 | `.cursor/rules/12-http-routing.mdc` |
| 契约与事件版本 | `.cursor/rules/20-contracts.mdc`、`contracts/README.md` |

## 往这里补什么

只补**暂时没法自动化**的约定，并且要写清「为什么」：

- 提交信息格式、分支策略、PR 评审规则；
- 评审里反复出现的同类意见（先在这里固化一条，再想办法自动化）；
- 无法用脚本校验的命名问题（例如业务域该怎么切）。

## 加约定的规则

1. 先问「能不能写成检查脚本或规则文件」。能写就写脚本，别只写文档。
2. 写文档时给出**反例**。只说「应该怎样」的约定没人记得住，附一个「不要这样」才有效。
3. 与既有规则冲突时**先删旧的**，不要两份并存。
