"""SMTP 聚合 API 命名空间.

对应 /docs 文档 "SMTP API" 模块的全部 v1 公开接口.
鉴权:Bearer Token — ``Authorization: Bearer smk_xxx``.
"""

from __future__ import annotations
from typing import TYPE_CHECKING, Any, Dict, List

from ..errors import NexCoreError

if TYPE_CHECKING:
    from ..client import Client


class Smtp:
    """实现 5 个 v1 endpoint(对照 internal/handler/smtp_api.go + smtp_api_ext.go):

    =====================================  =================  ====================================
    Endpoint                               方法               作用
    =====================================  =================  ====================================
    POST /api/v1/smtp/send                 send               发送单封邮件
    POST /api/v1/smtp/send/batch           send_batch         批量发送(同主题/正文,多收件人)
    POST /api/v1/smtp/send/template        send_template      按模板渲染发送
    GET  /api/v1/smtp/quota                get_quota          查询本期配额与用量
    GET  /api/v1/smtp/status/:message_id   get_status         查询邮件投递状态
    =====================================  =================  ====================================
    """

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self) -> Dict[str, str]:
        k = self._c.get("smtp_api_key")
        if not k:
            raise NexCoreError("smtp_api_key not configured")
        return {"Authorization": f"Bearer {k}"}

    def send(
        self,
        to: str,
        subject: str,
        body: str,
        *,
        is_html: bool = False,
        account_id: int = None,
        reply_to: str = None,
    ) -> Dict[str, Any]:
        """发送单封邮件.

        ``POST /api/v1/smtp/send``

        Args:
            to: 收件人邮箱.
            subject: 邮件主题.
            body: 正文(纯文本或 HTML).
            is_html: body 是否为 HTML,默认 False.
            account_id: 指定发信账户 ID(可选,默认自动选最优).
            reply_to: 回信地址(可选).

        Returns:
            dict 含 ``message_id`` / ``status``.
        """
        payload: Dict[str, Any] = {"to": to, "subject": subject, "body": body, "is_html": is_html}
        if account_id is not None:
            payload["account_id"] = account_id
        if reply_to is not None:
            payload["reply_to"] = reply_to
        return self._c.http.request("POST", "/api/v1/smtp/send", body=payload, headers=self._headers())

    def send_batch(
        self,
        to: List[str],
        subject: str,
        body: str,
        *,
        is_html: bool = False,
        account_id: int = None,
    ) -> Dict[str, Any]:
        """批量发送(同主题/正文,多收件人).

        ``POST /api/v1/smtp/send/batch``

        Args:
            to: 收件人邮箱列表.
            subject: 统一主题.
            body: 统一正文.
            is_html: 默认 False.
            account_id: 指定发信账户(可选).

        Returns:
            dict 含 ``message_ids`` / ``total`` / ``accepted``.
        """
        payload: Dict[str, Any] = {"to": to, "subject": subject, "body": body, "is_html": is_html}
        if account_id is not None:
            payload["account_id"] = account_id
        return self._c.http.request("POST", "/api/v1/smtp/send/batch", body=payload, headers=self._headers())

    def send_template(
        self,
        to: str,
        template_id: int,
        variables: Dict[str, Any],
        *,
        account_id: int = None,
    ) -> Dict[str, Any]:
        """按模板渲染发送.模板需要先在用户后台 "SMTP API → 模板管理" 创建.

        ``POST /api/v1/smtp/send/template``

        Args:
            to: 收件人.
            template_id: 模板 ID.
            variables: 渲染变量,对应模板中 ``{{var_name}}`` 占位符.
            account_id: 指定发信账户(可选).
        """
        payload: Dict[str, Any] = {"to": to, "template_id": template_id, "variables": variables}
        if account_id is not None:
            payload["account_id"] = account_id
        return self._c.http.request("POST", "/api/v1/smtp/send/template", body=payload, headers=self._headers())

    def get_quota(self) -> Dict[str, Any]:
        """查询当前订阅期内的配额与已用量.

        ``GET /api/v1/smtp/quota``

        Returns:
            dict 含 ``today_used`` / ``today_quota`` / ``period_used`` / ``period_quota`` / ``expires_at``.
        """
        return self._c.http.request("GET", "/api/v1/smtp/quota", headers=self._headers())

    def get_status(self, message_id: str) -> Dict[str, Any]:
        """查询指定邮件的投递状态.

        ``GET /api/v1/smtp/status/:message_id``

        Args:
            message_id: send / send_batch / send_template 返回的 message_id.

        Returns:
            dict 含 ``status`` / ``sent_at`` / ``opened_at`` / ``clicked_at`` / ``error_msg`` 等.
        """
        return self._c.http.request("GET", f"/api/v1/smtp/status/{message_id}", headers=self._headers())
