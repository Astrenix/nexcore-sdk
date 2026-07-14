"""Tovanix (formerly NexCore) Official Python SDK — public package interface.

一次 ``from nexcore import Client`` 即可拿到全部业务能力
(Payment / Withdraw / Exchange / Energy / SMTP).

公开 API:
    - :class:`Client` — 主客户端
    - :class:`NexCoreError` — 统一异常

内部模块(也可显式 import 使用,但通常无需):
    - :mod:`nexcore.client` — Client 实现
    - :mod:`nexcore.errors` — 异常定义
    - :mod:`nexcore.http` — 底层 HTTP 传输
    - :mod:`nexcore.namespaces.payment` — 多链收款(HMAC-SHA256)
    - :mod:`nexcore.namespaces.withdraw` — 多链收款 · 提币(RSA-2048)
    - :mod:`nexcore.namespaces.exchange` — 汇率
    - :mod:`nexcore.namespaces.energy` — TRON 能量租赁
    - :mod:`nexcore.namespaces.smtp` — SMTP 聚合 API
    - :mod:`nexcore.namespaces.account` — 账户余额 / 充值地址
    - :mod:`nexcore.namespaces.vcard` — 虚拟信用卡(含 verify_webhook 回调验签)
"""

from .client import Client
from .errors import NexCoreError
from .namespaces.vcard import verify_webhook as verify_vcard_webhook

__version__ = "3.3.0"
__all__ = ["Client", "NexCoreError", "verify_vcard_webhook", "__version__"]
