"""NexCore SDK 主客户端."""

from __future__ import annotations
from typing import Any, Dict, Optional

from .errors import NexCoreError
from .http import Http
from .namespaces.payment import Payment
from .namespaces.exchange import Exchange
from .namespaces.energy import Energy
from .namespaces.smtp import Smtp
from .namespaces.withdraw import Withdraw


class Client:
    """NexCore 全能 Python 客户端.

    一次配置覆盖 NexCore 平台全部 v1 公开接口,业务按 namespace 划分:

        - ``client.payment``  — 多链收款(HMAC-SHA256 签名)
        - ``client.exchange`` — 汇率(X-App-Key + X-App-Secret header)
        - ``client.energy``   — TRON 能量租赁(X-API-Key + X-Secret-Key)
        - ``client.smtp``     — SMTP 聚合(Bearer Token)

    用法::

        from nexcore import Client, NexCoreError

        client = Client(
            base_url="https://your-domain.com",
            payment_app_id="APP20260412XXXX",
            payment_app_key="your_app_key_here",
            energy_api_key="energy_key",
            energy_secret_key="energy_secret",
            smtp_api_key="smk_xxx",
        )

        order = client.payment.create_order(
            out_order_id="ORDER_001",
            amount="100.00",
            currency="CNY",
            trade_type="usdt.trc20",
            call_type="rotation",
        )

    所有错误统一抛 :class:`NexCoreError`(含 ``code`` / ``request_id`` / ``http_status``).
    """

    VERSION = "3.1.0"

    def __init__(
        self,
        base_url: str,
        *,
        payment_app_id: Optional[str] = None,
        payment_app_key: Optional[str] = None,
        energy_api_key: Optional[str] = None,
        energy_secret_key: Optional[str] = None,
        smtp_api_key: Optional[str] = None,
        withdraw_api_key: Optional[str] = None,
        withdraw_private_key_pem: Optional[str] = None,
        withdraw_platform_public_key_pem: Optional[str] = None,
        timeout: int = 30,
        verify_ssl: bool = True,
        user_agent: Optional[str] = None,
    ):
        """初始化客户端.

        Args:
            base_url: NexCore 平台基础 URL,如 ``https://your-domain.com``.
            payment_app_id: 多链收款应用 ID(Payment / Exchange 都需要).
            payment_app_key: 多链收款应用密钥.
            energy_api_key: 能量租赁 X-API-Key.
            energy_secret_key: 能量租赁 X-Secret-Key.
            smtp_api_key: SMTP 聚合 API 的 ``smk_`` 前缀 Token.
            timeout: HTTP 超时秒数,默认 30.
            verify_ssl: 是否验证 SSL 证书,默认 True.
            user_agent: 自定义 UA(可选).
        """
        if not base_url:
            raise NexCoreError("base_url is required")

        self._cfg: Dict[str, Any] = {
            "base_url": base_url,
            "payment_app_id": payment_app_id,
            "payment_app_key": payment_app_key,
            "energy_api_key": energy_api_key,
            "energy_secret_key": energy_secret_key,
            "smtp_api_key": smtp_api_key,
            "withdraw_api_key": withdraw_api_key,
            "withdraw_private_key_pem": withdraw_private_key_pem,
            "withdraw_platform_public_key_pem": withdraw_platform_public_key_pem,
        }

        self.http = Http(
            base_url=base_url,
            timeout=timeout,
            verify_ssl=verify_ssl,
            user_agent=user_agent or f"NexCore-Python-SDK/{self.VERSION}",
        )

        self.payment = Payment(self)
        self.exchange = Exchange(self)
        self.energy = Energy(self)
        self.smtp = Smtp(self)
        self.withdraw = Withdraw(self)

    def get(self, key: str) -> Optional[Any]:
        """取配置字段(各 namespace 内部使用).

        Args:
            key: 配置字段名,如 ``"payment_app_key"``.

        Returns:
            字段值或 None.
        """
        return self._cfg.get(key)
