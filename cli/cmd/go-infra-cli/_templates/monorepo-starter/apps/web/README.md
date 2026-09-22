# apps/web

面向用户的前端 Project 示范（Vite + TypeScript，零框架）。

它示范的不是"怎么写页面"，而是**一个非 Go Project 在 Monorepo 里怎么落位**：

| 问题 | 这个 Project 的答案 |
|---|---|
| 放哪 | `apps/web/`，自带 `package.json`，是一个独立 Project |
| 怎么复用公共代码 | 相对路径依赖 `packages/typescript/contracts`（`"@repo/contracts": "file:../../..."`） |
| 怎么调别的 Project | HTTP。`/api/*` 由 vite 反代到 `services/gateway`（见 `vite.config.ts`） |
| 不能做什么 | 不 import `services/gateway/internal/**`，也不直接读它的配置 |

## 跑起来

```bash
# 1. 先起后端（仓库根目录另开一个终端）
cd services/gateway && go run ./cmd/http      # :8080

# 2. 再起前端
cd apps/web
npm install
npm run dev                                    # http://127.0.0.1:5173
```

页面上的按钮会请求 `/api/demo/ping`，把 gateway 返回的原始 JSON 打出来。
后端没起时会显示请求失败——这是预期的，说明反代链路本身是通的。

```bash
npm run build      # tsc --noEmit + vite build
```

## 目录

```text
apps/web/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts     # /api 反代到 services/gateway
└── src/
    ├── api.ts         # 唯一的对外调用点，类型来自 @repo/contracts
    ├── main.ts        # 入口与最小 UI
    └── style.css
```

## 换成真实框架

要换成 Next / Nuxt / Taro，直接用它自己的脚手架重建本目录即可，**位置和引用方式不变**：

```bash
rm -rf apps/web && npx create-next-app@latest apps/web
```

Monorepo 只约束「放在 `apps/`」「通过 `packages/` 和 `contracts/` 复用」这两件事。

> `npm install` 会按 `package.json` 里的 `file:` 依赖把 `packages/typescript/contracts`
> 以软链接形式装进来，不需要额外的 workspace 配置。
