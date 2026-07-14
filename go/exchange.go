package nexcore

import (
	"encoding/json"
	"strings"
)

// ExchangeNamespace implements all 5 exchange-related v1 endpoints.
//
// 对应 /docs 文档 "多链收款 → 汇率服务" 5 个 endpoint
// (对照 internal/handler/exchange_api.go):
//
//	GET  /api/v1/rate          GetRate         单对币种汇率
//	POST /api/v1/convert       Convert         金额换算
//	GET  /api/v1/rates         GetRates        批量获取多币种汇率
//	GET  /api/v1/rates/fiat    GetFiatRates    主流法币汇率
//	GET  /api/v1/rates/all     GetAllRates     所有支持币种快照
//
// 鉴权:走 APIAuth 中间件,用 X-App-Key + X-App-Secret(应用密钥)header.
type ExchangeNamespace struct {
	c *Client
}

// authHeaders 构造 APIAuth header — 汇率接口用应用密钥(PaymentAppID/Key).
func (n *ExchangeNamespace) authHeaders() (map[string]string, error) {
	if n.c.cfg.PaymentAppID == "" || n.c.cfg.PaymentAppKey == "" {
		return nil, &Error{Message: "PaymentAppID / PaymentAppKey not configured", Code: -1}
	}
	return map[string]string{
		"X-App-Key":    n.c.cfg.PaymentAppID,
		"X-App-Secret": n.c.cfg.PaymentAppKey,
	}, nil
}

// GetRate queries a single rate pair.
//
// GET /api/v1/rate?from=USDT&to=CNY
//
// 返回 {from, to, rate, inverse, updated_at},inverse = 1/rate.
func (n *ExchangeNamespace) GetRate(from, to string) (json.RawMessage, error) {
	h, err := n.authHeaders()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/rate", &requestOpts{
		Query:   map[string]any{"from": from, "to": to},
		Headers: h,
	})
}

// Convert performs an amount conversion between two currencies.
//
// POST /api/v1/convert
//
// 参数 amount 可以是 string 或 number,推荐 string 避免浮点误差.
// 返回 {from, to, amount, result, rate, updated_at}.
func (n *ExchangeNamespace) Convert(from, to string, amount any) (json.RawMessage, error) {
	h, err := n.authHeaders()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/convert", &requestOpts{
		Body:    map[string]any{"from": from, "to": to, "amount": amount},
		Headers: h,
	})
}

// GetRates batch-queries multiple symbols against a base currency.
//
// GET /api/v1/rates?symbols=USDT,TRX,ETH&base=CNY
//
// base 传空字符串时不携带该参数,由后端取默认值 USDT.
//
// 返回 {base, rates: {USDT: 7.23, TRX: 0.85, ...}, updated_at}.
func (n *ExchangeNamespace) GetRates(symbols []string, base string) (json.RawMessage, error) {
	h, err := n.authHeaders()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/rates", &requestOpts{
		Query:   map[string]any{"symbols": strings.Join(symbols, ","), "base": base},
		Headers: h,
	})
}

// GetFiatRates returns rates of major fiat currencies against a base fiat.
//
// GET /api/v1/rates/fiat?base=USD
func (n *ExchangeNamespace) GetFiatRates(base string) (json.RawMessage, error) {
	if base == "" {
		base = "USD"
	}
	h, err := n.authHeaders()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/rates/fiat", &requestOpts{
		Query:   map[string]any{"base": base},
		Headers: h,
	})
}

// GetAllRates returns a snapshot of all supported currencies (crypto + fiat).
//
// GET /api/v1/rates/all?base=USDT
func (n *ExchangeNamespace) GetAllRates(base string) (json.RawMessage, error) {
	if base == "" {
		base = "USDT"
	}
	h, err := n.authHeaders()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/rates/all", &requestOpts{
		Query:   map[string]any{"base": base},
		Headers: h,
	})
}
