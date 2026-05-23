# NexCore Go SDK

全能 Go 客户端,覆盖 Payment / Exchange / Energy / SMTP **全部 25 个 v1 公开 endpoint**.

**零依赖**(仅标准库 `net/http` / `crypto/hmac` / `encoding/json`).

## 环境

- Go 1.21+

## 安装

```bash
go get github.com/DoBestone/nexcore-sdk/go
```

## 文件结构

```
go.mod                 (module: github.com/DoBestone/nexcore-sdk/go)
doc.go                 package 文档
client.go              主客户端 Client + Config + NewClient()
http.go                底层 HTTP 传输
errors.go              统一异常 Error + AsError()
payment.go             多链收款(7 endpoints)
exchange.go            汇率(5 endpoints)
energy.go              TRON 能量租赁(8 endpoints)
smtp.go                SMTP 聚合 API(5 endpoints)
examples/
├── create_order/main.go
└── webhook/main.go
```

## 用法

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    nexcore "github.com/DoBestone/nexcore-sdk/go"
)

func main() {
    c := nexcore.NewClient(nexcore.Config{
        BaseURL:         "https://your-domain.com",
        PaymentAppID:    "APP20260412XXXX",
        PaymentAppKey:   "your_app_key_here",
        EnergyAPIKey:    "energy_api_key_here",
        EnergySecretKey: "energy_secret_key_here",
        SMTPAPIKey:      "smk_xxx",
    })

    // 创建支付订单
    raw, err := c.Payment.CreateOrder(map[string]any{
        "out_order_id": fmt.Sprintf("ORDER_%d", time.Now().Unix()),
        "amount":       "100.00",
        "currency":     "CNY",
        "trade_type":   "usdt.trc20",
        "call_type":    "rotation",
        "timeout":      1800,
    })
    if err != nil {
        if ne := nexcore.AsError(err); ne != nil {
            log.Fatalf("Error #%d: %s (trace=%s)", ne.Code, ne.Message, ne.RequestID)
        }
        log.Fatal(err)
    }
    var order struct {
        PayAddress string `json:"pay_address"`
    }
    _ = json.Unmarshal(raw, &order)
    fmt.Println("支付地址:", order.PayAddress)

    // 查询汇率
    raw, _ = c.Exchange.GetRate("USDT", "CNY")
    var rate struct { Rate float64 `json:"rate"` }
    _ = json.Unmarshal(raw, &rate)
    fmt.Printf("USDT/CNY: %.4f\n", rate.Rate)

    // 估算能量
    raw, _ = c.Energy.EstimateEnergy("TXxxxxxxxxxxxxxxxxxxxxx")
    var est struct { EstimatedEnergy int `json:"estimated_energy"` }
    _ = json.Unmarshal(raw, &est)
    fmt.Println("需要能量:", est.EstimatedEnergy)

    // 发送邮件
    raw, _ = c.SMTP.Send(map[string]any{
        "to":      "user@example.com",
        "subject": "验证码",
        "body":    "<h1>123456</h1>",
        "is_html": true,
    })
    var mail struct { MessageID string `json:"message_id"` }
    _ = json.Unmarshal(raw, &mail)
    fmt.Println("消息 ID:", mail.MessageID)
}
```

## API 列表

### `client.Payment` — 多链收款(7 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `CreateOrder(params)` | POST | `/api/v1/pay/create` |
| `QueryOrder(outOrderID)` | GET | `/api/v1/pay/query` |
| `CloseOrder(outOrderID)` | POST | `/api/v1/pay/close` |
| `GetAppConfig()` | GET | `/api/v1/pay/app-config` |
| `BindAddress(userID, tradeType)` | POST | `/api/v1/pay/bind-address` |
| `GetUserAddress(userID, tradeType)` | POST | `/api/v1/pay/get-address` |
| `UnbindAddress(userID)` | POST | `/api/v1/pay/unbind-address` |
| `Sign(params)` | (工具) | HMAC-SHA256 签名 |
| `VerifyNotifySign(payload)` | (工具) | webhook 校验(常量时间) |

### `client.Exchange` — 汇率(5 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `GetRate(from, to)` | GET | `/api/v1/rate` |
| `Convert(from, to, amount)` | POST | `/api/v1/convert` |
| `GetRates(symbols, base)` | GET | `/api/v1/rates` |
| `GetFiatRates(base)` | GET | `/api/v1/rates/fiat` |
| `GetAllRates(base)` | GET | `/api/v1/rates/all` |

### `client.Energy` — TRON 能量租赁(8 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `GetInfo()` | GET | `/api/v1/energy/info` |
| `GetPrice(energy, period)` | GET | `/api/v1/energy/price` |
| `EstimateEnergy(receiveAddr)` | GET | `/api/v1/energy/estimate-energy` |
| `CreateOrder(params)` | POST | `/api/v1/energy/order` |
| `CreateOnetimeOrder(params)` | POST | `/api/v1/energy/order/onetime` |
| `QueryOrder(serial)` | GET | `/api/v1/energy/order/:serial` |
| `ListOrders(filter)` | GET | `/api/v1/energy/orders` |
| `ReclaimOrder(serial)` | POST | `/api/v1/energy/order/reclaim` |

### `client.SMTP` — SMTP 聚合(5 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `Send(params)` | POST | `/api/v1/smtp/send` |
| `SendBatch(params)` | POST | `/api/v1/smtp/send/batch` |
| `SendTemplate(params)` | POST | `/api/v1/smtp/send/template` |
| `GetQuota()` | GET | `/api/v1/smtp/quota` |
| `GetStatus(messageID)` | GET | `/api/v1/smtp/status/:message_id` |

## Webhook 签名校验

```go
http.HandleFunc("/payment/notify", func(w http.ResponseWriter, r *http.Request) {
    var payload map[string]any
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "bad request", 400)
        return
    }
    if !c.Payment.VerifyNotifySign(payload) {
        http.Error(w, "invalid sign", 400)
        return
    }
    // 处理回调... 务必幂等
    w.Write([]byte("OK"))
})
```

`VerifyNotifySign` 内部用 `hmac.Equal`,常量时间比较防时序攻击.

## 异常

所有方法返回错误用 `*nexcore.Error`,字段:

- `Code` — 平台错误码(0=成功)
- `Message` — 错误描述
- `RequestID` — 服务端追踪 ID(响应头 `X-Trace-Id`)
- `HTTPStatus` — HTTP 状态码

用 `nexcore.AsError(err)` 把通用 `error` 转成 `*nexcore.Error`.

## 返回值约定

所有方法返回 `json.RawMessage` — 业务方自行 `json.Unmarshal` 到具体 struct.

这样设计的好处:**后端 API 加字段不需要升级 SDK**,业务方按需取字段即可.

## 示例

见 [`examples/`](./examples/):
- `create_order/main.go` — 完整下单(`go run ./examples/create_order`)
- `webhook/main.go` — HTTP webhook 接收(`go run ./examples/webhook`)
