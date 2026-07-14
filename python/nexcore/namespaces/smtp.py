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
    """实现 6 个 v1 endpoint(对照 internal/handler/smtp_api.go + smtp_api_ext.go):

    =====================================  =================  ====================================
    Endpoint                               方法               作用
    =====================================  =================  ====================================
    POST /api/v1/smtp/send                 send               发送单封邮件(定时/幂等)
    POST /api/v1/smtp/send/batch           send_batch         批量发送(recipients 逐人变量渲染)
    POST /api/v1/smtp/send/template        send_template      按模板 code 渲染发送
    GET  /api/v1/smtp/quota                get_quota          查询日/月配额与用量
    GET  /api/v1/smtp/status/:message_id   get_status         查询邮件投递状态
    POST /api/v1/smtp/inbound              report_inbound     上报退信/投诉(自动加入抑制名单)
    =====================================  =================  ====================================
    """

    def __init__(self, client: "Client"):
        self._c = client

    def _headers(self, idempotency_key: str = None) -> Dict[str, str]:
        k = self._c.get("smtp_api_key")
        if not k:
            raise NexCoreError("smtp_api_key not configured")
        h = {"Authorization": f"Bearer {k}"}
        if idempotency_key:
            h["Idempotency-Key"] = idempotency_key
        return h

    def send(
        self,
        to: str,
        subject: str,
        body: str,
        *,
        is_html: bool = False,
        from_name: str = None,
        reply_to: str = None,
        text_body: str = None,
        headers: Dict[str, str] = None,
        cc: List[str] = None,
        bcc: List[str] = None,
        attachments: List[Dict[str, str]] = None,
        account_id: int = None,
        send_at: str = None,
        idempotency_key: str = None,
    ) -> Dict[str, Any]:
        """发送单封邮件.

        ``POST /api/v1/smtp/send``

        Args:
            to: 收件人邮箱.
            subject: 邮件主题.
            body: 正文(纯文本或 HTML).
            is_html: body 是否为 HTML,默认 False.
            from_name: 发件人显示名(可选).
            reply_to: 回信地址(Reply-To 头,可选).
            text_body: 纯文本正文;与 HTML 同时提供时以 multipart/alternative 发送(可选).
            headers: 自定义邮件头键值对,核心头不可覆盖(可选).
            cc: 抄送列表(可选).
            bcc: 密送列表,只投递不写头(可选).
            attachments: 附件列表,元素 ``{filename, content_base64, content_type}``(可选).
            account_id: 指定发信账户 ID,默认自动选号;指定后不故障转移(可选).
            send_at: 定时发送时间(RFC3339);晚于当前 30 秒以上则排期(可选).
            idempotency_key: 幂等键(``Idempotency-Key`` 头);同 key 重试直接
                返回首次结果,不重复发送、不重复扣配额(可选).

        Returns:
            立即发送:dict 含 ``message_id`` / ``status`` / ``account_name`` /
            ``used_smtp`` / ``account_id`` / ``send_duration_ms``;
            定时分支:``{scheduled: True, scheduled_id, send_at}``.
        """
        payload: Dict[str, Any] = {"to": to, "subject": subject, "body": body, "is_html": is_html}
        optional = {
            "from_name": from_name, "reply_to": reply_to, "text_body": text_body,
            "headers": headers, "cc": cc, "bcc": bcc, "attachments": attachments,
            "account_id": account_id, "send_at": send_at,
        }
        payload.update({k: v for k, v in optional.items() if v is not None})
        return self._c.http.request(
            "POST", "/api/v1/smtp/send",
            body=payload, headers=self._headers(idempotency_key),
        )

    def send_batch(
        self,
        recipients: List[Dict[str, Any]],
        *,
        subject: str = None,
        body: str = None,
        template_code: str = None,
        is_html: bool = False,
        reply_to: str = None,
        cc: List[str] = None,
        bcc: List[str] = None,
        attachments: List[Dict[str, str]] = None,
        headers: Dict[str, str] = None,
        account_id: int = None,
        idempotency_key: str = None,
    ) -> Dict[str, Any]:
        """批量发送(逐收件人独立变量渲染;逐封独立扣配额).

        ``POST /api/v1/smtp/send/batch``

        静态模式传 ``subject`` + ``body``(支持 ``{{var}}`` 占位),
        模板模式传 ``template_code``;二者至少其一.
        单次收件人上限 = 订阅的 max_batch_size(默认 10);模板模式需套餐支持模板功能.

        Args:
            recipients: 收件人列表(必填),元素 ``{to(必填), variables?, from_name?}``.
            subject: 静态模式主题(可选).
            body: 静态模式正文,支持 ``{{var}}`` 替换(与 template_code 二选一).
            template_code: 模板模式:模板 code(与 body 二选一).
            is_html: 正文是否 HTML.
            reply_to: Reply-To 头,每封相同(可选).
            cc: 抄送,每封重复(可选).
            bcc: 密送,每封重复(可选).
            attachments: 附件,每封重复(可选).
            headers: 自定义邮件头(可选).
            account_id: 指定发信账户(可选).
            idempotency_key: 幂等键(``Idempotency-Key`` 头,可选).

        Returns:
            dict 含 ``total`` / ``success`` / ``failed`` /
            ``results``(元素 ``{to, status, message_id?, error?}``).
        """
        payload: Dict[str, Any] = {"recipients": recipients, "is_html": is_html}
        optional = {
            "subject": subject, "body": body, "template_code": template_code,
            "reply_to": reply_to, "cc": cc, "bcc": bcc,
            "attachments": attachments, "headers": headers, "account_id": account_id,
        }
        payload.update({k: v for k, v in optional.items() if v is not None})
        return self._c.http.request(
            "POST", "/api/v1/smtp/send/batch",
            body=payload, headers=self._headers(idempotency_key),
        )

    def send_template(
        self,
        to: str,
        template_code: str,
        variables: Dict[str, str] = None,
        *,
        from_name: str = None,
    ) -> Dict[str, Any]:
        """按模板 code 渲染发送单封邮件.模板需要先在用户后台 "SMTP API → 模板管理" 创建.

        ``POST /api/v1/smtp/send/template``

        Args:
            to: 收件邮箱(必填).
            template_code: 模板 code(必填).
            variables: 渲染变量,对应模板中 ``{{var_name}}`` 占位符(可选).
            from_name: 发件人显示名(可选).

        Returns:
            dict 含 ``message_id`` / ``status`` / ``used_smtp``.
        """
        payload: Dict[str, Any] = {"to": to, "template_code": template_code}
        if variables is not None:
            payload["variables"] = variables
        if from_name is not None:
            payload["from_name"] = from_name
        return self._c.http.request("POST", "/api/v1/smtp/send/template", body=payload, headers=self._headers())

    def get_quota(self) -> Dict[str, Any]:
        """查询当前订阅期内的配额与已用量.

        ``GET /api/v1/smtp/quota``

        Returns:
            dict 含 ``daily_limit`` / ``daily_used`` / ``daily_remaining`` /
            ``monthly_limit`` / ``monthly_used`` / ``monthly_remaining`` / ``expire_at``.
        """
        return self._c.http.request("GET", "/api/v1/smtp/quota", headers=self._headers())

    def get_status(self, message_id: str) -> Dict[str, Any]:
        """查询指定邮件的投递状态.

        ``GET /api/v1/smtp/status/:message_id``

        Args:
            message_id: send / send_batch / send_template 返回的 message_id.

        Returns:
            dict 含 ``message_id`` / ``status``(pending=待处理,sending=发送中,
            success=成功,failed=失败)/ ``from_email`` / ``to_email`` / ``subject`` /
            ``is_html`` / ``account_id`` / ``account_name`` / ``error_message`` /
            ``smtp_response`` / ``send_duration_ms`` / ``opened_at`` / ``open_count`` /
            ``clicked_at`` / ``click_count`` / ``created_at`` 等.
        """
        return self._c.http.request("GET", f"/api/v1/smtp/status/{message_id}", headers=self._headers())

    def report_inbound(
        self,
        *,
        email: str = None,
        message_id: str = None,
        type: str = None,
    ) -> Dict[str, Any]:
        """上报退信/投诉事件(自动把邮箱加入抑制名单并标记对应 send_log).

        ``POST /api/v1/smtp/inbound``

        Args:
            email: 退信/投诉的邮箱(与 message_id 至少提供其一).
            message_id: 关联邮件的 message_id(与 email 至少提供其一).
            type: ``bounce`` | ``complaint``(可选).

        Returns:
            ``{ok: True}``.
        """
        payload = {k: v for k, v in {"email": email, "message_id": message_id, "type": type}.items() if v is not None}
        return self._c.http.request("POST", "/api/v1/smtp/inbound", body=payload, headers=self._headers())
