"""NexCore SDK 统一异常."""

from __future__ import annotations
from typing import Optional


class NexCoreError(Exception):
    """SDK 全局异常.

    所有调用失败(网络错误 / HTTP 4xx-5xx / 业务 code != 0)统一抛本异常.

    Attributes:
        code: 平台错误码(0=成功;-1=客户端层错误;其他参见错误码表).
        request_id: 服务端日志追踪 ID,通过响应头 ``X-Trace-Id`` 透传.
            排查问题时给后端工单提供本值.
        http_status: 实际 HTTP 状态码;客户端层错误时为 None.
    """

    def __init__(
        self,
        message: str,
        code: int = -1,
        request_id: Optional[str] = None,
        http_status: Optional[int] = None,
    ):
        super().__init__(message)
        self.code = code
        self.request_id = request_id
        self.http_status = http_status

    def __repr__(self) -> str:
        return (
            f"NexCoreError(code={self.code}, message={self.args[0]!r}, "
            f"request_id={self.request_id!r}, http_status={self.http_status!r})"
        )
