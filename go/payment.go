package nexcore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// PaymentNamespace implements all v1 endpoints under "/api/v1/pay/".
//
// 对应 /docs 文档 "多链收款" 模块的全部 7 个 v1 公开接口
// (对照 internal/handler/order.go + one_to_one.go):
//
//	POST /api/v1/pay/create          CreateOrder       创建收款订单
//	GET  /api/v1/pay/query           QueryOrder        查询订单状态
//	POST /api/v1/pay/close           CloseOrder        关闭订单
//	GET  /api/v1/pay/app-config      GetAppConfig      查询应用配置
//	POST /api/v1/pay/bind-address    BindAddress       一对一 — 绑定地址
//	POST /api/v1/pay/get-address     GetUserAddress    一对一 — 查询用户已绑地址
//	POST /api/v1/pay/unbind-address  UnbindAddress     一对一 — 解绑
//
// 另提供 VerifyNotifySign() 校验 webhook 回调签名(常量时间比较).
//
// 鉴权:HMAC-SHA256 签名 — 所有请求自动追加 app_id + sign 字段.
type PaymentNamespace struct {
	c *Client
}

// Sign computes HMAC-SHA256 signature for the given params.
//
// 业务方一般不需要直接调,SDK 内部自动调用.公开出来便于:
//   - 自行测试签名是否正确(对照 /docs 文档输出)
//   - 校验回调签名(VerifyNotifySign 内部也用)
//
// 签名算法:把所有非空、非 sign 的参数按 key 升序拼接成 k1=v1&k2=v2,
// 然后用 PaymentAppKey 做 HMAC-SHA256,返回 64 字符小写 hex.
func (n *PaymentNamespace) Sign(params map[string]any) (string, error) {
	if n.c.cfg.PaymentAppKey == "" {
		return "", &Error{Message: "PaymentAppKey not configured", Code: -1}
	}
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" {
			continue
		}
		s := fmt.Sprint(v)
		if s == "" || s == "<nil>" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(params[k]))
	}
	mac := hmac.New(sha256.New, []byte(n.c.cfg.PaymentAppKey))
	mac.Write([]byte(b.String()))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// signed 自动注入 app_id + 计算 sign,返回签好的参数.
func (n *PaymentNamespace) signed(params map[string]any) (map[string]any, error) {
	if n.c.cfg.PaymentAppID == "" {
		return nil, &Error{Message: "PaymentAppID not configured", Code: -1}
	}
	out := make(map[string]any, len(params)+2)
	for k, v := range params {
		out[k] = v
	}
	out["app_id"] = n.c.cfg.PaymentAppID
	sig, err := n.Sign(out)
	if err != nil {
		return nil, err
	}
	out["sign"] = sig
	return out, nil
}

// CreateOrder creates a payment order.
//
// POST /api/v1/pay/create
//
// 必填字段:
//
//	out_order_id (string) — 商户侧订单号,必须唯一
//	amount       (string) — 法币金额,推荐两位小数 string 避免浮点误差
//	currency     (string) — 法币代码 CNY/USD/EUR/JPY/KRW/HKD
//	trade_type   (string) — 加密币种.链,如 "usdt.trc20"
//
// 可选字段:
//
//	call_type    (string) — "rotation"(轮播)或 "one_to_one",默认 rotation
//	user_id      (string) — 一对一模式必填
//	timeout      (int)    — 订单过期秒数,默认 1800
//	subject      (string) — 订单描述
//	notify_url   (string) — webhook 回调 URL
//	return_url   (string) — 支付成功跳转 URL
//
// 返回 {order_id, pay_address, crypto_amount, crypto_currency, expires_at, ...}.
func (n *PaymentNamespace) CreateOrder(params map[string]any) (json.RawMessage, error) {
	p, err := n.signed(params)
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/pay/create", &requestOpts{Body: p})
}

// QueryOrder queries an order by merchant out_order_id.
//
// GET /api/v1/pay/query
func (n *PaymentNamespace) QueryOrder(outOrderID string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"out_order_id": outOrderID})
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/pay/query", &requestOpts{Query: p})
}

// CloseOrder closes an open order.
//
// POST /api/v1/pay/close
func (n *PaymentNamespace) CloseOrder(outOrderID string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"out_order_id": outOrderID})
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/pay/close", &requestOpts{Body: p})
}

// GetAppConfig returns current application configuration.
//
// GET /api/v1/pay/app-config
func (n *PaymentNamespace) GetAppConfig() (json.RawMessage, error) {
	p, err := n.signed(map[string]any{})
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/pay/app-config", &requestOpts{Query: p})
}

// BindAddress binds a user to a fixed receiving address (one-to-one mode).
//
// POST /api/v1/pay/bind-address
func (n *PaymentNamespace) BindAddress(userID, tradeType string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"user_id": userID, "trade_type": tradeType})
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/pay/bind-address", &requestOpts{Body: p})
}

// GetUserAddress queries the address bound to a user (one-to-one mode).
//
// POST /api/v1/pay/get-address (注意:后端是 POST,不是 GET)
func (n *PaymentNamespace) GetUserAddress(userID, tradeType string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"user_id": userID, "trade_type": tradeType})
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/pay/get-address", &requestOpts{Body: p})
}

// UnbindAddress unbinds a user's address (one-to-one mode).
//
// POST /api/v1/pay/unbind-address
func (n *PaymentNamespace) UnbindAddress(userID string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"user_id": userID})
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/pay/unbind-address", &requestOpts{Body: p})
}

// VerifyNotifySign verifies a webhook notification signature with constant-time comparison.
//
// NexCore 平台通过 notify_url 推送 JSON 通知时会带 sign 字段.
// 本方法用 hmac.Equal 常量时间比较防止时序攻击.
//
//	返回 true=签名正确,可信
//	    false=签名错误/缺失,应拒绝该回调
func (n *PaymentNamespace) VerifyNotifySign(payload map[string]any) bool {
	signAny, ok := payload["sign"]
	if !ok {
		return false
	}
	sign, ok := signAny.(string)
	if !ok || sign == "" {
		return false
	}
	expected, err := n.Sign(payload)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(sign))
}
