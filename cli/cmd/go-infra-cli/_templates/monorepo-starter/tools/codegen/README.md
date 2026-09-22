# tools/codegen/

契约 → 各语言桩代码的生成脚本。

```text
codegen/
└── （当前为空）
```

## 什么时候需要

示范契约只有一个方法：Go 侧**手写 ServiceDesc**、TS / Python 绑定也是手写的，够用。
出现下面任一情况就该上生成：

- `.proto` 的方法数上到个位数，手写 `ServiceDesc` 开始出错；
- 同一份契约要在 3 个以上语言出绑定，手抄必然漂移；
- 契约改动频繁，每次都要人肉同步所有绑定。

## 该写什么

按语言各一个脚本，输入是 `contracts/`，输出到 `packages/<lang>/`：

```text
proto/**/*.proto       → protoc + protoc-gen-go / -go-grpc / -es / -python
openapi/*.yaml         → openapi-generator，或手写模板
events/*.schema.json   → 按 schema 生成各语言的 payload 类型
```

## 约定

- 生成物**入库**（否则没装工具链的人无法构建）。
- 脚本要幂等：同样的输入跑两次，输出逐字节一致。
- 生成的目录里放一个说明文件，写明「本目录由 `tools/codegen/xxx` 生成，不要手改」。
