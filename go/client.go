// Package nexcore is the official Go SDK for NexCore platform — an all-in-one
// client covering Payment / Energy / SMTP / AI API business modules.
//
// Usage:
//
//	import "github.com/your-org/nexcore-sdk"
//
//	c := nexcore.NewClient(nexcore.Config{
//	    BaseURL:         "https://your-domain.com",
//	    PaymentAppID:    "APP20260412XXXX",
//	    PaymentAppKey:   "your_app_key_here",
//	    EnergyAPIKey:    "energy_api_key_here",
//	    EnergySecretKey: "energy_secret_key_here",
//	    AIAPIKey:        "sk-nc-xxx",
//	})
//
//	// 创建支付订单
//	order, err := c.Payment.CreateOrder(map[string]any{
//	    "out_order_id": fmt.Sprintf("ORDER_%d", time.Now().Unix()),
//	    "amount":       "100.00",
//	    "currency":     "CNY",
//	    "trade_type":   "usdt.trc20",
//	    "call_type":    "rotation",
//	})
//
//	// AI chat
//	reply, err := c.AI.Chat([]nexcore.Message{{Role: "user", Content: "Hello"}}, "claude-opus-4-7", nil)
package nexcore

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Error 是 SDK 统一的业务错误.
type Error struct {
	Message    string
	Code       int
	RequestID  string
	HTTPStatus int
}

func (e *Error) Error() string {
	return fmt.Sprintf("nexcore: %s (code=%d, http=%d, trace=%s)", e.Message, e.Code, e.HTTPStatus, e.RequestID)
}

// Config 客户端配置.
type Config struct {
	BaseURL string

	PaymentAppID  string
	PaymentAppKey string

	EnergyAPIKey    string
	EnergySecretKey string

	SMTPAPIKey string

	AIAPIKey string

	Timeout time.Duration // 默认 30s
}

// Client 全能客户端.
type Client struct {
	cfg  Config
	http *http.Client

	Payment *PaymentNamespace
	Energy  *EnergyNamespace
	SMTP    *SMTPNamespace
	AI      *AINamespace
}

// NewClient 创建客户端实例.
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	c := &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
	c.Payment = &PaymentNamespace{c: c}
	c.Energy = &EnergyNamespace{c: c}
	c.SMTP = &SMTPNamespace{c: c}
	c.AI = &AINamespace{c: c}
	return c
}

// request 底层 HTTP 调用,返回 data 段.
func (c *Client) request(method, path string, opts *requestOpts) (json.RawMessage, error) {
	if opts == nil {
		opts = &requestOpts{}
	}
	urlStr := c.cfg.BaseURL + path
	if len(opts.Query) > 0 {
		q := url.Values{}
		for k, v := range opts.Query {
			s := fmt.Sprint(v)
			if s == "" || s == "<nil>" {
				continue
			}
			q.Set(k, s)
		}
		if enc := q.Encode(); enc != "" {
			if strings.Contains(urlStr, "?") {
				urlStr += "&" + enc
			} else {
				urlStr += "?" + enc
			}
		}
	}

	var body io.Reader
	if opts.Body != nil {
		b, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, &Error{Message: "marshal body: " + err.Error(), Code: -1}
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(strings.ToUpper(method), urlStr, body)
	if err != nil {
		return nil, &Error{Message: "build request: " + err.Error(), Code: -1}
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &Error{Message: "http: " + err.Error(), Code: -1}
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	traceID := resp.Header.Get("X-Trace-Id")

	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if jsonErr := json.Unmarshal(raw, &env); jsonErr != nil {
		return nil, &Error{
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200)),
			Code:       -1, RequestID: traceID, HTTPStatus: resp.StatusCode,
		}
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		return nil, &Error{
			Message: env.Message, Code: env.Code, RequestID: traceID, HTTPStatus: resp.StatusCode,
		}
	}
	if len(env.Data) > 0 {
		return env.Data, nil
	}
	return raw, nil
}

type requestOpts struct {
	Body    any
	Query   map[string]any
	Headers map[string]string
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (c *Client) need(field, name string) (string, error) {
	if field == "" {
		return "", &Error{Message: name + " not configured", Code: -1}
	}
	return field, nil
}

// =========================== Payment ===========================

// PaymentNamespace 链收款 — HMAC-SHA256 签名.
type PaymentNamespace struct{ c *Client }

// Sign 计算业务参数签名(给业务层手动调用 / webhook 校验用).
func (n *PaymentNamespace) Sign(params map[string]any) (string, error) {
	key, err := n.c.need(n.c.cfg.PaymentAppKey, "PaymentAppKey")
	if err != nil {
		return "", err
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
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(b.String()))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (n *PaymentNamespace) signed(params map[string]any) (map[string]any, error) {
	appID, err := n.c.need(n.c.cfg.PaymentAppID, "PaymentAppID")
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(params)+2)
	for k, v := range params {
		out[k] = v
	}
	out["app_id"] = appID
	sig, err := n.Sign(out)
	if err != nil {
		return nil, err
	}
	out["sign"] = sig
	return out, nil
}

// CreateOrder 创建支付订单.
func (n *PaymentNamespace) CreateOrder(params map[string]any) (json.RawMessage, error) {
	p, err := n.signed(params)
	if err != nil {
		return nil, err
	}
	return n.c.request("POST", "/api/v1/pay/create", &requestOpts{Body: p})
}

// QueryOrder 查询订单.
func (n *PaymentNamespace) QueryOrder(outOrderID string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"out_order_id": outOrderID})
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/pay/query", &requestOpts{Query: p})
}

// CloseOrder 关闭订单.
func (n *PaymentNamespace) CloseOrder(outOrderID string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"out_order_id": outOrderID})
	if err != nil {
		return nil, err
	}
	return n.c.request("POST", "/api/v1/pay/close", &requestOpts{Body: p})
}

// BindAddress 绑定一对一收款地址.
func (n *PaymentNamespace) BindAddress(userID, tradeType string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"user_id": userID, "trade_type": tradeType})
	if err != nil {
		return nil, err
	}
	return n.c.request("POST", "/api/v1/pay/bind-address", &requestOpts{Body: p})
}

// GetAddress 查询用户绑定的地址.
func (n *PaymentNamespace) GetAddress(userID, tradeType string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"user_id": userID, "trade_type": tradeType})
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/pay/get-address", &requestOpts{Query: p})
}

// UnbindAddress 解绑地址.
func (n *PaymentNamespace) UnbindAddress(userID string) (json.RawMessage, error) {
	p, err := n.signed(map[string]any{"user_id": userID})
	if err != nil {
		return nil, err
	}
	return n.c.request("POST", "/api/v1/pay/unbind-address", &requestOpts{Body: p})
}

// AppConfig 查询应用配置.
func (n *PaymentNamespace) AppConfig() (json.RawMessage, error) {
	p, err := n.signed(map[string]any{})
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/pay/app-config", &requestOpts{Query: p})
}

// VerifyNotifySign 校验 webhook 回调签名(常量时间比较).
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

// =========================== Energy ===========================

// EnergyNamespace TRON 能量租赁.
type EnergyNamespace struct{ c *Client }

func (n *EnergyNamespace) headers() (map[string]string, error) {
	k, err := n.c.need(n.c.cfg.EnergyAPIKey, "EnergyAPIKey")
	if err != nil {
		return nil, err
	}
	s, err := n.c.need(n.c.cfg.EnergySecretKey, "EnergySecretKey")
	if err != nil {
		return nil, err
	}
	return map[string]string{"X-API-Key": k, "X-Secret-Key": s}, nil
}

func (n *EnergyNamespace) Info() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/energy/info", &requestOpts{Headers: h})
}

func (n *EnergyNamespace) Price(energy int, period string) (json.RawMessage, error) {
	if period == "" {
		period = "1D"
	}
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/energy/price", &requestOpts{
		Query: map[string]any{"energy": energy, "period": period}, Headers: h,
	})
}

func (n *EnergyNamespace) EstimateEnergy(receiveAddr string) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/energy/estimate-energy", &requestOpts{
		Query: map[string]any{"receive_addr": receiveAddr}, Headers: h,
	})
}

func (n *EnergyNamespace) CreateOrder(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("POST", "/api/v1/energy/order", &requestOpts{Body: params, Headers: h})
}

func (n *EnergyNamespace) QueryOrder(orderID int64) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", fmt.Sprintf("/api/v1/energy/order/%d", orderID), &requestOpts{Headers: h})
}

func (n *EnergyNamespace) ListOrders(filter map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/energy/orders", &requestOpts{Query: filter, Headers: h})
}

// =========================== SMTP ===========================

// SMTPNamespace SMTP 聚合 API.
type SMTPNamespace struct{ c *Client }

func (n *SMTPNamespace) headers() (map[string]string, error) {
	k, err := n.c.need(n.c.cfg.SMTPAPIKey, "SMTPAPIKey")
	if err != nil {
		return nil, err
	}
	return map[string]string{"X-API-Key": k}, nil
}

func (n *SMTPNamespace) SendMail(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("POST", "/api/v1/smtp/send", &requestOpts{Body: params, Headers: h})
}

func (n *SMTPNamespace) ListAccounts() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/smtp/accounts", &requestOpts{Headers: h})
}

func (n *SMTPNamespace) ListTemplates() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/api/v1/smtp/templates", &requestOpts{Headers: h})
}

// =========================== AI ===========================

// Message LLM chat 消息.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AINamespace Astrenix AI(OpenAI 兼容).
type AINamespace struct{ c *Client }

func (n *AINamespace) headers() (map[string]string, error) {
	k, err := n.c.need(n.c.cfg.AIAPIKey, "AIAPIKey")
	if err != nil {
		return nil, err
	}
	return map[string]string{"Authorization": "Bearer " + k}, nil
}

// Chat 调用 LLM 对话.
func (n *AINamespace) Chat(messages []Message, model string, extra map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	body := map[string]any{"model": model, "messages": messages}
	for k, v := range extra {
		body[k] = v
	}
	return n.c.request("POST", "/v1/chat/completions", &requestOpts{Body: body, Headers: h})
}

// Models 列出可用模型.
func (n *AINamespace) Models() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.request("GET", "/v1/models", &requestOpts{Headers: h})
}

// AsError 把 error 转成 *Error,方便业务层拿 Code/RequestID.
func AsError(err error) *Error {
	var ne *Error
	if errors.As(err, &ne) {
		return ne
	}
	return nil
}
