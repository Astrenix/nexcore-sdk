"""TRON 能量租赁命名空间.

对应 /docs 文档 "能量租赁" 模块的全部 v1 公开接口.
鉴权:X-API-Key + X-Secret-Key 双 header.
"""

from __future__ import annotations
from typing import TYPE_CHECKING, Any, Dict

from ..errors import NexCoreError

if TYPE_CHECKING:
    from ..client import Client


class Energy:
    """实现 8 个 v1 endpoint(对照 internal/handler/trxx_api.go):

    =====================================  =====================  ======================================
    Endpoint                               方法                   作用
    =====================================  =====================  ======================================
    GET  /api/v1/energy/info               get_info               平台公开信息(可用能量/阶梯定价)
    GET  /api/v1/energy/price              get_price              指定能量 + 周期的报价
    GET  /api/v1/energy/estimate-energy    estimate_energy        根据地址估算 TRC20 转账所需能量
    POST /api/v1/energy/order              create_order           创建常规租赁订单
    POST /api/v1/energy/order/onetime      create_onetime_order   创建一次性订单
    GET  /api/v1/energy/order/:serial      query_order            查询订单(serial 字符串)
    GET  /api/v1/energy/orders             list_orders            列出订单
    POST /api/v1/energy/order/reclaim      reclaim_order          主动回收订单
    =====================================  =====================  ======================================
    """

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        k = self._c.get("energy_api_key")
        s = self._c.get("energy_secret_key")
        if not k or not s:
            raise NexCoreError("energy_api_key / energy_secret_key not configured")
        return {"X-API-Key": k, "X-Secret-Key": s}

    def get_info(self) -> Dict[str, Any]:
        """平台公开信息.

        ``GET /api/v1/energy/info``

        Returns:
            dict 含 ``platform_avail_energy`` / ``minimum_order_energy`` /
            ``maximum_order_energy`` / ``tiered_pricing`` 等.
        """
        return self._c.http.request("GET", "/api/v1/energy/info", headers=self._headers())

    def get_price(self, energy_amount: int, period: str = "1D") -> Dict[str, Any]:
        """获取指定能量数 + 周期的报价.

        ``GET /api/v1/energy/price?period=1D&energy_amount=65000``

        Args:
            energy_amount: 需要的能量值.
            period: 租期 1H / 1D / 3D / 7D / 30D,默认 1D.

        Returns:
            dict 含 ``period`` / ``energy_amount`` / ``price_trx``
            (price_trx 为终价,已含 API 加价).
        """
        return self._c.http.request(
            "GET", "/api/v1/energy/price",
            query={"period": period, "energy_amount": energy_amount},
            headers=self._headers(),
        )

    def estimate_energy(self, to_address: str) -> Dict[str, Any]:
        """根据目标地址估算 TRC20 转账所需能量.

        ``GET /api/v1/energy/estimate-energy?to_address=TXxxxxxxxx``

        Args:
            to_address: 目标 TRON 地址(T 开头,34 位).

        Returns:
            dict 含 ``to_address`` / ``initialized`` / ``suggested_energy``
            (initialized=False 表示地址未持有 USDT,首笔转账消耗更多能量).
        """
        return self._c.http.request(
            "GET", "/api/v1/energy/estimate-energy",
            query={"to_address": to_address},
            headers=self._headers(),
        )

    def create_order(self, **params: Any) -> Dict[str, Any]:
        """创建常规租赁订单.

        ``POST /api/v1/energy/order``

        Args:
            receive_address (str): 接收能量的 TRON 地址.
            energy_amount (int): 能量数(>= minimum_order_energy).
            period (str): 1H / 1D / 3D / 7D / 30D.
            out_trade_no (str, optional): 商户侧自定义订单号.
            remark (str, optional): 备注.

        Returns:
            dict 含 ``serial`` / ``price_trx`` / ``deducted_usd``.
        """
        return self._c.http.request("POST", "/api/v1/energy/order", body=params, headers=self._headers())

    def create_onetime_order(self, **params: Any) -> Dict[str, Any]:
        """单笔能量下单(笔数策略,系统按策略自动分配能量数).

        ``POST /api/v1/energy/order/onetime``

        Args:
            receive_address (str): 接收能量的 TRON 地址.
            period (str): 1H / 1D / 3D / 7D / 30D.
            out_trade_no (str, optional): 商户侧自定义订单号.
            remark (str, optional): 备注.

        Returns:
            dict 含 ``serial`` / ``price_trx`` / ``deducted_usd``
            (price_trx 按上游实际结算,多退少不补).
        """
        return self._c.http.request("POST", "/api/v1/energy/order/onetime", body=params, headers=self._headers())

    def query_order(self, serial: str) -> Dict[str, Any]:
        """查询订单状态(会先向上游同步一次最新状态).

        ``GET /api/v1/energy/order/:serial``

        Args:
            serial: 订单序列号(string,**不是**数字 id).

        Returns:
            dict 含 ``serial`` / ``receive_address`` / ``energy_amount`` / ``period`` /
            ``price_trx`` / ``status`` / ``status_msg`` / ``out_trade_no`` /
            ``order_type`` / ``created_at``;status:0=待处理/处理中,40=成功,41=失败.
        """
        return self._c.http.request("GET", f"/api/v1/energy/order/{serial}", headers=self._headers())

    def list_orders(self, **filter_: Any) -> Dict[str, Any]:
        """列出所有订单(可按状态过滤).

        ``GET /api/v1/energy/orders``

        Args:
            **filter_: page(默认 1)/ page_size(默认 20,上限 100)/
                status(-1=全部,0=待处理/处理中,40=成功,41=失败).

        Returns:
            dict 含 ``list`` / ``total`` / ``page`` / ``page_size``.
        """
        return self._c.http.request("GET", "/api/v1/energy/orders", query=filter_, headers=self._headers())

    def reclaim_order(self, serial: str) -> Dict[str, Any]:
        """主动回收订单.

        ``POST /api/v1/energy/order/reclaim``

        Args:
            serial: 订单序列号.

        Returns:
            dict 含 ``errno`` / ``message``(errno=0 回收成功).
        """
        return self._c.http.request("POST", "/api/v1/energy/order/reclaim", body={"serial": serial}, headers=self._headers())
