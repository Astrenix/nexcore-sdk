"""提币 API namespace — 多链收款业务的资金出库端.

鉴权:RSA-PKCS1v15-SHA256 签名 + 4 个请求头

    X-API-Key            账户级 API Key(控制台「账号 → API 密钥」)
    X-Timestamp          unix ms,与服务器时差 ≤ 60s
    X-Nonce              一次性 nonce(uuid v4),5 分钟内不可重复
    X-Withdraw-Signature RSA-PKCS1v15-SHA256(caller_private_key, signString),Base64

signString = METHOD + "\\n" + PATH + "\\n" + TIMESTAMP + "\\n" + NONCE + "\\n" + BODY
其中 BODY 为 HTTP body 原文(JSON 字符串原样,GET 请求为空字符串).

对应 /docs 文档 "提币 API" 章节的 4 个 endpoint:

    POST /api/v1/withdraw                 create_withdraw            发起提币
    GET  /api/v1/withdraw/:id             get_withdraw               查询单笔状态
    GET  /api/v1/balance/withdrawable     get_withdrawable_balance   查询可提余额
    GET  /api/v1/fee/quote                quote_fee                  费用预估

另提供 verify_callback() 校验平台回调签名(用平台公钥).
"""

from __future__ import annotations

import base64
import json
import time
import uuid
from typing import Any, Dict, Optional, TYPE_CHECKING

from ..errors import NexCoreError

if TYPE_CHECKING:  # 仅类型检查时引入,运行时 lazy load
    from cryptography.hazmat.primitives.asymmetric import rsa


def _lazy_crypto():
    """运行时按需 import cryptography — 用户不用提币就不需要装这个依赖."""
    try:
        from cryptography.hazmat.primitives import hashes, serialization
        from cryptography.hazmat.primitives.asymmetric import padding, rsa as _rsa
        from cryptography.exceptions import InvalidSignature
        return hashes, serialization, padding, _rsa, InvalidSignature
    except ImportError as e:
        raise NexCoreError(
            "提币 API 需要 `cryptography` 库,请运行:pip install cryptography"
        ) from e


class Withdraw:
    """提币 API namespace(RSA-2048 签名).

    用法::

        order = client.withdraw.create_withdraw({
            "chain": "tron",
            "symbol": "USDT",
            "amount": "100.5",
            "to_address": "TXxxxxxxxxxxxxxxxxxxxxxxxxxx",
            "request_id": "your-idempotency-uuid",
        })
    """

    def __init__(self, client):
        self._c = client
        self._priv_key = None  # cached after first call
        self._platform_pub = None

    # ---------- 内部:密钥懒解析 ----------

    def _get_priv_key(self):
        if self._priv_key is not None:
            return self._priv_key
        pem = self._c.get("withdraw_private_key_pem")
        if not pem:
            raise NexCoreError("withdraw_private_key_pem not configured")
        _, serialization, _, _rsa, _ = _lazy_crypto()
        try:
            key = serialization.load_pem_private_key(
                pem.encode("utf-8") if isinstance(pem, str) else pem,
                password=None,
            )
        except Exception as e:
            raise NexCoreError(f"withdraw: invalid private key PEM: {e}") from e
        if not isinstance(key, _rsa.RSAPrivateKey):
            raise NexCoreError("withdraw: configured private key is not RSA")
        self._priv_key = key
        return key

    def _get_platform_pub(self):
        if self._platform_pub is not None:
            return self._platform_pub
        pem = self._c.get("withdraw_platform_public_key_pem")
        if not pem:
            raise NexCoreError("withdraw_platform_public_key_pem not configured")
        _, serialization, _, _rsa, _ = _lazy_crypto()
        try:
            key = serialization.load_pem_public_key(
                pem.encode("utf-8") if isinstance(pem, str) else pem,
            )
        except Exception as e:
            raise NexCoreError(f"withdraw: invalid platform public key PEM: {e}") from e
        if not isinstance(key, _rsa.RSAPublicKey):
            raise NexCoreError("withdraw: platform key is not RSA")
        self._platform_pub = key
        return key

    # ---------- 签名 ----------

    def sign(self, method: str, path: str, timestamp: str, nonce: str, body: str) -> str:
        """计算请求的 RSA-PKCS1v15-SHA256 签名(Base64).

        业务方一般不需要直接调,SDK 内部 _do 时自动调用.
        公开出来便于测试 / 自实现非标场景(比如 curl 调试).
        """
        sign_string = f"{method.upper()}\n{path}\n{timestamp}\n{nonce}\n{body}"
        priv = self._get_priv_key()
        hashes, _, padding, _, _ = _lazy_crypto()
        sig = priv.sign(
            sign_string.encode("utf-8"),
            padding.PKCS1v15(),
            hashes.SHA256(),
        )
        return base64.b64encode(sig).decode("ascii")

    def _do(
        self,
        method: str,
        path: str,
        body: Optional[Dict[str, Any]] = None,
        query: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """统一发请求,自动加 4 个鉴权头."""
        api_key = self._c.get("withdraw_api_key")
        if not api_key:
            raise NexCoreError("withdraw_api_key not configured")
        timestamp = str(int(time.time() * 1000))
        nonce = str(uuid.uuid4())
        body_bytes: Optional[bytes] = None
        body_str = ""
        if body is not None:
            body_bytes = json.dumps(body, ensure_ascii=False).encode("utf-8")
            body_str = body_bytes.decode("utf-8")
        sig = self.sign(method, path, timestamp, nonce, body_str)
        return self._c.http.request(
            method,
            path,
            body_raw=body_bytes,
            query=query,
            headers={
                "X-API-Key": api_key,
                "X-Timestamp": timestamp,
                "X-Nonce": nonce,
                "X-Withdraw-Signature": sig,
            },
        )

    # ---------- 公开 endpoint ----------

    def create_withdraw(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """发起提币 — POST /api/v1/withdraw.

        下单后状态为 ``pending``,等待延迟到期由 worker 自动广播.
        期间可在控制台暂停 / 加速 / 取消.

        Args:
            params: 提币参数,必填 chain / symbol / amount / to_address;
                可选 memo / callback_url / request_id.

        Returns:
            dict: ``{order_id, status, amount, fee, fee_mode, delayed_until}``
        """
        return self._do("POST", "/api/v1/withdraw", body=params)

    def get_withdraw(self, order_id: str) -> Dict[str, Any]:
        """查询单笔提币状态 — GET /api/v1/withdraw/:id."""
        if not order_id:
            raise NexCoreError("order_id is required")
        return self._do("GET", f"/api/v1/withdraw/{order_id}")

    def get_withdrawable_balance(self) -> Dict[str, Any]:
        """查询可提余额 — GET /api/v1/balance/withdrawable.

        返回该账户在每条链 × 每种资产下的「已归集待提现」余额.
        只有这部分可用于 API 提币.
        """
        return self._do("GET", "/api/v1/balance/withdrawable")

    def quote_fee(self, chain: str, symbol: str, amount: Optional[str] = None) -> Dict[str, Any]:
        """费用预估 — GET /api/v1/fee/quote.

        Args:
            chain: tron / eth / bsc / polygon / arbitrum / btc
            symbol: USDT / TRX / ETH 等
            amount: 提币金额(字符串,可选)

        Returns:
            dict: ``{chain, symbol, amount, fee_amount, fee_asset}``
        """
        if not chain or not symbol:
            raise NexCoreError("chain and symbol are required")
        q: Dict[str, Any] = {"chain": chain, "symbol": symbol}
        if amount:
            q["amount"] = amount
        return self._do("GET", "/api/v1/fee/quote", query=q)

    # ---------- 回调验签 ----------

    def verify_callback(
        self,
        method: str,
        path: str,
        timestamp: str,
        nonce: str,
        body: bytes,
        base64_signature: str,
    ) -> None:
        """验证平台回调签名(对接方收到 webhook 时调用).

        用法::

            sig = request.headers["X-Platform-Signature"]
            ts  = request.headers["X-Timestamp"]
            nonce = request.headers["X-Nonce"]
            body = request.body  # 原始 bytes,不要 re-serialize
            try:
                client.withdraw.verify_callback(
                    request.method, request.path, ts, nonce, body, sig,
                )
            except NexCoreError:
                return 401

        验签算法与请求方向一致:RSA-PKCS1v15-SHA256(platform_public_key, signString).
        """
        pub = self._get_platform_pub()
        hashes, _, padding, _, InvalidSignature = _lazy_crypto()
        body_str = body.decode("utf-8") if isinstance(body, (bytes, bytearray)) else str(body)
        sign_string = f"{method.upper()}\n{path}\n{timestamp}\n{nonce}\n{body_str}"
        try:
            sig = base64.b64decode(base64_signature)
        except Exception as e:
            raise NexCoreError(f"withdraw: bad signature base64: {e}") from e
        try:
            pub.verify(sig, sign_string.encode("utf-8"), padding.PKCS1v15(), hashes.SHA256())
        except InvalidSignature:
            raise NexCoreError("withdraw: signature verify failed")

