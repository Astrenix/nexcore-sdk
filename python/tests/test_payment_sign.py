"""NexCore Python SDK — Payment 签名算法测试.

跨语言一致性 fixture(PHP / Python / Node / Go 4 个 SDK 跑出同样结果):
    key      = "test-key-abc-123"
    params   = { app_id, amount, currency, out_order_id, trade_type }
    expected = 44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72

不依赖真实后端 / 不发 HTTP 请求,纯算法验证.

运行::

    cd python && python -m unittest discover tests
    # 或
    cd python && pytest tests/
"""
import os
import sys
import unittest

# 让 tests/ 能 import 仓库根目录的 nexcore/
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from nexcore import Client


CANONICAL_KEY = "test-key-abc-123"
CANONICAL_PARAMS = {
    "app_id": "APP_TEST",
    "amount": "100.00",
    "currency": "CNY",
    "out_order_id": "ORDER_001",
    "trade_type": "usdt.trc20",
}
CANONICAL_SIGN = "44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72"


def _new_client(key: str = CANONICAL_KEY) -> Client:
    return Client(
        base_url="https://example.com",
        payment_app_id="APP_TEST",
        payment_app_key=key,
    )


class PaymentSignTest(unittest.TestCase):
    """Payment.sign() 跨语言一致性 + 各边界情况."""

    def test_canonical_fixture(self):
        """跟 PHP / Node / Go 跑出同样的 hex 签名."""
        client = _new_client()
        self.assertEqual(client.payment.sign(CANONICAL_PARAMS), CANONICAL_SIGN)

    def test_empty_values_filtered(self):
        """空字符串和 None 字段必须被过滤,不影响签名."""
        client = _new_client()
        params = dict(CANONICAL_PARAMS, empty_field="", null_field=None)
        self.assertEqual(client.payment.sign(params), CANONICAL_SIGN)

    def test_sign_field_excluded(self):
        """sign 字段自身必须被过滤(防自指)."""
        client = _new_client()
        params = dict(CANONICAL_PARAMS, sign="tampered-value")
        self.assertEqual(client.payment.sign(params), CANONICAL_SIGN)

    def test_wrong_key_produces_different_sign(self):
        client_wrong = _new_client(key="wrong-key")
        self.assertNotEqual(client_wrong.payment.sign(CANONICAL_PARAMS), CANONICAL_SIGN)


class VerifyNotifySignTest(unittest.TestCase):
    """Webhook 回调签名校验."""

    def setUp(self):
        self.client = _new_client()

    def test_valid_sign_accepted(self):
        payload = dict(CANONICAL_PARAMS, sign=CANONICAL_SIGN)
        self.assertTrue(self.client.payment.verify_notify_sign(payload))

    def test_tampered_sign_rejected(self):
        payload = dict(CANONICAL_PARAMS, sign="0" * 64)
        self.assertFalse(self.client.payment.verify_notify_sign(payload))

    def test_missing_sign_rejected(self):
        self.assertFalse(self.client.payment.verify_notify_sign(CANONICAL_PARAMS))


if __name__ == "__main__":
    unittest.main(verbosity=2)
