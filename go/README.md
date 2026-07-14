# NexCore Go SDK

全能 Go 客户端,覆盖 Payment / Exchange / Energy / SMTP / Withdraw / Account / VCard **7 大命名空间全部 44 个 v1 公开 endpoint**.

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
payment_test.go        签名语义测试(amount 归一 / timeout 恒入签 / get-address 签名串)
exchange.go            汇率(5 endpoints)
energy.go              TRON 能量租赁(8 endpoints)
smtp.go                SMTP 聚合 API(6 endpoints)
withdraw.go            提币(4 endpoints,RSA 签名)
account.go             账户(2 endpoints)
vcard.go               虚拟信用卡(12 endpoints)
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
    var est struct { SuggestedEnergy int `json:"suggested_energy"` }
    _ = json.Unmarshal(raw, &est)
    fmt.Println("建议能量:", est.SuggestedEnergy)

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
| `GetUserAddress(userID)` | POST | `/api/v1/pay/get-address` |
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

注:`GetRates` 的 `base` 传空字符串时由后端取默认(USDT).

### `client.Energy` — TRON 能量租赁(8 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `GetInfo()` | GET | `/api/v1/energy/info` |
| `GetPrice(energyAmount, period)` | GET | `/api/v1/energy/price?energy_amount=&period=` |
| `EstimateEnergy(toAddress)` | GET | `/api/v1/energy/estimate-energy?to_address=` |
| `CreateOrder(params)` | POST | `/api/v1/energy/order` |
| `CreateOnetimeOrder(params)` | POST | `/api/v1/energy/order/onetime` |
| `QueryOrder(serial)` | GET | `/api/v1/energy/order/:serial` |
| `ListOrders(filter)` | GET | `/api/v1/energy/orders` |
| `ReclaimOrder(serial)` | POST | `/api/v1/energy/order/reclaim` |

注:租期 `period` 枚举 `1H / 1D / 3D / 7D / 30D`;`CreateOrder` 必填 `receive_address` / `energy_amount` / `period`,可选 `out_trade_no` / `remark`.

### `client.SMTP` — SMTP 聚合(6 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `Send(params, opts...)` | POST | `/api/v1/smtp/send` |
| `SendBatch(params, opts...)` | POST | `/api/v1/smtp/send/batch` |
| `SendTemplate(params)` | POST | `/api/v1/smtp/send/template` |
| `GetQuota()` | GET | `/api/v1/smtp/quota` |
| `GetStatus(messageID)` | GET | `/api/v1/smtp/status/:message_id` |
| `ReportInbound(params)` | POST | `/api/v1/smtp/inbound` |

- `Send` 可选字段:`from_name` / `reply_to` / `text_body` / `headers` / `cc` / `bcc` / `attachments` / `account_id` / `send_at`(定时,RFC3339);`opts`(`SMTPSendOptions`)可带 IdempotencyKey 写入 `Idempotency-Key` 幂等头
- `SendBatch` 必填 `recipients` 数组(元素 `{to, variables?, from_name?}`),静态 `subject`+`body` 或 `template_code` 二选一;同样支持幂等 opts
- `SendTemplate` 必填 `to` + `template_code`,可选 `variables` / `from_name`
- `GetQuota` 返回 `daily_limit/daily_used/daily_remaining` / `monthly_*` / `expire_at`
- `ReportInbound` 上报退信/投诉(`email` 与 `message_id` 至少其一,`type` = `bounce` | `complaint`)

### `client.Withdraw` — 提币(4 endpoint,RSA-PKCS1v15-SHA256 签名)

| 方法 | HTTP | endpoint |
|---|---|---|
| `CreateWithdraw(params)` | POST | `/api/v1/withdraw` |
| `GetWithdraw(id)` | GET | `/api/v1/withdraw/:id` |
| `GetWithdrawableBalance()` | GET | `/api/v1/balance/withdrawable` |
| `QuoteFee(chain, symbol, amount)` | GET | `/api/v1/fee/quote`(amount 必填) |
| `Sign(...)` / `VerifyCallback(...)` | (工具) | RSA 签名 / 平台回调验签 |

### `client.Account` — 账户(2 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `GetBalance()` | GET | `/api/v1/account/balance` |
| `GetDepositAddress()` | GET | `/api/v1/account/deposit-address` |

### `client.VCard` — 虚拟信用卡(12 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `GetInfo()` / `ListBins()` / `ListCards()` | GET | `/api/v1/vcard/*`(读,X-API-Key) |
| `GetCardTransactions(cardID)` / `ListOrders(query)` / `GetOrder(orderID)` | GET | 同上 |
| `UpdateCardRemark(cardID, remark)` | POST | 同上 |
| `GetCardDetails(cardID)` / `GetCardCode(cardID)` | GET | 敏感读(HMAC 头签名) |
| `OpenCard(params)` / `RechargeCard(cardID, params)` / `CancelCard(cardID)` | POST | 资金操作(HMAC 头签名) |

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
