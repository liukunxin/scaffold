# deploy/

**跨 Project** 的部署编排。Project 自己的 `Dockerfile` 留在 Project 内部。

```text
deploy/
├── docker/       # 本地 / 联调用的 compose（现有 docker-compose.dev.yml）
├── k8s/          # K8s 清单
├── helm/         # Helm chart
└── terraform/    # 基础设施
```

## 边界

| 放这里 | 放 Project 内部 |
|---|---|
| 多个 Project 一起起的环境（compose、K8s 编排） | 单个服务的 `Dockerfile` |
| 共享中间件（redis / mysql / mq）的部署 | 服务自己的 `configs/config.<env>.yml` |
| 集群、网络、域名等基础设施 | 服务自己的构建脚本 |

判断标准：**「删掉某个 Project，这份文件还需要存在吗？」** 需要 → 放 `deploy/`；
不需要 → 它属于那个 Project。

## 本地联调

```bash
docker compose -f deploy/docker/docker-compose.dev.yml up --build
```

起 `gateway` + `redis`。注意 compose 里 `build.context` 是**仓库根**（`../..`），不是服务目录
—— 服务的 `go.mod` 用相对路径 `replace` 引用了 `packages/go/contracts`，只拷服务自己会在依赖
解析阶段就失败。

## 现有内容

| 文件 | 说明 |
|---|---|
| `docker/docker-compose.dev.yml` | 最小联调环境：gateway + redis，带健康检查 |

## 约定

- 各子目录的细节见各自的 `README.md`。
- 环境差异用**变量**表达（compose 的 `env`、Helm 的 `values-<env>.yaml`），不要复制多份编排文件。
- 密钥不进仓库：用 K8s Secret / 外部密钥管理，仓库里只放占位与文档。
