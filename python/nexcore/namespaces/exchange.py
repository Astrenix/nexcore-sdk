"""汇率服务命名空间.

对应 /docs 文档 "多链收款 → 汇率服务" 5 个 endpoint.
走 APIAuth 中间件,用 X-App-Key + X-App-Secret(应用密钥)进行 header 鉴权.
"""

from __future__ import annotations
from typing import TYPE_CHECKING, Any, Dict, List

from ..errors import NexCoreError

if TYPE_CHECKING:
    from ..client import Client


class Exchange:
    """实现 5 个 v1 endpoint(对照 internal/handler/exchange_api.go):

    ============================  ==================  ===========================
    Endpoint                      方法                作用
    ============================  ==================  ===========================
    GET  /api/v1/rate             get_rate            单对币种汇率
    POST /api/v1/convert          convert             金额换算
    GET  /api/v1/rates            get_rates           批量获取多币种汇率
    GET  /api/v1/rates/fiat       get_fiat_rates      主流法币汇率
    GET  /api/v1/rates/all        get_all_rates       所有支持币种快照
    ============================  ==================  ===========================
    """

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        app_id = self._c.get("payment_app_id")
        app_key = self._c.get("payment_app_key")
        if not app_id or not app_key:
            raise NexCoreError("payment_app_id / payment_app_key not configured")
        return {"X-App-Key": app_id, "X-App-Secret": app_key}

    def get_rate(self, from_: str, to: str) -> Dict[str, Any]:
        """查询单对币种汇率.

        ``GET /api/v1/rate?from=USDT&to=CNY``

        Args:
            from_: 来源币种代码(USDT / TRX / ETH / BTC / USD / CNY ...).
                参数名 ``from_`` 是为了避开 Python 关键字.
            to: 目标币种代码.

        Returns:
            dict 含 ``from`` / ``to`` / ``rate`` / ``inverse`` / ``updated_at``.
        """
        return self._c.http.request(
            "GET", "/api/v1/rate",
            query={"from": from_, "to": to},
            headers=self._headers(),
        )

    def convert(self, from_: str, to: str, amount: Any) -> Dict[str, Any]:
        """金额换算.

        ``POST /api/v1/convert``

        Args:
            from_: 来源币种代码.
            to: 目标币种代码.
            amount: 待换算金额(string / float / int).

        Returns:
            dict 含 ``from`` / ``to`` / ``amount`` / ``result`` / ``rate`` /
            ``updated_at``(``result`` 为换算结果).
        """
        return self._c.http.request(
            "POST", "/api/v1/convert",
            body={"from": from_, "to": to, "amount": amount},
            headers=self._headers(),
        )

    def get_rates(self, symbols: List[str], base: str = None) -> Dict[str, Any]:
        """批量获取多币种到指定基准币的汇率.

        ``GET /api/v1/rates?symbols=USDT,TRX,ETH&base=USDT``

        Args:
            symbols: 待查询的币种代码列表.
            base: 基准币;不传由后端取默认(USDT).
        """
        query: Dict[str, Any] = {"symbols": ",".join(symbols)}
        if base:
            query["base"] = base
        return self._c.http.request(
            "GET", "/api/v1/rates",
            query=query,
            headers=self._headers(),
        )

    def get_fiat_rates(self, base: str = "USD") -> Dict[str, Any]:
        """主流法币到指定基准法币的汇率.

        ``GET /api/v1/rates/fiat?base=USD``
        """
        return self._c.http.request(
            "GET", "/api/v1/rates/fiat",
            query={"base": base},
            headers=self._headers(),
        )

    def get_all_rates(self, base: str = "USDT") -> Dict[str, Any]:
        """所有支持币种的汇率快照(加密币 + 法币).

        ``GET /api/v1/rates/all?base=USDT``
        """
        return self._c.http.request(
            "GET", "/api/v1/rates/all",
            query={"base": base},
            headers=self._headers(),
        )
