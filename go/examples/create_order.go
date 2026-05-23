// NexCore Go SDK — 创建支付订单(轮播模式)
//
// 运行:
//   cd sdk/go && go run examples/create_order.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	nexcore "github.com/nexcore-platform/nexcore-sdk-go"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	c := nexcore.NewClient(nexcore.Config{
		BaseURL:       envOr("NEXCORE_BASE_URL", "https://your-domain.com"),
		PaymentAppID:  envOr("NEXCORE_APP_ID", "APP20260412XXXX"),
		PaymentAppKey: envOr("NEXCORE_APP_KEY", "your_app_key_here"),
	})

	raw, err := c.Payment.CreateOrder(map[string]any{
		"out_order_id": fmt.Sprintf("ORDER_%d", time.Now().Unix()),
		"amount":       "100.00",          // 必填:法币金额(string,两位小数)
		"currency":     "CNY",             // CNY / USD / EUR / JPY / KRW / HKD
		"trade_type":   "usdt.trc20",      // 加密币种.链
		"call_type":    "rotation",        // rotation=轮播 / one_to_one=一对一
		"timeout":      1800,
		"subject":      "会员充值",
		"notify_url":   "https://your-domain.com/payment/notify",
		"return_url":   "https://your-domain.com/payment/success",
	})
	if err != nil {
		if ne := nexcore.AsError(err); ne != nil {
			log.Fatalf("❌ Error #%d: %s (trace=%s)", ne.Code, ne.Message, ne.RequestID)
		}
		log.Fatal(err)
	}

	var order struct {
		OrderID        string `json:"order_id"`
		PayAddress     string `json:"pay_address"`
		CryptoAmount   string `json:"crypto_amount"`
		CryptoCurrency string `json:"crypto_currency"`
		ExpiresAt      string `json:"expires_at"`
	}
	if err := json.Unmarshal(raw, &order); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ 订单创建成功")
	fmt.Printf("  订单号:    %s\n", order.OrderID)
	fmt.Printf("  支付地址:  %s\n", order.PayAddress)
	fmt.Printf("  加密金额:  %s %s\n", order.CryptoAmount, order.CryptoCurrency)
	fmt.Printf("  过期时间:  %s\n", order.ExpiresAt)
}
