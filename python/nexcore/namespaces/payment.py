"""多链收款命名空间.

对应 /docs 文档 "多链收款" 模块的全部 v1 公开接口.

鉴权:HMAC-SHA256 签名 — 所有请求自动追加 ``app_id`` + ``sign`` 字段.
签名算法:把所有参数按 key 升序拼接成 ``k1=v1&k2=v2``,然后用 ``app_key`` 做 HMAC-SHA256.
"""

from __future__ import annotations
import hmac
import hashlib
from typing import TYPE_CHECKING, Any, Dict

from ..errors import NexCoreError

if TYPE_CHECKING:
    from ..client import Client


class Payment:
    """实现 7 个 v1 endpoint(对照 internal/handler/order.go + one_to_one.go):

    ===========================  ==================  ===========================
    Endpoint                     方法                作用
    ===========================  ==================  ===========================
    POST /api/v1/pay/create      create_order        创建收款订单
    GET  /api/v1/pay/query       query_order         查询订单状态
    POST /api/v1/pay/close       close_order         关闭订单
    GET  /api/v1/pay/app-config  get_app_config      查询应用配置
    POST /api/v1/pay/bind-       bind_address        一对一 — 绑定地址
         address
    POST /api/v1/pay/get-        get_user_address    一对一 — 查询用户已绑地址
         address
    POST /api/v1/pay/unbind-     unbind_address      一对一 — 解绑
         address
    ===========================  ==================  ===========================

    另提供 :meth:`verify_notify_sign` 校验 webhook 回调签名(常量时间比较).
    """

    def __init__(self, client: "Client"):
        self._c = client

    def sign(self, params: Dict[str, Any]) -> str:
        """计算 HMAC-SHA256 签名.

        业务方一般不需要直接调,SDK 内部自动调用.公开出来便于:
            - 自行测试签名是否正确(对照 /docs 文档输出)
            - 校验回调签名

        Args:
            params: 待签名参数(会自动过滤 ``sign`` 字段和空值,按 key 升序排).

        Returns:
            64 字符小写 hex 签名.

        Raises:
            NexCoreError: ``payment_app_key`` 未配置.
        """
        key = self._c.get("payment_app_key")
        if not key:
            raise NexCoreError("payment_app_key not configured")
        items = sorted(
            (k, v) for k, v in params.items()
            if v not in (None, "") and k != "sign"
        )
        msg = "&".join(f"{k}={v}" for k, v in items)
        return hmac.new(key.encode(), msg.encode(), hashlib.sha256).hexdigest()

    def _signed(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """自动注入 app_id + 计算 sign."""
        app_id = self._c.get("payment_app_id")
        if not app_id:
            raise NexCoreError("payment_app_id not configured")
        p = dict(params, app_id=app_id)
        p["sign"] = self.sign(p)
        return p

    # ============ Endpoints ============

    def create_order(self, **params: Any) -> Dict[str, Any]:
        """创建收款订单.

        ``POST /api/v1/pay/create``

        Args:
            out_order_id (str): 商户侧订单号(必须唯一).
            amount (str|float): 法币金额,推荐两位小数 string 避免浮点误差.
            currency (str): 法币代码 CNY/USD/EUR/JPY/KRW/HKD.
            trade_type (str): 加密币种.链,如 ``"usdt.trc20"``.
            call_type (str, optional): ``"rotation"``(轮播)或 ``"one_to_one"``(一对一),默认 rotation.
            user_id (str, optional): 一对一模式必填.
            timeout (int, optional): 订单过期秒数,默认 1800.
            subject (str, optional): 订单描述.
            notify_url (str, optional): webhook 回调 URL.
            return_url (str, optional): 支付成功后跳转 URL.

        Returns:
            dict 含 ``order_id`` / ``pay_address`` / ``crypto_amount`` / ``crypto_currency`` / ``expires_at`` 等字段.
        """
        return self._c.http.request("POST", "/api/v1/pay/create", body=self._signed(params))

    def query_order(self, out_order_id: str) -> Dict[str, Any]:
        """查询订单当前状态.

        ``GET /api/v1/pay/query``
        """
        return self._c.http.request("GET", "/api/v1/pay/query", query=self._signed({"out_order_id": out_order_id}))

    def close_order(self, out_order_id: str) -> Dict[str, Any]:
        """主动关闭订单.

        ``POST /api/v1/pay/close``
        """
        return self._c.http.request("POST", "/api/v1/pay/close", body=self._signed({"out_order_id": out_order_id}))

    def get_app_config(self) -> Dict[str, Any]:
        """查询当前应用配置(启用的币种 / 支付模式 / 回调 URL 等).

        ``GET /api/v1/pay/app-config``
        """
        return self._c.http.request("GET", "/api/v1/pay/app-config", query=self._signed({}))

    def bind_address(self, user_id: str, trade_type: str) -> Dict[str, Any]:
        """一对一模式 — 给用户绑定一个固定收款地址.

        ``POST /api/v1/pay/bind-address``
        """
        return self._c.http.request(
            "POST", "/api/v1/pay/bind-address",
            body=self._signed({"user_id": user_id, "trade_type": trade_type}),
        )

    def get_user_address(self, user_id: str, trade_type: str) -> Dict[str, Any]:
        """一对一模式 — 查询用户已绑定的地址.

        ``POST /api/v1/pay/get-address``(注意:后端是 POST,不是 GET)
        """
        return self._c.http.request(
            "POST", "/api/v1/pay/get-address",
            body=self._signed({"user_id": user_id, "trade_type": trade_type}),
        )

    def unbind_address(self, user_id: str) -> Dict[str, Any]:
        """一对一模式 — 解绑用户地址.

        ``POST /api/v1/pay/unbind-address``
        """
        return self._c.http.request(
            "POST", "/api/v1/pay/unbind-address",
            body=self._signed({"user_id": user_id}),
        )

    def verify_notify_sign(self, payload: Dict[str, Any]) -> bool:
        """校验 webhook 回调签名.

        NexCore 平台通过 ``notify_url`` 推送 JSON 通知时会带 ``sign`` 字段.
        本方法用 :func:`hmac.compare_digest` 常量时间比较防止时序攻击.

        Args:
            payload: 回调 JSON 完整解码后的 dict.

        Returns:
            True=签名正确,可信;False=签名错误/缺失,应拒绝该回调.
        """
        sign = payload.get("sign")
        if not sign:
            return False
        expected = self.sign(payload)
        return hmac.compare_digest(expected, sign)
