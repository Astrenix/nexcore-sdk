package nexcore

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// VCardNamespace implements the v1 虚拟信用卡 endpoints.
//
// 对应 NexCore 虚拟信用卡模块(对照 internal/handler/virtual_card_api.go):
//
// 只读 / 普通写(X-API-Key + X-Secret-Key 双 header):
//
//	GET  /api/v1/vcard/info                       GetInfo              平台公开信息
//	GET  /api/v1/vcard/bins                        ListBins             可开卡 BIN 列表
//	GET  /api/v1/vcard/cards                       ListCards            我的卡片列表
//	GET  /api/v1/vcard/cards/{id}/transactions     GetCardTransactions  卡片消费流水
//	GET  /api/v1/vcard/orders                      ListOrders           订单列表
//	GET  /api/v1/vcard/orders/{id}                 GetOrder             单笔订单
//	PUT  /api/v1/vcard/cards/{id}/remark           UpdateCardRemark     修改卡备注
//
// 资金 / 敏感(HMAC-SHA256 签名 + X-Key-ID/X-Timestamp/X-Nonce/X-Signature):
//
//	GET  /api/v1/vcard/cards/{id}/details          GetCardDetails       卡敏感信息(卡号等)
//	GET  /api/v1/vcard/cards/{id}/code             GetCardCode          CVV / 安全码
//	POST /api/v1/vcard/cards                        OpenCard             开卡
//	POST /api/v1/vcard/cards/{id}/recharge          RechargeCard         充值
//	POST /api/v1/vcard/cards/{id}/cancel            CancelCard           注销
//
// 鉴权密钥来自 cfg.APIKey / cfg.APISecret(MPK 商户密钥),与 account 命名空间共用.
type VCardNamespace struct {
	c *Client
}

// headers 构造 X-API-Key + X-Secret-Key(双密钥读 / 普通写,复用 energy 模式).
func (n *VCardNamespace) headers() (map[string]string, error) {
	if n.c.cfg.APIKey == "" || n.c.cfg.APISecret == "" {
		return nil, &Error{Message: "APIKey / APISecret not configured", Code: -1}
	}
	return map[string]string{
		"X-API-Key":    n.c.cfg.APIKey,
		"X-Secret-Key": n.c.cfg.APISecret,
	}, nil
}

// signedRequest 统一处理 HMAC 签名头 + BodyRaw 发送.
//
// 严格保证签名串里的 body 与实际发出的字节完全一致:bodyBytes 既参与签名,
// 又通过 requestOpts.BodyRaw 原样发出(像 withdraw 那样),杜绝 http 层二次序列化
// 导致签名不匹配.GET 请求 bodyBytes 传 nil,签名里 body 视为空字符串.
//
// 签名公式(与后端 internal/middleware/api_auth.go verifySignature 字节级一致):
//
//	payload = ts + nonce + method + path + rawQuery + body
//	sig     = hex_lower( HMAC_SHA256(APISecret, payload) )
//
// 这些签名接口都不带 query,rawQuery 固定为 "".
func (n *VCardNamespace) signedRequest(method, path string, bodyBytes []byte) (json.RawMessage, error) {
	if n.c.cfg.APIKey == "" || n.c.cfg.APISecret == "" {
		return nil, &Error{Message: "APIKey / APISecret not configured", Code: -1}
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := newHexNonce()
	if err != nil {
		return nil, &Error{Message: "vcard: gen nonce: " + err.Error(), Code: -1}
	}

	const rawQuery = "" // 签名接口均无 query
	body := ""
	if bodyBytes != nil {
		body = string(bodyBytes)
	}

	payload := ts + nonce + strings.ToUpper(method) + path + rawQuery + body
	mac := hmac.New(sha256.New, []byte(n.c.cfg.APISecret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	return n.c.transport.do(method, path, &requestOpts{
		BodyRaw: bodyBytes,
		Headers: map[string]string{
			"X-Key-ID":    n.c.cfg.APIKey,
			"X-Timestamp": ts,
			"X-Nonce":     nonce,
			"X-Signature": sig,
		},
	})
}

// newHexNonce 生成 16 字节(32 hex 字符)随机 nonce,满足后端 nonce 长度要求.
func newHexNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ---------- 双密钥读 / 普通写 ----------

// GetInfo returns vcard platform public info.
//
// GET /api/v1/vcard/info
func (n *VCardNamespace) GetInfo() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/vcard/info", &requestOpts{Headers: h})
}

// ListBins lists open-card BINs (card platforms / tiers).
//
// GET /api/v1/vcard/bins
func (n *VCardNamespace) ListBins() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/vcard/bins", &requestOpts{Headers: h})
}

// ListCards lists the merchant's cards.
//
// GET /api/v1/vcard/cards
func (n *VCardNamespace) ListCards() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/vcard/cards", &requestOpts{Headers: h})
}

// GetCardTransactions returns the consumption transactions of a card.
//
// GET /api/v1/vcard/cards/{id}/transactions
func (n *VCardNamespace) GetCardTransactions(cardID string) (json.RawMessage, error) {
	if cardID == "" {
		return nil, &Error{Message: "vcard: cardID is required", Code: -1}
	}
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", fmt.Sprintf("/api/v1/vcard/cards/%s/transactions", cardID), &requestOpts{Headers: h})
}

// ListOrders lists vcard orders.
//
// GET /api/v1/vcard/orders
//
// query 支持 page / page_size / status / order_type.
func (n *VCardNamespace) ListOrders(query map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/vcard/orders", &requestOpts{Query: query, Headers: h})
}

// GetOrder returns a single vcard order.
//
// GET /api/v1/vcard/orders/{id}
func (n *VCardNamespace) GetOrder(orderID string) (json.RawMessage, error) {
	if orderID == "" {
		return nil, &Error{Message: "vcard: orderID is required", Code: -1}
	}
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", fmt.Sprintf("/api/v1/vcard/orders/%s", orderID), &requestOpts{Headers: h})
}

// UpdateCardRemark updates a card's remark.
//
// PUT /api/v1/vcard/cards/{id}/remark   body {"remark": remark}
//
// 非资金敏感,使用双密钥写(X-API-Key + X-Secret-Key).
func (n *VCardNamespace) UpdateCardRemark(cardID, remark string) (json.RawMessage, error) {
	if cardID == "" {
		return nil, &Error{Message: "vcard: cardID is required", Code: -1}
	}
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("PUT", fmt.Sprintf("/api/v1/vcard/cards/%s/remark", cardID), &requestOpts{
		Body:    map[string]any{"remark": remark},
		Headers: h,
	})
}

// ---------- HMAC 签名(资金 / 敏感) ----------

// GetCardDetails returns the card's sensitive details (full PAN, expiry, etc).
//
// GET /api/v1/vcard/cards/{id}/details  — HMAC 签名鉴权.
func (n *VCardNamespace) GetCardDetails(cardID string) (json.RawMessage, error) {
	if cardID == "" {
		return nil, &Error{Message: "vcard: cardID is required", Code: -1}
	}
	return n.signedRequest("GET", fmt.Sprintf("/api/v1/vcard/cards/%s/details", cardID), nil)
}

// GetCardCode returns the card's security code (CVV/CVC).
//
// GET /api/v1/vcard/cards/{id}/code  — HMAC 签名鉴权.
func (n *VCardNamespace) GetCardCode(cardID string) (json.RawMessage, error) {
	if cardID == "" {
		return nil, &Error{Message: "vcard: cardID is required", Code: -1}
	}
	return n.signedRequest("GET", fmt.Sprintf("/api/v1/vcard/cards/%s/code", cardID), nil)
}

// OpenCard opens a new virtual card.
//
// POST /api/v1/vcard/cards  — HMAC 签名鉴权.
//
//	params := map[string]any{
//	    "bin_platform_id": 1,
//	    "amount":          "20.00",
//	}
func (n *VCardNamespace) OpenCard(params map[string]any) (json.RawMessage, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, &Error{Message: "vcard: marshal body: " + err.Error(), Code: -1}
	}
	return n.signedRequest("POST", "/api/v1/vcard/cards", body)
}

// RechargeCard recharges an existing card.
//
// POST /api/v1/vcard/cards/{id}/recharge  — HMAC 签名鉴权.
//
//	params := map[string]any{"amount": "10.00"}
func (n *VCardNamespace) RechargeCard(cardID string, params map[string]any) (json.RawMessage, error) {
	if cardID == "" {
		return nil, &Error{Message: "vcard: cardID is required", Code: -1}
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, &Error{Message: "vcard: marshal body: " + err.Error(), Code: -1}
	}
	return n.signedRequest("POST", fmt.Sprintf("/api/v1/vcard/cards/%s/recharge", cardID), body)
}

// CancelCard cancels (closes) a card.
//
// POST /api/v1/vcard/cards/{id}/cancel  — HMAC 签名鉴权,无 body.
func (n *VCardNamespace) CancelCard(cardID string) (json.RawMessage, error) {
	if cardID == "" {
		return nil, &Error{Message: "vcard: cardID is required", Code: -1}
	}
	return n.signedRequest("POST", fmt.Sprintf("/api/v1/vcard/cards/%s/cancel", cardID), nil)
}

// VerifyWebhook 校验平台推送的 webhook 签名(给对接方验证平台主动推送的真实性).
//
// 复刻后端 pkg.GenerateSign 的算法:
//   - 取 params 中所有「非空、且 key != "sign"」的字段
//   - 按 key 升序拼成 "k1=v1&k2=v2&..."
//   - sign = hex_lower( HMAC_SHA256(secret, 拼接串) )
//   - 与 params["sign"] 用 hmac.Equal 做常量时间比较
//
// 防重放:平台推送一般同时带 sign_ts(签名时刻,unix 秒,±300s 窗口)与 nonce(随机串).
// 本函数只校验签名;时间窗(now-sign_ts 在 ±300s 内)与 nonce 去重需由调用方自行实现,
// 二者配合才能完整防重放.
//
// 用法:
//
//	params := map[string]string{
//	    "event":   "card.transaction",
//	    "card_id": "123",
//	    "sign_ts": "1718000000",
//	    "nonce":   "abc123...",
//	    "sign":    req.Header...或 body 里的 sign 字段,
//	}
//	if !nexcore.VerifyWebhook(params, webhookSecret) {
//	    // 验签失败,拒绝处理
//	}
func VerifyWebhook(params map[string]string, secret string) bool {
	got := params["sign"]
	if got == "" {
		return false
	}

	// 过滤空值与 sign 字段,按 key 升序
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" || k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params[k])
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(b.String()))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(got))
}
