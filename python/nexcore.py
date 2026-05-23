"""
NexCore Official Python SDK.

全能客户端,一次配置覆盖 Payment / Energy / SMTP / AI 全部业务。

用法:
    from nexcore import Client, NexCoreError

    client = Client(
        base_url="https://your-domain.com",
        payment_app_id="APP20260412XXXX",
        payment_app_key="your_app_key_here",
        energy_api_key="energy_api_key_here",
        energy_secret_key="energy_secret_key_here",
        ai_api_key="sk-nc-xxx",
        timeout=30,
    )

    # 链收款
    order = client.payment.create_order(
        out_order_id="ORDER_001",
        amount="100.00",
        currency="CNY",
        trade_type="usdt.trc20",
        call_type="rotation",
    )

    # 能量租赁
    est = client.energy.estimate_energy("TXxxxxxxxxxxxxxxxxxxxxx")

    # AI chat
    reply = client.ai.chat(
        messages=[{"role": "user", "content": "Hello"}],
        model="claude-opus-4-7",
    )
"""

from __future__ import annotations
import hmac
import hashlib
import json
import time
from typing import Any, Dict, List, Optional
from urllib.parse import urlencode

try:
    import requests
except ImportError as e:
    raise ImportError("nexcore SDK requires `requests`. Install: pip install requests") from e


class NexCoreError(Exception):
    """SDK 统一异常."""

    def __init__(self, message: str, code: int = -1, request_id: Optional[str] = None,
                 http_status: Optional[int] = None):
        super().__init__(message)
        self.code = code
        self.request_id = request_id
        self.http_status = http_status

    def __repr__(self) -> str:
        return f"NexCoreError(code={self.code}, message={self.args[0]!r}, request_id={self.request_id!r})"


class _PaymentNamespace:
    """链收款 — HMAC-SHA256 签名."""

    def __init__(self, client: "Client"):
        self._c = client

    def _sign(self, params: Dict[str, Any]) -> str:
        key = self._c._cfg.get("payment_app_key")
        if not key:
            raise NexCoreError("payment_app_key not configured")
        items = sorted(
            (k, v) for k, v in params.items()
            if v not in (None, "") and k != "sign"
        )
        msg = "&".join(f"{k}={v}" for k, v in items)
        return hmac.new(key.encode(), msg.encode(), hashlib.sha256).hexdigest()

    def _signed(self, params: Dict[str, Any]) -> Dict[str, Any]:
        app_id = self._c._cfg.get("payment_app_id")
        if not app_id:
            raise NexCoreError("payment_app_id not configured")
        p = dict(params, app_id=app_id)
        p["sign"] = self._sign(p)
        return p

    def create_order(self, **params) -> Dict[str, Any]:
        return self._c._request("POST", "/api/v1/pay/create", body=self._signed(params))

    def query_order(self, out_order_id: str) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/pay/query", query=self._signed({"out_order_id": out_order_id}))

    def close_order(self, out_order_id: str) -> Dict[str, Any]:
        return self._c._request("POST", "/api/v1/pay/close", body=self._signed({"out_order_id": out_order_id}))

    def bind_address(self, user_id: str, trade_type: str) -> Dict[str, Any]:
        return self._c._request("POST", "/api/v1/pay/bind-address",
                                body=self._signed({"user_id": user_id, "trade_type": trade_type}))

    def get_address(self, user_id: str, trade_type: str) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/pay/get-address",
                                query=self._signed({"user_id": user_id, "trade_type": trade_type}))

    def unbind_address(self, user_id: str) -> Dict[str, Any]:
        return self._c._request("POST", "/api/v1/pay/unbind-address",
                                body=self._signed({"user_id": user_id}))

    def app_config(self) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/pay/app-config", query=self._signed({}))

    def verify_notify_sign(self, payload: Dict[str, Any]) -> bool:
        """校验 webhook 回调签名."""
        sign = payload.get("sign")
        if not sign:
            return False
        expected = self._sign(payload)
        return hmac.compare_digest(expected, sign)


class _EnergyNamespace:
    """能量租赁 — X-API-Key + X-Secret-Key 双 header."""

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        k = self._c._cfg.get("energy_api_key")
        s = self._c._cfg.get("energy_secret_key")
        if not k or not s:
            raise NexCoreError("energy_api_key / energy_secret_key not configured")
        return {"X-API-Key": k, "X-Secret-Key": s}

    def info(self) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/energy/info", headers=self._headers())

    def price(self, energy: int, period: str = "1D") -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/energy/price",
                                query={"energy": energy, "period": period},
                                headers=self._headers())

    def estimate_energy(self, receive_addr: str) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/energy/estimate-energy",
                                query={"receive_addr": receive_addr},
                                headers=self._headers())

    def create_order(self, **params) -> Dict[str, Any]:
        return self._c._request("POST", "/api/v1/energy/order", body=params, headers=self._headers())

    def query_order(self, order_id: int) -> Dict[str, Any]:
        return self._c._request("GET", f"/api/v1/energy/order/{order_id}", headers=self._headers())

    def list_orders(self, **filter) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/energy/orders", query=filter, headers=self._headers())


class _SmtpNamespace:
    """SMTP 聚合 API."""

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        k = self._c._cfg.get("smtp_api_key")
        if not k:
            raise NexCoreError("smtp_api_key not configured")
        return {"X-API-Key": k}

    def send_mail(self, **params) -> Dict[str, Any]:
        return self._c._request("POST", "/api/v1/smtp/send", body=params, headers=self._headers())

    def list_accounts(self) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/smtp/accounts", headers=self._headers())

    def list_templates(self) -> Dict[str, Any]:
        return self._c._request("GET", "/api/v1/smtp/templates", headers=self._headers())


class _AiNamespace:
    """Astrenix AI(OpenAI 兼容协议)."""

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        k = self._c._cfg.get("ai_api_key")
        if not k:
            raise NexCoreError("ai_api_key not configured")
        return {"Authorization": f"Bearer {k}"}

    def chat(self, messages: List[Dict[str, str]], model: str, **extra) -> Dict[str, Any]:
        body = {"model": model, "messages": messages, **extra}
        return self._c._request("POST", "/v1/chat/completions", body=body, headers=self._headers())

    def models(self) -> Dict[str, Any]:
        return self._c._request("GET", "/v1/models", headers=self._headers())


class Client:
    """NexCore 全能客户端."""

    def __init__(self, base_url: str, **config):
        self.base_url = base_url.rstrip("/")
        self._cfg = config
        self._timeout = config.get("timeout", 30)
        self._verify_ssl = config.get("verify_ssl", True)

        self.payment = _PaymentNamespace(self)
        self.energy = _EnergyNamespace(self)
        self.smtp = _SmtpNamespace(self)
        self.ai = _AiNamespace(self)

    def _request(self, method: str, path: str,
                 body: Optional[Dict[str, Any]] = None,
                 query: Optional[Dict[str, Any]] = None,
                 headers: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
        url = self.base_url + path
        h = {"Accept": "application/json"}
        if body is not None:
            h["Content-Type"] = "application/json"
        if headers:
            h.update(headers)

        kwargs = {"headers": h, "timeout": self._timeout, "verify": self._verify_ssl}
        if query:
            kwargs["params"] = query
        if body is not None:
            kwargs["data"] = json.dumps(body, ensure_ascii=False).encode("utf-8")

        try:
            resp = requests.request(method.upper(), url, **kwargs)
        except requests.RequestException as e:
            raise NexCoreError(f"HTTP request failed: {e}") from e

        request_id = resp.headers.get("X-Trace-Id")
        try:
            data = resp.json()
        except ValueError:
            raise NexCoreError(
                f"HTTP {resp.status_code}: {resp.text[:200]}",
                http_status=resp.status_code, request_id=request_id
            )

        if resp.status_code >= 400 or not isinstance(data, dict):
            raise NexCoreError(
                data.get("message", f"HTTP {resp.status_code}") if isinstance(data, dict) else str(data),
                code=data.get("code", -1) if isinstance(data, dict) else -1,
                http_status=resp.status_code, request_id=request_id,
            )
        if data.get("code") not in (0, None):
            raise NexCoreError(
                data.get("message", "unknown error"),
                code=data.get("code", -1),
                http_status=resp.status_code, request_id=request_id,
            )
        # 业务响应若有 "data" 字段优先解包
        return data.get("data", data)
