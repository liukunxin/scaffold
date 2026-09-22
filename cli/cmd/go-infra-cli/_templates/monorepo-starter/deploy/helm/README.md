# deploy/helm/

Helm chart。适用于同一应用要多套环境、多个团队复用的场景。

```text
helm/
└── （当前为空）
```

## 该放什么

- 每个可复用应用一个 chart：`helm/<name>/`（`Chart.yaml` / `values.yaml` / `templates/`）。
- 环境差异走 `values-<env>.yaml`，**不要**在模板里堆 `if env == "prod"`。

```bash
helm upgrade --install gateway helm/gateway -f helm/gateway/values-prod.yaml
```

## 约定

- `values.yaml` 的默认值必须能在空集群里跑起来（哪怕是单副本、最小资源）。
- 密钥不在 `values-*.yaml` 里；用 `existingSecret` 引集群 Secret。
- chart 版本与 Project 版本**分开**：Project 发版不必每次发 chart，但 chart 改动要写 changelog。
