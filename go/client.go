package nexcore

import (
	"net/http"
	"strings"
	"time"
)

// Config 客户端配置.
type Config struct {
	// BaseURL NexCore 平台基础 URL,例如 "https://your-domain.com"(必填).
	BaseURL string

	// PaymentAppID 多链收款 / 汇率 应用 ID.
	PaymentAppID string
	// PaymentAppKey 多链收款 / 汇率 应用密钥(HMAC 签名 + X-App-Secret 同用).
	PaymentAppKey string

	// EnergyAPIKey TRON 能量租赁 X-API-Key.
	EnergyAPIKey string
	// EnergySecretKey TRON 能量租赁 X-Secret-Key.
	EnergySecretKey string

	// SMTPAPIKey SMTP 聚合 API 的 smk_ 前缀 Token.
	SMTPAPIKey string

	// Timeout HTTP 超时,默认 30s.
	Timeout time.Duration

	// UserAgent 自定义 User-Agent(可选).
	UserAgent string
}

// Client NexCore 全能客户端.
//
// 业务 namespace 通过字段访问:
//
//	c.Payment   - 多链收款
//	c.Exchange  - 汇率
//	c.Energy    - TRON 能量租赁
//	c.SMTP      - SMTP 聚合
//
// 所有方法返回 json.RawMessage,业务方自行 json.Unmarshal 到具体 struct.
type Client struct {
	cfg       Config
	transport *httpTransport

	// Payment 多链收款命名空间
	Payment *PaymentNamespace
	// Exchange 汇率命名空间
	Exchange *ExchangeNamespace
	// Energy TRON 能量租赁命名空间
	Energy *EnergyNamespace
	// SMTP SMTP 聚合命名空间
	SMTP *SMTPNamespace
}

// NewClient creates a new NexCore client.
//
//	c := nexcore.NewClient(nexcore.Config{
//	    BaseURL: "https://your-domain.com",
//	    PaymentAppID: "APP20260412XXXX",
//	    PaymentAppKey: "your_app_key_here",
//	})
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.UserAgent == "" {
		cfg.UserAgent = "NexCore-Go-SDK/" + Version
	}

	c := &Client{
		cfg: cfg,
		transport: &httpTransport{
			baseURL:    cfg.BaseURL,
			httpClient: &http.Client{Timeout: cfg.Timeout},
			userAgent:  cfg.UserAgent,
		},
	}
	c.Payment = &PaymentNamespace{c: c}
	c.Exchange = &ExchangeNamespace{c: c}
	c.Energy = &EnergyNamespace{c: c}
	c.SMTP = &SMTPNamespace{c: c}
	return c
}
