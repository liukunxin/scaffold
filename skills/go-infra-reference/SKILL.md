---
name: go-infra-reference
description: >-
  Complete API reference for github.com/liukunxin/go-infra — the company's
  internal Go infrastructure library. Read this skill before writing any code
  that uses Redis, LLM, HTTP client, MySQL, logging, tracing, error handling,
  traffic control, project bootstrapping, SSE, WebSocket, Kafka, Milvus,
  Elasticsearch, object storage, login, payment, or realtime collab in a Go
  service that imports go-infra.
  Also read when deciding which go-infra utility to reuse instead of writing custom code.
disable-model-invocation: true
---

# go-infra Reference

> Module: `github.com/liukunxin/go-infra`
> Repo: `D:\WorkSpace\infra\go-infra`（内网，完整 README 见 `go-infra/README.md`）

## 完整包地图

### pkg/base — 原子基础能力（项目默认启用）

| Import path | Purpose | Key entry point |
|---|---|---|
| `pkg/base/log` | 异步结构化日志 | `log.WithContext(ctx).Info/Warn/Errorf(...)` |
| `pkg/base/errors` | 业务错误封装 | `kerr.WrapError(status, code, err)` |
| `pkg/base/config` | YAML + 环境覆盖 + 加密 | `kconfig.Load[T](opts...)` |
| `pkg/base/trace` | OpenTelemetry 链路追踪 | `trace.Init(opts...)` |
| `pkg/base/env` | 应用名/环境名 | `env.SetName()` / `env.SetEnv()` |
| `pkg/base/uuid` | Snowflake ID | `uuid.GetIDService().GenerateUserID()` → int64 |
| `pkg/base/xutil` | 泛型工具 | `xutil.Map/Filter/Contain/...` |

### pkg/infra — 基础设施客户端

| Import path | 初始化 | 运行时入口 | 说明 |
|---|---|---|---|
| `pkg/infra/redis` | `iredis.Init(&cfg.Redis)` | `iredis.GetClient()` → `*redis.Client` | go-redis/v9 |
| `pkg/infra/mysql` | `mysql.Init(cfg.Mysql)` | `mysql.GetClient()` → `*gorm.DB` | GORM，无外键 |
| `pkg/infra/http_client` | `httpclient.Init(cfg.HTTP)` | `httpclient.GetHTTPClient()` / `GetTransport()` | 共享连接池 |
| `pkg/infra/llm` | `llm.New(opts...)` | `client.Generate/GenerateStream` | 多厂商统一 |
| `pkg/infra/metrics` | `metrics.InitMetrics(name, router)` | 自动注册 `/metrics` | Prometheus |
| `pkg/infra/pprof` | `pprof.Start()` | 自动 | pprof HTTP |
| `pkg/infra/traffic` | `kitraffic.Init(opts...)` | 自动 | 限流/熔断 |
| `pkg/infra/grpc` | `grpc.NewServer/Client(cfg)` | - | 拦截器/治理 |
| `pkg/infra/websocket` | `websocket.NewHub(cfg)` | `Hub.Run()` | 心跳/重连 |
| `pkg/infra/kafka` | `kafka.NewProducer/Consumer(cfg)` | `producer.Publish/consumer.Subscribe` | Sarama |
| `pkg/infra/milvus` | `milvus.Init(cfg)` | `milvus.GetClient()` | 向量数据库 |
| `pkg/infra/es` | `es.Init(cfg)` | `es.GetClient()` | Elasticsearch |
| `pkg/infra/objstore` | `objstore.New(cfg)` | `client.Put/Get/Delete` | KS3/OSS/MinIO |
| `pkg/infra/apollo` | `apollo.Init(cfg)` | - | 配置中心热更新 |

### pkg/biz — 业务组件

| Import path | 入口 | 说明 |
|---|---|---|
| `pkg/biz/controller` | embed `kcontroller.GinBase` | `SuccessResponse/ErrorResponse` |
| `pkg/biz/middlewares` | `bizmw.GinTraceMiddleware()` / `bizmw.HttpLogRecord()` | Gin 中间件 |
| `pkg/biz/sse` | SSE 推送封装 | - |
| `pkg/biz/login` | 多方式登录 + JWT | 密码/手机/邮箱/微信 |
| `pkg/biz/account` | 账号管理与多登录绑定 | - |
| `pkg/biz/pay` | 微信 APIv3 / 支付宝 RSA2 | - |
| `pkg/biz/collab` | 定序/幂等/回放/订阅引擎 | 实时协作 |
| `pkg/biz/image` | 图片处理服务 | - |

---

## Critical Patterns

### 日志
```go
import log "github.com/liukunxin/go-infra/pkg/base/log"

log.WithContext(ctx).Infof("user %d login", uid)
log.WithContext(ctx).WithFields(map[string]interface{}{"uid": uid, "action": "login"}).Warn("quota")
// bootstrap only:
log.Init(cfg.Log)
defer log.Close()
```

### 错误
```go
import kerr "github.com/liukunxin/go-infra/pkg/base/errors"

// Status 常量: StatusOK/BadRequest/Unauthorized/Forbidden/NotFound/InternalServerError/TooManyRequests
return kerr.WrapError(kerr.StatusForbidden, codes.ErrQuotaExceeded, err)
```

### 配置加载
```go
import kconfig "github.com/liukunxin/go-infra/pkg/base/config"

cfg, err := kconfig.Load[App](
    kconfig.WithEnvFrom("env"),
    kconfig.WithValidate(true),
    kconfig.WithTagValidation(true),
    // 生产环境解密：
    kconfig.WithDecrypt(kconfig.AESKeyFromEnv("CONFIG_ENCRYPT_KEY")),
)
```

配置加密工作流：
```bash
go-infra-cli keygen                    # 生成 32 字节密钥（hex）
go-infra-cli encrypt -key $KEY -value "mysecret"  # → ENC(base64...)
```

### Redis（含原子配额模式）
```go
import iredis "github.com/liukunxin/go-infra/pkg/infra/redis"

// 普通操作（返回 go-redis/v9 *redis.Client）：
iredis.GetClient().Set(ctx, key, val, ttl).Err()
iredis.GetClient().Get(ctx, key).Result()  // → (string, error)

// 原子配额（INCR → 超限 DECR → 首次写入 Expire）：
n, _ := iredis.GetClient().Incr(ctx, key).Result()
if int(n) > limit {
    iredis.GetClient().Decr(ctx, key)
    return errQuota
}
if n == 1 {
    iredis.GetClient().Expire(ctx, key, ttl) // 防 Redis 内存泄漏
}
```

### HTTP 连接池
```go
import httpclient "github.com/liukunxin/go-infra/pkg/infra/http_client"

// 短请求（内置 timeout 和 body 大小限制）：
body, status, err := httpclient.GetHTTPClient().Get(ctx, url, headers)
body, status, err := httpclient.GetHTTPClient().Post(ctx, url, reqBody, headers)

// SSE / 流式（Timeout=0，用 context 控超时）：
streamCl := &http.Client{Transport: httpclient.GetTransport(), Timeout: 0}
reqCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(reqCtx, http.MethodPost, url, body)
```

### LLM 流式
```go
import "github.com/liukunxin/go-infra/pkg/infra/llm"

p, _ := llm.NewOpenAICompatibleProvider("gw", llm.OpenAICompatibleConfig{
    BaseURL:      cfg.BaseURL,
    APIKey:       cfg.APIKey,
    DefaultModel: "deepseek-v3",
    Headers: map[string]string{
        "AI-Gateway-Uid":            uid,
        "AI-Gateway-Product-Name":   productName,
        "AI-Gateway-Intention-Code": intentionCode,
    },
    HTTPTimeout: 30 * time.Second,
})
client, _ := llm.New(llm.WithProvider(p))

// 非流式：
resp, _ := client.Generate(ctx, llm.GenerateRequest{Messages: msgs})
resp.Content // string

// 流式：
ch, _ := client.GenerateStream(ctx, llm.GenerateRequest{Messages: msgs})
for ev := range ch {
    if ev.Err != nil || ev.Done { break }
    fmt.Print(ev.Delta)
}

// ⚠️ 重要限制：GenerateStream 仅返回文本 delta，不含 tool_calls 解析
// 需要 tool_calls 时须自行实现 SSE 解析（见 reference.md）
```

### Gin 统一响应
```go
import kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"

type myCtrl struct { kcontroller.GinBase }
func (c *myCtrl) Handle(ctx *gin.Context) {
    data, err := c.logic.Do(ctx.Request.Context())
    if err != nil { c.ErrorResponse(ctx, err); return } // 自动映射 HTTP status
    c.SuccessResponse(ctx, data) // {"code":0,"msg":"ok","data":{...}}
}
```

---

## 初始化顺序（bootstrap/app.go）

```
log → trace → mysql → redis → pprof → httpclient → traffic → gin → metrics → route.Setup
```

## 禁止事项

- 业务代码中 `&http.Client{}`（绕过连接池）
- `fmt.Println` / 自建 zap（绕过统一日志）
- 手写 JSON 错误体（用 `ErrorResponse`）
- MySQL `FOREIGN KEY` 约束

---

## Additional Details

- tool_call 流式实现模式、KSO 中间件、配置结构完整示例：[reference.md](reference.md)
