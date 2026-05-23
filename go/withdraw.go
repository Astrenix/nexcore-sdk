package nexcore

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WithdrawNamespace implements the v1 提币 API — 多链收款业务的资金出库端.
//
// 鉴权:RSA-PKCS1v15-SHA256 签名 + 4 个请求头
//
//	X-API-Key            账户级 API Key(控制台「账号 → API 密钥」)
//	X-Timestamp          unix ms,与服务器时差 ≤ 60s
//	X-Nonce              一次性 nonce(uuid v4),5 分钟内不可重复
//	X-Withdraw-Signature RSA-PKCS1v15-SHA256(caller_private_key, signString),Base64
//
// signString = METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + BODY
// 其中 BODY 为 HTTP body 原文(JSON 字符串原样,GET 请求为空字符串).
//
// 对应 /docs 文档 "提币 API" 章节的 4 个 endpoint(internal/handler/api_withdraw_v1.go):
//
//	POST /api/v1/withdraw                 CreateWithdraw       发起提币
//	GET  /api/v1/withdraw/:id             GetWithdraw          查询单笔状态
//	GET  /api/v1/balance/withdrawable     GetWithdrawableBalance 查询可提余额
//	GET  /api/v1/fee/quote                QuoteFee             费用预估
//
// 另提供 VerifyCallback() 校验平台回调签名(用平台公钥).
type WithdrawNamespace struct {
	c *Client

	// 私钥/公钥懒解析缓存,避免每次请求重复解码 PEM
	once       sync.Once
	privKey    *rsa.PrivateKey
	privParseErr error
	platformPubOnce sync.Once
	platformPub *rsa.PublicKey
	platformPubErr error
}

// parsePrivKey 解析配置的对接方私钥 PEM,带缓存.
func (n *WithdrawNamespace) parsePrivKey() (*rsa.PrivateKey, error) {
	n.once.Do(func() {
		if n.c.cfg.WithdrawPrivateKeyPEM == "" {
			n.privParseErr = &Error{Message: "WithdrawPrivateKeyPEM not configured", Code: -1}
			return
		}
		block, _ := pem.Decode([]byte(n.c.cfg.WithdrawPrivateKeyPEM))
		if block == nil {
			n.privParseErr = &Error{Message: "withdraw: invalid private key PEM", Code: -1}
			return
		}
		// 兼容 PKCS#1 和 PKCS#8
		if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			n.privKey = key
			return
		}
		if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if rsaKey, ok := k.(*rsa.PrivateKey); ok {
				n.privKey = rsaKey
				return
			}
			n.privParseErr = &Error{Message: "withdraw: PKCS#8 key is not RSA", Code: -1}
			return
		}
		n.privParseErr = &Error{Message: "withdraw: cannot parse private key (neither PKCS#1 nor PKCS#8)", Code: -1}
	})
	return n.privKey, n.privParseErr
}

// parsePlatformPub 解析平台公钥(回调验签用).
func (n *WithdrawNamespace) parsePlatformPub() (*rsa.PublicKey, error) {
	n.platformPubOnce.Do(func() {
		if n.c.cfg.WithdrawPlatformPublicKeyPEM == "" {
			n.platformPubErr = &Error{Message: "WithdrawPlatformPublicKeyPEM not configured", Code: -1}
			return
		}
		block, _ := pem.Decode([]byte(n.c.cfg.WithdrawPlatformPublicKeyPEM))
		if block == nil {
			n.platformPubErr = &Error{Message: "withdraw: invalid platform public key PEM", Code: -1}
			return
		}
		if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
			if rsaKey, ok := key.(*rsa.PublicKey); ok {
				n.platformPub = rsaKey
				return
			}
			n.platformPubErr = &Error{Message: "withdraw: PKIX key is not RSA", Code: -1}
			return
		}
		// 兼容 PKCS#1
		if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
			n.platformPub = key
			return
		}
		n.platformPubErr = &Error{Message: "withdraw: cannot parse platform public key", Code: -1}
	})
	return n.platformPub, n.platformPubErr
}

// Sign computes the RSA-PKCS1v15-SHA256 signature for a withdraw request.
//
// 业务方一般不需要直接调,SDK 内部 do 时自动调用.公开出来便于:
//   - 测试签名正确性
//   - 自行实现非标场景(比如 curl 调试)
func (n *WithdrawNamespace) Sign(method, path, timestamp, nonce, body string) (string, error) {
	priv, err := n.parsePrivKey()
	if err != nil {
		return "", err
	}
	signString := strings.ToUpper(method) + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + body
	h := sha256.Sum256([]byte(signString))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		return "", &Error{Message: "withdraw: RSA sign failed: " + err.Error(), Code: -1}
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// newNonce 生成 uuid v4(本地实现避免引入第三方依赖).
func newNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// RFC 4122 v4 variant
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// do 内部统一发请求 — 自动加 4 个鉴权头.
//
// body 必须传入"已序列化"的 JSON bytes(确保和签名串里的 BODY 字符串完全一致).
// GET 请求 body 传 nil 即可.
func (n *WithdrawNamespace) do(method, path string, body []byte, query map[string]any) (json.RawMessage, error) {
	if n.c.cfg.WithdrawAPIKey == "" {
		return nil, &Error{Message: "WithdrawAPIKey not configured", Code: -1}
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	nonce, err := newNonce()
	if err != nil {
		return nil, &Error{Message: "withdraw: gen nonce: " + err.Error(), Code: -1}
	}
	bodyStr := ""
	if body != nil {
		bodyStr = string(body)
	}
	sig, err := n.Sign(method, path, timestamp, nonce, bodyStr)
	if err != nil {
		return nil, err
	}
	return n.c.transport.do(method, path, &requestOpts{
		BodyRaw: body,
		Query:   query,
		Headers: map[string]string{
			"X-API-Key":             n.c.cfg.WithdrawAPIKey,
			"X-Timestamp":           timestamp,
			"X-Nonce":               nonce,
			"X-Withdraw-Signature":  sig,
		},
	})
}

// CreateWithdraw 发起提币 — POST /api/v1/withdraw
//
// 下单后状态为 pending,等延迟到期由 worker 自动广播.期间可在控制台暂停 / 加速 / 取消.
//
//	params := map[string]any{
//	    "chain":        "tron",
//	    "symbol":       "USDT",
//	    "amount":       "100.5",
//	    "to_address":   "TXxxxxxxxx",
//	    "memo":         "withdraw to user #1024",  // 可选
//	    "callback_url": "https://your-domain.com/cb", // 可选
//	    "request_id":   "your-idempotency-uuid",   // 可选,推荐传
//	}
//	raw, err := client.Withdraw.CreateWithdraw(params)
func (n *WithdrawNamespace) CreateWithdraw(params map[string]any) (json.RawMessage, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, &Error{Message: "withdraw: marshal body: " + err.Error(), Code: -1}
	}
	return n.do("POST", "/api/v1/withdraw", body, nil)
}

// GetWithdraw 查询单笔提币状态 — GET /api/v1/withdraw/:id
//
// 返回订单详情(可用来轮询状态,也建议优先用回调).
func (n *WithdrawNamespace) GetWithdraw(id string) (json.RawMessage, error) {
	if id == "" {
		return nil, &Error{Message: "withdraw: id is required", Code: -1}
	}
	return n.do("GET", "/api/v1/withdraw/"+id, nil, nil)
}

// GetWithdrawableBalance 查询可提余额 — GET /api/v1/balance/withdrawable
//
// 返回该账户在每条链 × 每种资产下的「已归集待提现」余额.
// 只有这部分可用于 API 提币.
func (n *WithdrawNamespace) GetWithdrawableBalance() (json.RawMessage, error) {
	return n.do("GET", "/api/v1/balance/withdrawable", nil, nil)
}

// QuoteFee 费用预估 — GET /api/v1/fee/quote?chain=&symbol=&amount=
//
// 返回管理端为该 chain × symbol 配置的预扣费(OKX 式固定值).
//
//	raw, err := client.Withdraw.QuoteFee("tron", "USDT", "100")
func (n *WithdrawNamespace) QuoteFee(chain, symbol, amount string) (json.RawMessage, error) {
	if chain == "" || symbol == "" {
		return nil, &Error{Message: "withdraw: chain and symbol are required", Code: -1}
	}
	q := map[string]any{"chain": chain, "symbol": symbol}
	if amount != "" {
		q["amount"] = amount
	}
	return n.do("GET", "/api/v1/fee/quote", nil, q)
}

// VerifyCallback 验证平台回调签名.
//
// 用法(对接方收到回调时):
//
//	sig := req.Header.Get("X-Platform-Signature")
//	body, _ := io.ReadAll(req.Body)
//	ts := req.Header.Get("X-Timestamp")
//	nonce := req.Header.Get("X-Nonce")
//	if err := client.Withdraw.VerifyCallback(req.Method, req.URL.Path, ts, nonce, body, sig); err != nil {
//	    // 验签失败,拒绝处理
//	}
//
// 验签算法与请求方向一致:RSA-PKCS1v15-SHA256(platform_public_key, signString).
func (n *WithdrawNamespace) VerifyCallback(method, path, timestamp, nonce string, body []byte, base64Sig string) error {
	pub, err := n.parsePlatformPub()
	if err != nil {
		return err
	}
	sig, err := base64.StdEncoding.DecodeString(base64Sig)
	if err != nil {
		return &Error{Message: "withdraw: bad signature base64: " + err.Error(), Code: -1}
	}
	signString := strings.ToUpper(method) + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + string(body)
	h := sha256.Sum256([]byte(signString))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
		return &Error{Message: "withdraw: signature verify failed", Code: -1}
	}
	return nil
}

