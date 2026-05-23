"""NexCore Official Python SDK — public package interface.

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
"""

from .client import Client
from .errors import NexCoreError

__version__ = "3.1.0"
__all__ = ["Client", "NexCoreError", "__version__"]
