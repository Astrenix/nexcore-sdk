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

	// WithdrawAPIKey 提币 API 的 X-API-Key(账户级 API Key).
	WithdrawAPIKey string
	// WithdrawPrivateKeyPEM 对接方 RSA 私钥 PEM 字符串(用于请求签名).
	// 推荐运行时从环境变量或密钥管理服务读出,不要硬编码.
	WithdrawPrivateKeyPEM string
	// WithdrawPlatformPublicKeyPEM 平台 RSA 公钥 PEM(用于回调验签,可选).
	WithdrawPlatformPublicKeyPEM string

	// APIKey 商户 API Key(MPK 商户密钥的 key 部分);account 与 vcard 命名空间共用.
	// 双密钥读场景作为 X-API-Key,HMAC 签名写场景作为 X-Key-ID.
	APIKey string
	// APISecret 商户 API Secret(MPK 商户密钥的 secret 部分);account 与 vcard 命名空间共用.
	// 双密钥读场景作为 X-Secret-Key,HMAC 签名写场景作为 HMAC-SHA256 的密钥.
	APISecret string

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
//	c.Account   - 账户(余额 / 充值地址)
//	c.VCard     - 虚拟信用卡
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
	// Withdraw 提币命名空间(多链收款业务的资金出库端,RSA-2048 签名)
	Withdraw *WithdrawNamespace
	// Account 账户命名空间(余额 / 充值地址,双密钥读)
	Account *AccountNamespace
	// VCard 虚拟信用卡命名空间(读用双密钥,开卡/充值/注销/敏感信息用 HMAC 签名)
	VCard *VCardNamespace
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
	c.Withdraw = &WithdrawNamespace{c: c}
	c.Account = &AccountNamespace{c: c}
	c.VCard = &VCardNamespace{c: c}
	return c
}
