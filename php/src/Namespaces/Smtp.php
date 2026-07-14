<?php
/**
 * SMTP 聚合 API 命名空间.
 *
 * 对应 /docs 文档 "SMTP API" 模块的全部 v1 公开接口.
 *
 * 鉴权:Bearer Token — Authorization: Bearer smk_xxx
 * smk_ 前缀的 API Key 在用户后台 "SMTP API → API Key 管理" 创建.
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

/**
 * Smtp 实现以下 6 个 v1 endpoint(对照 internal/handler/smtp_api.go + smtp_api_ext.go):
 *
 *   POST /api/v1/smtp/send                 send           发送单封邮件(支持定时 send_at)
 *   POST /api/v1/smtp/send/batch           sendBatch      批量发送(recipients 数组,每人独立一封)
 *   POST /api/v1/smtp/send/template        sendTemplate   按模板渲染发送(template_code)
 *   GET  /api/v1/smtp/quota                getQuota       查询本期配额与用量
 *   GET  /api/v1/smtp/status/:message_id   getStatus      查询邮件状态(投递成功/失败/打开/点击)
 *   POST /api/v1/smtp/inbound              reportInbound  上报退信(bounce)/投诉(complaint)→ 自动 suppression
 */
class Smtp
{
    private Client $client;

    public function __construct(Client $c) { $this->client = $c; }

    /**
     * @return list<string>
     * @throws NexCoreError 凭据未配置
     */
    private function authHeaders(): array
    {
        $k = $this->client->get('smtp_api_key');
        if (!$k) {
            throw new NexCoreError('smtp_api_key not configured', -1);
        }
        return ["Authorization: Bearer $k"];
    }

    /**
     * 发送单封邮件.
     *
     * POST /api/v1/smtp/send
     *
     * @param array{
     *     to: string,             收件人邮箱(必填)
     *     subject: string,        邮件主题(必填)
     *     body: string,           正文(必填,纯文本或 HTML)
     *     is_html?: bool,         body 是否为 HTML,默认 false
     *     from_name?: string,     发件人显示名(可选)
     *     reply_to?: string,      回信地址(Reply-To 头,可选)
     *     text_body?: string,     纯文本版本;HTML 邮件带此值时输出 multipart/alternative 提升送达率
     *     headers?: array<string, string>,  自定义邮件头(核心头不可覆盖)
     *     cc?: list<string>,      抄送(写 Cc 头 + 投递)
     *     bcc?: list<string>,     密送(只投递不写头)
     *     attachments?: list<array{filename: string, content_base64: string, content_type: string}>,  附件
     *     account_id?: int,       指定发信账户 ID(可选,0/不传=自动选最优;指定后不故障转移)
     *     send_at?: string,       定时发送(RFC3339,如 2026-07-01T10:00:00Z);> now+30s 则排期到点发
     * } $params
     * @param string|null $idempotencyKey 可选 Idempotency-Key 头,防网络超时重试导致重复发送 + 双扣配额
     * @return array<string, mixed> 立即发送:{message_id, status, account_name, used_smtp, account_id, send_duration_ms}
     *                              定时分支(send_at > now+30s):{scheduled: true, scheduled_id, send_at}
     */
    public function send(array $params, ?string $idempotencyKey = null): array
    {
        $headers = $this->authHeaders();
        if ($idempotencyKey !== null && $idempotencyKey !== '') {
            $headers[] = "Idempotency-Key: $idempotencyKey";
        }
        return $this->client->http->request('POST', '/api/v1/smtp/send', [
            'headers' => $headers,
            'body'    => $params,
        ]);
    }

    /**
     * 批量发送(recipients 数组,每个收件人各发一封独立邮件,逐封扣配额).
     *
     * POST /api/v1/smtp/send/batch
     *
     * 两种内容模式(二选一):
     *   - 静态模式:传 subject + body(body 支持 {{var}} 按每人 variables 替换)
     *   - 模板模式:传 template_code + 每人 variables,subject/body 留空
     *
     * 上限:单次收件人数受订阅 max_batch_size 限制,默认 10.
     *
     * @param array{
     *     recipients: list<array{to: string, variables?: array<string, string>, from_name?: string}>,  收件人数组(必填,to 必填)
     *     subject?: string,       静态模式主题
     *     body?: string,          静态模式正文(支持 {{var}} 占位符)
     *     template_code?: string, 模板模式:模板 code(与 subject/body 二选一)
     *     is_html?: bool,
     *     reply_to?: string,      回信地址(每封相同)
     *     cc?: list<string>,      抄送(每封重复)
     *     bcc?: list<string>,     密送(每封重复)
     *     attachments?: list<array{filename: string, content_base64: string, content_type: string}>,  附件(每封重复)
     *     account_id?: int,       指定发信账户 ID
     *     headers?: array<string, string>,  自定义邮件头
     * } $params
     * @param string|null $idempotencyKey 可选 Idempotency-Key 头,防整请求超时重试重复发全部收件人
     * @return array<string, mixed> {total, success, failed, results: list<{to, status, message_id?, error?}>}
     */
    public function sendBatch(array $params, ?string $idempotencyKey = null): array
    {
        $headers = $this->authHeaders();
        if ($idempotencyKey !== null && $idempotencyKey !== '') {
            $headers[] = "Idempotency-Key: $idempotencyKey";
        }
        return $this->client->http->request('POST', '/api/v1/smtp/send/batch', [
            'headers' => $headers,
            'body'    => $params,
        ]);
    }

    /**
     * 按模板渲染发送.模板需要在用户后台 "SMTP API → 模板管理" 创建.
     *
     * POST /api/v1/smtp/send/template
     *
     * @param array{
     *     to: string,                 收件人(必填)
     *     template_code: string,      模板 code(string,必填,**不是**数字模板 ID)
     *     variables?: array<string, string>,  渲染变量(对应模板中 {{var_name}} 占位符)
     *     from_name?: string,         发件人显示名(可选)
     * } $params
     * @return array<string, mixed> {message_id, status, used_smtp}
     */
    public function sendTemplate(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/smtp/send/template', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }

    /**
     * 查询当前订阅期内的配额与已用量.
     *
     * GET /api/v1/smtp/quota
     *
     * @return array<string, mixed> {daily_limit, daily_used, daily_remaining,
     *                               monthly_limit, monthly_used, monthly_remaining, expire_at}
     */
    public function getQuota(): array
    {
        return $this->client->http->request('GET', '/api/v1/smtp/quota', [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 查询指定邮件的投递状态(打开 / 点击 / 退订 / 失败原因).
     *
     * GET /api/v1/smtp/status/:message_id
     *
     * @param string $messageId send / sendBatch / sendTemplate 返回的 message_id
     * @return array<string, mixed> {message_id, status(pending/sending/success/failed), from_email, to_email,
     *                               subject, is_html, account_id, account_name, error_message, smtp_response,
     *                               send_duration_ms, opened_at, open_count, clicked_at, click_count, created_at}
     */
    public function getStatus(string $messageId): array
    {
        return $this->client->http->request('GET', "/api/v1/smtp/status/$messageId", [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 上报退信(bounce)/投诉(complaint),自动加入 suppression 名单并标记 send_log.
     *
     * POST /api/v1/smtp/inbound
     *
     * @param array{
     *     email?: string,        退信/投诉的收件人邮箱(与 message_id 至少传其一)
     *     message_id?: string,   对应的 message_id(与 email 至少传其一)
     *     type?: string,         bounce | complaint
     * } $params
     * @return array<string, mixed> {ok: true}
     */
    public function reportInbound(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/smtp/inbound', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }
}
