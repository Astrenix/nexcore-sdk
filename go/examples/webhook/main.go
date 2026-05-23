// NexCore Go SDK — Webhook 回调签名校验
//
// 部署:
//   cd sdk/go && go run examples/webhook.go
//
// 然后在 NexCore 用户后台「应用配置」的 notify_url 填你的 URL。
//
// NexCore 支付成功后会 POST JSON 到这里,本示例:
//   1. 校验签名(SDK 一行搞定,内部用 hmac.Equal 常量时间比较)
//   2. 业务处理(发货 / 更新 DB,务必幂等)
//   3. 返回 200 OK(否则平台会重试)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	nexcore "github.com/DoBestone/nexcore-sdk/go"
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

	http.HandleFunc("/payment/notify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		// 1. 校验签名(常量时间比较,防时序攻击)
		if !c.Payment.VerifyNotifySign(payload) {
			log.Printf("[nexcore] sign 校验失败: %v", payload)
			http.Error(w, "invalid sign", http.StatusBadRequest)
			return
		}

		// 2. 业务处理(示例)
		// 同一订单可能因网络重试收到多次回调,务必做幂等(DB 唯一索引 out_order_id 等)
		outOrder, _ := payload["out_order_id"].(string)
		amount, _ := payload["amount"].(string)
		txHash, _ := payload["tx_hash"].(string)
		var status int
		switch v := payload["status"].(type) {
		case float64:
			status = int(v)
		case int:
			status = v
		}

		// 状态:1=已支付  2=待支付  3=已关闭  4=已退款
		if status == 1 {
			log.Printf("[nexcore] 订单已支付: %s = %s (tx: %s)", outOrder, amount, txHash)
			// TODO: DB 查 out_order_id,判断是否已发货,未发货才发货
		}

		// 3. 必须返回 200
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	port := envOr("PORT", "8000")
	fmt.Printf("nexcore webhook listening on :%s/payment/notify\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
