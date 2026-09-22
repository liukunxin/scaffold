# packages/typescript/

前端共享包。每个子目录是一个 npm package，包名统一 `@repo/<name>`。

```text
typescript/
└── contracts/     # @repo/contracts：契约的 TS 绑定
```

## 怎么被引用

**不发布到 registry**，使用方用 `file:` 协议直接指向本仓库目录：

```json
{
  "dependencies": {
    "@repo/contracts": "file:../../packages/typescript/contracts"
  }
}
```

`apps/web/package.json` 就是这么写的。改完共享包，使用方 `npm install` 一次即可生效。

## 加一个共享包

```text
packages/typescript/<name>/
├── package.json     # "name": "@repo/<name>"、"private": true、"types": "./src/index.ts"
├── tsconfig.json
└── src/index.ts     # 从源码引用，不产出 dist
```

照 `contracts/` 那三个文件抄结构最省事，然后在使用方的 `package.json` 里加 `file:` 依赖。

## 现有内容

| 包 | 说明 |
|---|---|
| `@repo/contracts` | 事件信封的 TS 绑定（`Envelope` / `newEnvelope` / `isEnvelope` / `EVENT_TYPE_PATTERN`），字段与 Go、Python 绑定严格对齐。见 `contracts/README.md` |

## 约束

- **禁止** import `apps/*` 或 `services/*`。
- 只放**纯 TS/JS**：类型、纯函数、协议解析。依赖 React 或浏览器 API 的组件属于具体 `apps/`，不要上提。
- 不做构建产物管理（不产出、不入库 `dist/`），使用方直接按源码引用 `src/`。
- 改完在使用方跑一次 `npm run build --prefix apps/web` 验证能编过（或至少 `tsc --noEmit`）。
