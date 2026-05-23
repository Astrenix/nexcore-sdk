# NexCore Go SDK

全能 Go 客户端,覆盖 Payment / Energy / SMTP / AI 全部 NexCore 业务。**零依赖**(仅标准库 `net/http` / `crypto/hmac`)。

## 环境

- Go 1.21+

## 安装

```bash
go get github.com/nexcore-platform/nexcore-sdk-go
```

(SDK 发布到独立公开仓库后)

## 用法

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    nexcore "github.com/nexcore-platform/nexcore-sdk-go"
)

func main() {
    c := nexcore.NewClient(nexcore.Config{
        BaseURL:         "https://your-domain.com",
        PaymentAppID:    "APP20260412XXXX",
        PaymentAppKey:   "your_app_key_here",
        EnergyAPIKey:    "energy_api_key_here",
        EnergySecretKey: "energy_secret_key_here",
        AIAPIKey:        "sk-nc-xxx",
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
        Amount     string `json:"amount"`
    }
    _ = json.Unmarshal(raw, &order)
    fmt.Println("支付地址:", order.PayAddress)

    // 估算能量
    rawEst, _ := c.Energy.EstimateEnergy("TXxxxxxxxxxxxxxxxxxxxxx")
    var est struct {
        EstimatedEnergy int `json:"estimated_energy"`
    }
    _ = json.Unmarshal(rawEst, &est)
    fmt.Println("需要能量:", est.EstimatedEnergy)

    // AI 对话
    rawReply, err := c.AI.Chat(
        []nexcore.Message{{Role: "user", Content: "你好"}},
        "claude-opus-4-7",
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(rawReply))
}
```

## 异常

所有错误统一返回 `*nexcore.Error`,字段:

- `Code` — 平台错误码(0 = 成功)
- `Message` — 错误描述
- `RequestID` — 服务端日志追踪 ID(响应头 `X-Trace-Id`)
- `HTTPStatus` — HTTP 状态码

使用 `nexcore.AsError(err)` 把通用 `error` 转成 `*nexcore.Error`。

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
    // 处理回调...
    w.Write([]byte("OK"))
})
```

签名校验用 `hmac.Equal`,常量时间比较,防时序攻击。

## 返回值约定

所有方法返回 `json.RawMessage` — 业务方自行 `json.Unmarshal` 到具体 struct,**避免 SDK 跟后端 API 字段强耦合**(后端加字段不需要升级 SDK)。

## 示例

更多示例见 [`examples/`](./examples/) 目录。
