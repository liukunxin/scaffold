# tools/release/

按 Project 发版：changelog 生成、版本号、打 tag。

```text
release/
└── （当前为空）
```

## 为什么是「按 Project」而不是「按仓库」

本仓库是多 Project：`services/order` 与 `services/gateway` 各自独立发版、独立 tag。
不要做「整个仓库一个版本号」—— 那会把所有 Project 的发布节奏绑死在一起，
某个服务要紧急修一个 bug 也得跟别人一起发。

## 该写什么

至少两份：

1. **changelog 生成**：按目录（`apps/<名>`、`services/<名>`）筛 commit，生成各自的 `CHANGELOG.md` 片段。
2. **打 tag**：约定 tag 规则，例如 `<category>-<name>-v<semver>`（`service-order-v1.2.0`），
   由一个脚本统一创建，避免手打不一致。

## 约定

- tag 规则一旦定下**不要改** —— 改了 CI 与历史发布记录就全断了。
- 发版前必须：该 Project `make test` 通过、仓库 `make structure` 通过。
- 交叉依赖要在 changelog 里写清（例如「本版起依赖 `packages/go/contracts` v0.3.0」），
  否则回滚时会漏掉对方。
