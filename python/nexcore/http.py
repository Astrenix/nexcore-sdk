"""底层 HTTP 传输层.

各业务命名空间(payment / exchange / energy / smtp)调用本类发请求,
不直接接触 requests,保持 namespace 关注业务逻辑.
"""

from __future__ import annotations
import json
from typing import Any, Dict, Optional

try:
    import requests
except ImportError as e:
    raise ImportError(
        "nexcore SDK 需要 `requests`,请运行:pip install requests"
    ) from e

from .errors import NexCoreError


class Http:
    """封装 requests 调用,处理 envelope 解包 + 异常统一."""

    def __init__(
        self,
        base_url: str,
        timeout: int = 30,
        verify_ssl: bool = True,
        user_agent: str = "NexCore-Python-SDK/3.0.0",
    ):
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.verify_ssl = verify_ssl
        self.user_agent = user_agent

    def request(
        self,
        method: str,
        path: str,
        body: Optional[Dict[str, Any]] = None,
        query: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> Dict[str, Any]:
        """发送 HTTP 请求.

        Args:
            method: HTTP 方法 GET / POST / PUT / DELETE.
            path: 以 / 开头的路径,如 ``/api/v1/pay/create``.
            body: JSON 请求体(自动序列化).
            query: query 参数.
            headers: 额外 header.

        Returns:
            响应中的 ``data`` 段(已解 ``{code, message, data}`` envelope).

        Raises:
            NexCoreError: 网络错误 / HTTP 错误 / 业务 code != 0.
        """
        url = self.base_url + path
        h: Dict[str, str] = {
            "Accept": "application/json",
            "User-Agent": self.user_agent,
        }
        if body is not None:
            h["Content-Type"] = "application/json"
        if headers:
            h.update(headers)

        kwargs: Dict[str, Any] = {
            "headers": h,
            "timeout": self.timeout,
            "verify": self.verify_ssl,
        }
        if query:
            # 过滤空值
            kwargs["params"] = {k: v for k, v in query.items() if v not in (None, "")}
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
                http_status=resp.status_code,
                request_id=request_id,
            )

        if resp.status_code >= 400 or not isinstance(data, dict):
            msg = data.get("message", f"HTTP {resp.status_code}") if isinstance(data, dict) else str(data)
            code = data.get("code", -1) if isinstance(data, dict) else -1
            raise NexCoreError(msg, code=code, http_status=resp.status_code, request_id=request_id)

        if data.get("code") not in (0, None):
            raise NexCoreError(
                data.get("message", "unknown error"),
                code=data.get("code", -1),
                http_status=resp.status_code,
                request_id=request_id,
            )

        return data.get("data", data)
