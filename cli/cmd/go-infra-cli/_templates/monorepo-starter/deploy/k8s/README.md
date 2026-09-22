# deploy/k8s/

Kubernetes 清单（原生 YAML）。

```text
k8s/
└── （当前为空）
```

## 该放什么

一个 Project 一组清单，至少四样：

```text
k8s/<category>-<name>/
├── deployment.yaml
├── service.yaml
├── configmap.yaml      # 非敏感配置，键与 config.<env>.yml 对应
└── secret.yaml         # 只放占位或引用，真实值走集群 Secret
```

## 环境差异怎么处理

用 Kustomize 的 `overlays/<env>/` 覆盖，**不要复制整套清单** —— 复制之后必然出现
「dev 的清单改了、prod 忘了改」。

```text
k8s/<category>-<name>/
├── base/
└── overlays/{dev,gray,prod}/
```

## 什么时候换成 helm/

清单开始需要参数化（同一应用部署给多个团队 / 多套环境）时换 `helm/`。
只有一两套环境的话，Kustomize 足够。
