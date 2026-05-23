package nexcore

import (
	"testing"
)

// Payment 签名算法测试.
//
// 跨语言一致性 fixture(PHP / Python / Node / Go 4 个 SDK 跑出同样结果):
//
//	key      = "test-key-abc-123"
//	params   = { app_id, amount, currency, out_order_id, trade_type }
//	expected = 44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72
//
// 不依赖真实后端 / 不发 HTTP 请求,纯算法验证.
//
// 运行: cd go && go test ./...

const (
	canonicalKey  = "test-key-abc-123"
	canonicalSign = "44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72"
)

func canonicalParams() map[string]any {
	return map[string]any{
		"app_id":       "APP_TEST",
		"amount":       "100.00",
		"currency":     "CNY",
		"out_order_id": "ORDER_001",
		"trade_type":   "usdt.trc20",
	}
}

func newTestClient(key string) *Client {
	return NewClient(Config{
		BaseURL:       "https://example.com",
		PaymentAppID:  "APP_TEST",
		PaymentAppKey: key,
	})
}

// TestPaymentSign_CanonicalFixture 验证签名算法跟 PHP/Python/Node 一致.
func TestPaymentSign_CanonicalFixture(t *testing.T) {
	c := newTestClient(canonicalKey)
	sig, err := c.Payment.Sign(canonicalParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig != canonicalSign {
		t.Fatalf("sign mismatch\nexpected: %s\nactual:   %s", canonicalSign, sig)
	}
}

// TestPaymentSign_EmptyValuesFiltered — 空字符串字段被过滤,签名跟基础一致.
func TestPaymentSign_EmptyValuesFiltered(t *testing.T) {
	c := newTestClient(canonicalKey)
	params := canonicalParams()
	params["empty_field"] = ""
	params["nil_field"] = nil

	sig, err := c.Payment.Sign(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig != canonicalSign {
		t.Fatalf("expected empty/nil filtered, got %s", sig)
	}
}

// TestPaymentSign_SignFieldExcluded — sign 字段自身不参与签名.
func TestPaymentSign_SignFieldExcluded(t *testing.T) {
	c := newTestClient(canonicalKey)
	params := canonicalParams()
	params["sign"] = "tampered-value"

	sig, err := c.Payment.Sign(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig != canonicalSign {
		t.Fatalf("expected sign field excluded, got %s", sig)
	}
}

// TestPaymentSign_WrongKeyDifferent — 不同 key 产生不同签名.
func TestPaymentSign_WrongKeyDifferent(t *testing.T) {
	c := newTestClient("wrong-key")
	sig, err := c.Payment.Sign(canonicalParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig == canonicalSign {
		t.Fatal("expected different sign for different key")
	}
	if len(sig) != 64 {
		t.Fatalf("expected 64-char hex sign, got %d chars", len(sig))
	}
}

// TestVerifyNotifySign_Valid — 正确签名通过.
func TestVerifyNotifySign_Valid(t *testing.T) {
	c := newTestClient(canonicalKey)
	payload := canonicalParams()
	payload["sign"] = canonicalSign

	if !c.Payment.VerifyNotifySign(payload) {
		t.Fatal("expected valid sign to be accepted")
	}
}

// TestVerifyNotifySign_Tampered — 伪造签名拒绝.
func TestVerifyNotifySign_Tampered(t *testing.T) {
	c := newTestClient(canonicalKey)
	payload := canonicalParams()
	payload["sign"] = "0000000000000000000000000000000000000000000000000000000000000000"

	if c.Payment.VerifyNotifySign(payload) {
		t.Fatal("expected tampered sign to be rejected")
	}
}

// TestVerifyNotifySign_Missing — 缺 sign 字段拒绝.
func TestVerifyNotifySign_Missing(t *testing.T) {
	c := newTestClient(canonicalKey)
	if c.Payment.VerifyNotifySign(canonicalParams()) {
		t.Fatal("expected missing sign to be rejected")
	}
}
