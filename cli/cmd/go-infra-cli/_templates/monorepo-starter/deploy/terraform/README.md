# deploy/terraform/

基础设施即代码：集群、网络、数据库实例、域名、对象存储。

```text
terraform/
└── （当前为空）
```

## 该放什么

- 云资源定义：VPC / 集群 / RDS / Redis / 对象存储 / DNS。
- **不放**应用部署 —— 那是 `k8s/` 或 `helm/` 的事。
- **不放**任何明文密钥 —— 用变量 + 远端 state + 密钥管理。

## 目录怎么分

```text
terraform/
├── modules/          # 可复用模块
└── envs/{dev,prod}/  # 每套环境一份，各自 state
```

## 约定

- state 必须**远端存储 + 加锁**，不要留在本地或提交进仓库。
- `terraform plan` 的输出进 CI 记录，`apply` 需人工确认。
- 破坏性变更（换实例类型、删库）在 PR 描述里单独标出来。
