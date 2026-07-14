package nexcore

import (
	"encoding/json"
	"fmt"
)

// SMTPNamespace implements all 6 SMTP aggregation v1 endpoints.
//
// 对应 /docs 文档 "SMTP API" 模块的全部 v1 公开接口
// (对照 internal/handler/smtp_api.go + smtp_api_ext.go):
//
//	POST /api/v1/smtp/send                 Send           发送单封邮件
//	POST /api/v1/smtp/send/batch           SendBatch      批量发送(每收件人一封独立邮件)
//	POST /api/v1/smtp/send/template        SendTemplate   按模板渲染发送
//	GET  /api/v1/smtp/quota                GetQuota       查询本期配额与用量
//	GET  /api/v1/smtp/status/:message_id   GetStatus      查询邮件投递状态
//	POST /api/v1/smtp/inbound              ReportInbound  上报退信/投诉(自动入黑名单)
//
// 鉴权:Bearer Token — "Authorization: Bearer smk_xxx".
type SMTPNamespace struct {
	c *Client
}

// SMTPSendOptions Send / SendBatch 的可选项.
type SMTPSendOptions struct {
	// IdempotencyKey 幂等键:同 key 重试直接返回首次成功结果,
	// 防网络超时重试导致重复发送 + 双扣配额.通过 Idempotency-Key 请求头传递.
	IdempotencyKey string
}

// applySendOptions 把可选项合并进请求 headers.
func applySendOptions(h map[string]string, opts []SMTPSendOptions) map[string]string {
	for _, o := range opts {
		if o.IdempotencyKey != "" {
			h["Idempotency-Key"] = o.IdempotencyKey
		}
	}
	return h
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
// params 必填:
//
//	to       (string) — 收件人邮箱
//	subject  (string) — 邮件主题
//	body     (string) — 正文(纯文本或 HTML)
//
// 可选:
//
//	is_html     (bool)     — body 是否为 HTML,默认 false
//	from_name   (string)   — 发件人显示名
//	reply_to    (string)   — 回信地址(Reply-To 头)
//	text_body   (string)   — 纯文本版本;HTML 邮件带此值时输出 multipart/alternative 提升送达率
//	headers     (map)      — 自定义邮件头(核心头不可覆盖)
//	cc          ([]string) — 抄送(写 Cc 头 + 投递)
//	bcc         ([]string) — 密送(只投递不写头)
//	attachments ([]object) — 附件列表,元素 {filename, content_base64, content_type}
//	account_id  (int)      — 指定发信账户 ID(默认自动选最优)
//	send_at     (string)   — 定时发送(RFC3339,如 2026-07-01T10:00:00Z);> now+30s 则排期到点发
//
// opts 可传 SMTPSendOptions{IdempotencyKey: "..."} 启用幂等(向后兼容,可不传).
//
// 返回 {message_id, status, account_name, used_smtp, account_id, send_duration_ms};
// 定时分支(send_at 命中排期)返回 {scheduled: true, scheduled_id, send_at}.
func (n *SMTPNamespace) Send(params map[string]any, opts ...SMTPSendOptions) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/smtp/send", &requestOpts{Body: params, Headers: applySendOptions(h, opts)})
}

// SendBatch sends one independent mail per recipient.
//
// POST /api/v1/smtp/send/batch
//
// params 必填:
//
//	recipients ([]object) — 收件人列表,元素 {to(必填), variables?(map), from_name?}
//
// 内容二选一(必须指定其一):
//
//	静态模式 — subject + body 直接传,每人复用同样内容(仍支持 {{var}} 用 variables 逐人替换)
//	模板模式 — template_code 传已保存的模板编码,subject/body 留空,每人 variables 渲染
//
// 其余可选:is_html / reply_to / cc / bcc / attachments / account_id / headers
// (message 级字段,对每个收件人邮件相同;cc/bcc/附件会随每封重复).
//
// 限制:单次 recipients 数量受订阅 max_batch_size 上限(默认 10);逐封扣配额.
//
// opts 可传 SMTPSendOptions{IdempotencyKey: "..."} 启用幂等(向后兼容,可不传).
//
// 返回 {total, success, failed, results: [{to, status, message_id?, error?}]}.
func (n *SMTPNamespace) SendBatch(params map[string]any, opts ...SMTPSendOptions) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/smtp/send/batch", &requestOpts{Body: params, Headers: applySendOptions(h, opts)})
}

// SendTemplate sends mail rendered from a saved template.
//
// POST /api/v1/smtp/send/template
//
// 模板需要先在用户后台 "SMTP API → 模板管理" 创建.
//
// params 必填:
//
//	to            (string) — 收件人邮箱
//	template_code (string) — 模板编码(注意是字符串 code,不是数字 id)
//
// 可选:
//
//	variables (map)    — 渲染变量(对应模板中 {{var_name}} 占位符)
//	from_name (string) — 发件人显示名
//
// 返回 {message_id, status, used_smtp}.
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
// 返回 {daily_limit, daily_used, daily_remaining,
// monthly_limit, monthly_used, monthly_remaining, expire_at}.
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
// 返回发送日志记录,主要字段:
//
//	message_id       (string) — 消息 ID
//	status           (string) — pending / sending / success / failed
//	error_message    (string) — 失败原因
//	smtp_response    (string) — SMTP 服务器响应
//	send_duration_ms (int)    — 发送耗时(毫秒)
//	opened_at / open_count / clicked_at / click_count — 打开/点击追踪(启用追踪时)
//	created_at       (string) — 创建时间
func (n *SMTPNamespace) GetStatus(messageID string) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", fmt.Sprintf("/api/v1/smtp/status/%s", messageID), &requestOpts{Headers: h})
}

// ReportInbound reports a bounce/complaint event (auto-adds to suppression list).
//
// POST /api/v1/smtp/inbound
//
// params:
//
//	email      (string) — 退信/投诉的收件人邮箱
//	message_id (string) — 关联的消息 ID
//	type       (string) — "bounce"(退信)或 "complaint"(投诉)
//
// email / message_id 至少传其一.
//
// 返回 {ok: true}.
func (n *SMTPNamespace) ReportInbound(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/smtp/inbound", &requestOpts{Body: params, Headers: h})
}
