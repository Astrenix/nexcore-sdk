"""账户命名空间.

对应 /docs 文档 "账户 API" 模块的 v1 公开接口(``/api/v1/account/*``).
鉴权:X-API-Key + X-Secret-Key 双 header(与 energy 同一套 MPK 商户密钥).

account 与 vcard 共用 ``api_key`` / ``api_secret`` 这一对 MPK 商户密钥
(在 :class:`~nexcore.client.Client` 的 ``api_key`` / ``api_secret`` 字段配置).
"""

from __future__ import annotations
from typing import TYPE_CHECKING, Any, Dict

from ..errors import NexCoreError

if TYPE_CHECKING:
    from ..client import Client


class Account:
    """实现 2 个 v1 endpoint(对照 internal/handler/account_api.go):

    =====================================  ====================  ======================================
    Endpoint                               方法                  作用
    =====================================  ====================  ======================================
    GET  /api/v1/account/balance           get_balance           查询账户余额
    GET  /api/v1/account/deposit-address   get_deposit_address   查询充值地址
    =====================================  ====================  ======================================

    鉴权:X-API-Key + X-Secret-Key 双 header(只读接口走双密钥即可,无需签名).
    """

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        k = self._c.get("api_key")
        s = self._c.get("api_secret")
        if not k or not s:
            raise NexCoreError("api_key / api_secret not configured")
        return {"X-API-Key": k, "X-Secret-Key": s}

    def get_balance(self) -> Dict[str, Any]:
        """查询账户余额.

        ``GET /api/v1/account/balance``

        Returns:
            dict 含 ``balance``(USD 余额)/ ``currency``(恒 USD)/
            ``deposit_address``(TRON 固定充值地址对象,未绑定时为 None).
        """
        return self._c.http.request("GET", "/api/v1/account/balance", headers=self._headers())

    def get_deposit_address(self) -> Dict[str, Any]:
        """查询充值地址.

        ``GET /api/v1/account/deposit-address``

        Returns:
            dict 含链 / 币种 / 充值地址等字段.
        """
        return self._c.http.request("GET", "/api/v1/account/deposit-address", headers=self._headers())
