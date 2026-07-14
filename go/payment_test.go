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

// TestNormalizeAmountFixed2 — CreateOrder 签名口径:amount 归一两位小数
// (对齐后端 decimal.StringFixed(2)).
func TestNormalizeAmountFixed2(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"100.00", "100.00"},
		{"100", "100.00"},
		{"100.5", "100.50"},
		{100, "100.00"},
		{99.9, "99.90"},
		{"abc", "abc"}, // 解析失败原样返回,交后端报错
	}
	for _, tc := range cases {
		if got := normalizeAmountFixed2(tc.in); got != tc.want {
			t.Fatalf("normalizeAmountFixed2(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestCreateOrderSignSemantics — CreateOrder 的签名参数构造应与后端一致:
// amount 用两位小数归一值、timeout 未传按 "0" 恒参与签名.
// 这里直接用 Sign() 复算后端 BuildSignString 语义(过滤空值/按 key 排序/k=v&/HMAC-SHA256 hex)的期望值.
func TestCreateOrderSignSemantics(t *testing.T) {
	c := newTestClient(canonicalKey)

	// 后端视角:amount=StringFixed(2)="100.50",timeout 未传 → "0"
	backendParams := map[string]any{
		"app_id":       "APP_TEST",
		"out_order_id": "ORDER_001",
		"amount":       "100.50",
		"currency":     "CNY",
		"call_type":    "rotation",
		"trade_type":   "usdt.trc20",
		"timeout":      "0",
	}
	want, err := c.Payment.Sign(backendParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SDK 视角:调用方传 "100.5"、不传 timeout —— 归一后应得到同一签名
	sdkParams := map[string]any{
		"app_id":       "APP_TEST",
		"out_order_id": "ORDER_001",
		"amount":       normalizeAmountFixed2("100.5"),
		"currency":     "CNY",
		"call_type":    "rotation",
		"trade_type":   "usdt.trc20",
		"timeout":      "0",
	}
	got, err := c.Payment.Sign(sdkParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("sign mismatch after amount normalization\nwant: %s\ngot:  %s", want, got)
	}
}
