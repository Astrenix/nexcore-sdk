"""虚拟信用卡命名空间.

对应 /docs 文档 "虚拟卡 API" 模块的 v1 公开接口(``/api/v1/vcard/*``).

两套鉴权(对照 internal/handler/virtual_card_api.go 的路由分组):

    1. **双密钥读**(X-API-Key + X-Secret-Key)— 只读 / 低敏接口
       与 energy / account 同一套 MPK 商户密钥.
    2. **HMAC 头签名**(X-Key-ID + X-Timestamp + X-Nonce + X-Signature)
       — 涉及资金或敏感数据(开卡 / 充值 / 注销 / 卡密 / CVV)的接口,
       后端挂 ``RequireAPISignature()`` 强制签名.

account 与 vcard 共用 ``api_key`` / ``api_secret`` 这一对 MPK 商户密钥
(在 :class:`~nexcore.client.Client` 的 ``api_key`` / ``api_secret`` 字段配置).

签名算法(已与后端 ``internal/middleware/api_auth.go`` 字节级对齐)::

    ts        = str(int(time.time()))          # 秒级 unix
    nonce     = secrets.token_hex(8)           # 一次性,5min 不可重复
    payload   = ts + nonce + METHOD + path + raw_query + body
    signature = hex(hmac_sha256(api_secret, payload))

    headers:
        X-Key-ID    = api_key
        X-Timestamp = ts
        X-Nonce     = nonce
        X-Signature = signature

其中 ``path`` 含 ``:id`` 但不含 query;``raw_query`` 这些签名接口均为空串("");
``body`` 为实际发出的 JSON 字符串(GET 为 "")。

⚠️ POST 关键:必须**先 ``json.dumps`` 成字符串、对该串签名、再用 http 的 ``body_raw``
发同一串字节**,避免二次序列化导致签名串与实际 body 字节不一致(http.py 的
``body_raw`` 优先正是为此).
"""

from __future__ import annotations
import hashlib
import hmac
import json
import secrets
import time
from typing import TYPE_CHECKING, Any, Dict, Optional

from ..errors import NexCoreError

if TYPE_CHECKING:
    from ..client import Client


def verify_webhook(params: Dict[str, Any], secret: str) -> bool:
    """校验虚拟卡 webhook 回调签名(常量时间比较).

    复刻后端 ``pkg.BuildSignString`` + ``GenerateSign``:
    取所有 **非空、非 ``sign``** 字段,按 key 升序拼成 ``k1=v1&k2=v2``,
    用 ``secret`` 做 HMAC-SHA256(hex),与 ``params['sign']`` 用
    :func:`hmac.compare_digest` 常量时间比较.

    回调体内还带 ``sign_ts`` / ``nonce`` 字段,接收方应自行校验
    ``sign_ts`` 在合理时间窗内、``nonce`` 未重复使用以防重放攻击.

    Args:
        params: 回调 JSON 完整解码后的 dict.
        secret: 用户后台「虚拟卡 webhook」配置里的 secret.

    Returns:
        True=签名正确,可信;False=签名错误/缺失,应拒绝该回调.
    """
    sign = params.get("sign")
    if not sign:
        return False
    items = sorted(
        (k, v) for k, v in params.items()
        if k != "sign" and v not in (None, "")
    )
    msg = "&".join(f"{k}={v}" for k, v in items)
    expected = hmac.new(secret.encode(), msg.encode(), hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, str(sign))


class VCard:
    """实现 12 个 v1 endpoint(对照 internal/handler/virtual_card_api.go):

    双密钥读(X-API-Key + X-Secret-Key):

    =========================================  ======================  =====================
    Endpoint                                   方法                    作用
    =========================================  ======================  =====================
    GET  /api/v1/vcard/info                    get_info                平台 / 账户卡信息
    GET  /api/v1/vcard/bins                    list_bins               可开卡 BIN 列表
    GET  /api/v1/vcard/cards                   list_cards              我的卡列表
    GET  /api/v1/vcard/cards/:id/transactions  get_card_transactions   卡交易流水
    GET  /api/v1/vcard/orders                  list_orders             订单列表
    GET  /api/v1/vcard/orders/:id              get_order               订单详情
    PUT  /api/v1/vcard/cards/:id/remark        update_card_remark      修改卡备注
    =========================================  ======================  =====================

    HMAC 头签名(X-Key-ID + X-Timestamp + X-Nonce + X-Signature):

    =========================================  ======================  =====================
    Endpoint                                   方法                    作用
    =========================================  ======================  =====================
    GET  /api/v1/vcard/cards/:id/details       get_card_details        卡敏感详情(卡号等)
    GET  /api/v1/vcard/cards/:id/code          get_card_code           卡 CVV / 动态码
    POST /api/v1/vcard/cards                    open_card               开卡
    POST /api/v1/vcard/cards/:id/recharge      recharge_card           卡充值
    POST /api/v1/vcard/cards/:id/cancel        cancel_card             注销卡(无 body)
    =========================================  ======================  =====================

    另导出模块级 :func:`verify_webhook` 用于校验 webhook 回调签名.
    """

    def __init__(self, client: "Client"):
        self._c = client

    # ---------- 双密钥(只读)----------

    def _headers(self) -> Dict[str, str]:
        k = self._c.get("api_key")
        s = self._c.get("api_secret")
        if not k or not s:
            raise NexCoreError("api_key / api_secret not configured")
        return {"X-API-Key": k, "X-Secret-Key": s}

    # ---------- HMAC 头签名 ----------

    def _signed_request(
        self,
        method: str,
        path: str,
        body_dict: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """统一处理 HMAC 头签名请求.

        POST/PUT:把 ``body_dict`` 先 ``json.dumps`` 成字符串,对该串签名,
        再用 ``body_raw`` 发出同一串字节(确保签名串 == 实际 body 字节).
        GET:无 body,签名 body 段为空串.

        Args:
            method: HTTP 方法.
            path: 请求路径(含 :id,不含 query).
            body_dict: 请求体 dict;None 表示无 body(GET / 无参 POST).
        """
        api_key = self._c.get("api_key")
        api_secret = self._c.get("api_secret")
        if not api_key or not api_secret:
            raise NexCoreError("api_key / api_secret not configured")

        ts = str(int(time.time()))
        nonce = secrets.token_hex(8)
        method_up = method.upper()
        raw_query = ""  # 这些签名接口均无 query 参数

        body_bytes: Optional[bytes] = None
        body_str = ""
        if body_dict is not None:
            body_bytes = json.dumps(body_dict, ensure_ascii=False).encode("utf-8")
            body_str = body_bytes.decode("utf-8")

        payload = ts + nonce + method_up + path + raw_query + body_str
        sig = hmac.new(api_secret.encode(), payload.encode(), hashlib.sha256).hexdigest()

        return self._c.http.request(
            method_up,
            path,
            body_raw=body_bytes,
            headers={
                "X-Key-ID": api_key,
                "X-Timestamp": ts,
                "X-Nonce": nonce,
                "X-Signature": sig,
            },
        )

    # ============ 双密钥读 endpoints ============

    def get_info(self) -> Dict[str, Any]:
        """平台 / 账户卡信息.

        ``GET /api/v1/vcard/info``
        """
        return self._c.http.request("GET", "/api/v1/vcard/info", headers=self._headers())

    def list_bins(self) -> Dict[str, Any]:
        """可开卡 BIN 列表.

        ``GET /api/v1/vcard/bins``
        """
        return self._c.http.request("GET", "/api/v1/vcard/bins", headers=self._headers())

    def list_cards(self) -> Dict[str, Any]:
        """我的卡列表.

        ``GET /api/v1/vcard/cards``
        """
        return self._c.http.request("GET", "/api/v1/vcard/cards", headers=self._headers())

    def get_card_transactions(self, card_id: str) -> Dict[str, Any]:
        """卡交易流水.

        ``GET /api/v1/vcard/cards/:id/transactions``

        Args:
            card_id: 卡 ID.
        """
        if not card_id:
            raise NexCoreError("card_id is required")
        return self._c.http.request(
            "GET", f"/api/v1/vcard/cards/{card_id}/transactions", headers=self._headers()
        )

    def list_orders(self, **params: Any) -> Dict[str, Any]:
        """订单列表(可分页 / 过滤).

        ``GET /api/v1/vcard/orders``

        Args:
            **params: page / page_size / status / order_type 等.
        """
        return self._c.http.request(
            "GET", "/api/v1/vcard/orders", query=params, headers=self._headers()
        )

    def get_order(self, order_id: str) -> Dict[str, Any]:
        """订单详情.

        ``GET /api/v1/vcard/orders/:id``

        Args:
            order_id: 订单 ID.
        """
        if not order_id:
            raise NexCoreError("order_id is required")
        return self._c.http.request(
            "GET", f"/api/v1/vcard/orders/{order_id}", headers=self._headers()
        )

    def update_card_remark(self, card_id: str, remark: str) -> Dict[str, Any]:
        """修改卡备注.

        ``PUT /api/v1/vcard/cards/:id/remark``

        走双密钥鉴权(非签名接口).

        Args:
            card_id: 卡 ID.
            remark: 新备注.
        """
        if not card_id:
            raise NexCoreError("card_id is required")
        return self._c.http.request(
            "PUT", f"/api/v1/vcard/cards/{card_id}/remark",
            body={"remark": remark}, headers=self._headers(),
        )

    # ============ HMAC 头签名 endpoints ============

    def get_card_details(self, card_id: str) -> Dict[str, Any]:
        """卡敏感详情(完整卡号 / 有效期等).

        ``GET /api/v1/vcard/cards/:id/details`` — 签名鉴权.

        Args:
            card_id: 卡 ID.
        """
        if not card_id:
            raise NexCoreError("card_id is required")
        return self._signed_request("GET", f"/api/v1/vcard/cards/{card_id}/details")

    def get_card_code(self, card_id: str) -> Dict[str, Any]:
        """卡 CVV / 动态安全码.

        ``GET /api/v1/vcard/cards/:id/code`` — 签名鉴权.

        Args:
            card_id: 卡 ID.
        """
        if not card_id:
            raise NexCoreError("card_id is required")
        return self._signed_request("GET", f"/api/v1/vcard/cards/{card_id}/code")

    def open_card(self, **params: Any) -> Dict[str, Any]:
        """开卡.

        ``POST /api/v1/vcard/cards`` — 签名鉴权.

        Args:
            bin_platform_id (int): 卡段 platform_id(list_bins 返回,必填).
            amount (float): 开卡充值金额(必填,>0).

        Returns:
            dict 含 ``order_id`` / ``status`` / ``total_cost``.
        """
        return self._signed_request("POST", "/api/v1/vcard/cards", body_dict=dict(params))

    def recharge_card(self, card_id: str, **params: Any) -> Dict[str, Any]:
        """卡充值.

        ``POST /api/v1/vcard/cards/:id/recharge`` — 签名鉴权.

        Args:
            card_id: 卡 ID.
            **params: amount 等.
        """
        if not card_id:
            raise NexCoreError("card_id is required")
        return self._signed_request(
            "POST", f"/api/v1/vcard/cards/{card_id}/recharge", body_dict=dict(params)
        )

    def cancel_card(self, card_id: str) -> Dict[str, Any]:
        """注销卡(无 body).

        ``POST /api/v1/vcard/cards/:id/cancel`` — 签名鉴权.

        Args:
            card_id: 卡 ID.
        """
        if not card_id:
            raise NexCoreError("card_id is required")
        return self._signed_request("POST", f"/api/v1/vcard/cards/{card_id}/cancel")
