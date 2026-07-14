// Package nexcore is the official Go SDK for the Tovanix platform (formerly NexCore).
//
// 一次配置覆盖 Tovanix 平台全部 v1 公开接口,业务按 namespace 划分:
//
//   - client.Payment   — 多链收款(HMAC-SHA256 签名)
//   - client.Withdraw  — 多链收款 · 提币(RSA-2048 签名)
//   - client.Exchange  — 汇率(X-App-Key + X-App-Secret header)
//   - client.Energy    — TRON 能量租赁(X-API-Key + X-Secret-Key)
//   - client.SMTP      — SMTP 聚合(Bearer Token)
//   - client.Account   — 账户余额 / 充值地址(X-API-Key + X-Secret-Key)
//   - client.VCard     — 虚拟信用卡(读用双密钥;开卡/充值/注销/敏感信息用 HMAC 签名)
//
// 使用:
//
//	import nexcore "github.com/DoBestone/nexcore-sdk/go"
//
//	c := nexcore.NewClient(nexcore.Config{
//	    BaseURL:               "https://your-domain.com",
//	    PaymentAppID:          "APP20260412XXXX",
//	    PaymentAppKey:         "your_app_key_here",
//	    EnergyAPIKey:          "energy_key",
//	    EnergySecretKey:       "energy_secret",
//	    SMTPAPIKey:            "smk_xxx",
//	    WithdrawAPIKey:        "MPK_xxx",
//	    WithdrawPrivateKeyPEM: os.Getenv("WITHDRAW_RSA_PRIV"),
//	})
//
//	raw, err := c.Payment.CreateOrder(map[string]any{
//	    "out_order_id": fmt.Sprintf("ORDER_%d", time.Now().Unix()),
//	    "amount":       "100.00",
//	    "currency":     "CNY",
//	    "trade_type":   "usdt.trc20",
//	    "call_type":    "rotation",
//	})
//
// 所有方法返回 json.RawMessage,业务方自行 json.Unmarshal 到具体 struct,
// 这样后端 API 加字段不需要升级 SDK.
//
// 所有错误统一返回 *nexcore.Error(含 Code / Message / RequestID / HTTPStatus).
// 使用 nexcore.AsError(err) 把通用 error 转换成 *Error.
package nexcore

// Version is the SDK version, kept in sync with the public repository tags.
const Version = "3.2.0"
