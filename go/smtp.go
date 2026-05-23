package nexcore

import (
	"encoding/json"
	"fmt"
)

// SMTPNamespace implements all 5 SMTP aggregation v1 endpoints.
//
// 对应 /docs 文档 "SMTP API" 模块的全部 v1 公开接口
// (对照 internal/handler/smtp_api.go + smtp_api_ext.go):
//
//	POST /api/v1/smtp/send                 Send          发送单封邮件
//	POST /api/v1/smtp/send/batch           SendBatch     批量发送(同主题/正文,多收件人)
//	POST /api/v1/smtp/send/template        SendTemplate  按模板渲染发送
//	GET  /api/v1/smtp/quota                GetQuota      查询本期配额与用量
//	GET  /api/v1/smtp/status/:message_id   GetStatus     查询邮件投递状态
//
// 鉴权:Bearer Token — "Authorization: Bearer smk_xxx".
type SMTPNamespace struct {
	c *Client
}

// headers 构造 Authorization Bearer header.
func (n *SMTPNamespace) headers() (map[string]string, error) {
	if n.c.cfg.SMTPAPIKey == "" {
		return nil, &Error{Message: "SMTPAPIKey not configured", Code: -1}
	}
	return map[string]string{"Authorization": "Bearer " + n.c.cfg.SMTPAPIKey}, nil
}

// Send sends a single email.
//
// POST /api/v1/smtp/send
//
// params 至少包含:
//
//	to       (string) — 收件人邮箱
//	subject  (string) — 邮件主题
//	body     (string) — 正文(纯文本或 HTML)
//
// 可选:
//
//	is_html    (bool) — body 是否为 HTML,默认 false
//	account_id (int)  — 指定发信账户 ID(默认自动选最优)
//	reply_to   (string) — 回信地址
//
// 返回 {message_id, status}.
func (n *SMTPNamespace) Send(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/smtp/send", &requestOpts{Body: params, Headers: h})
}

// SendBatch sends one mail to multiple recipients with same subject/body.
//
// POST /api/v1/smtp/send/batch
//
// params 至少包含:
//
//	to       ([]string) — 收件人邮箱列表
//	subject  (string)   — 统一主题
//	body     (string)   — 统一正文
//
// 返回 {message_ids, total, accepted}.
func (n *SMTPNamespace) SendBatch(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/smtp/send/batch", &requestOpts{Body: params, Headers: h})
}

// SendTemplate sends mail rendered from a saved template.
//
// POST /api/v1/smtp/send/template
//
// 模板需要先在用户后台 "SMTP API → 模板管理" 创建.
//
// params 至少包含:
//
//	to          (string) — 收件人
//	template_id (int)    — 模板 ID
//	variables   (map)    — 渲染变量(对应模板中 {{var_name}} 占位符)
func (n *SMTPNamespace) SendTemplate(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/smtp/send/template", &requestOpts{Body: params, Headers: h})
}

// GetQuota returns the current subscription period quota and usage.
//
// GET /api/v1/smtp/quota
//
// 返回 {today_used, today_quota, period_used, period_quota, expires_at}.
func (n *SMTPNamespace) GetQuota() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/smtp/quota", &requestOpts{Headers: h})
}

// GetStatus queries delivery status of a sent message.
//
// GET /api/v1/smtp/status/:message_id
//
// 返回 {message_id, status, sent_at, opened_at, clicked_at, error_msg, ...}.
func (n *SMTPNamespace) GetStatus(messageID string) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", fmt.Sprintf("/api/v1/smtp/status/%s", messageID), &requestOpts{Headers: h})
}
