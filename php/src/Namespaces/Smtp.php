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
 * Smtp 实现以下 5 个 v1 endpoint(对照 internal/handler/smtp_api.go + smtp_api_ext.go):
 *
 *   POST /api/v1/smtp/send                 send           发送单封邮件
 *   POST /api/v1/smtp/send/batch           sendBatch      批量发送(同主题/正文,多收件人)
 *   POST /api/v1/smtp/send/template        sendTemplate   按模板渲染发送
 *   GET  /api/v1/smtp/quota                getQuota       查询本期配额与用量
 *   GET  /api/v1/smtp/status/:message_id   getStatus      查询邮件状态(投递成功/失败/打开/点击)
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
     *     to: string,             收件人邮箱
     *     subject: string,        邮件主题
     *     body: string,           正文(纯文本或 HTML)
     *     is_html?: bool,         body 是否为 HTML,默认 false
     *     account_id?: int,       指定发信账户 ID(可选,默认自动选最优)
     *     reply_to?: string,      回信地址(可选)
     * } $params
     * @return array<string, mixed> {message_id, status}
     */
    public function send(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/smtp/send', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }

    /**
     * 批量发送(同主题/正文,多收件人).
     *
     * POST /api/v1/smtp/send/batch
     *
     * @param array{
     *     to: list<string>,       收件人邮箱列表
     *     subject: string,        统一主题
     *     body: string,           统一正文
     *     is_html?: bool,
     *     account_id?: int,
     * } $params
     * @return array<string, mixed> {message_ids, total, accepted}
     */
    public function sendBatch(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/smtp/send/batch', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }

    /**
     * 按模板渲染发送.模板需要在用户后台 "SMTP API → 模板管理" 创建.
     *
     * POST /api/v1/smtp/send/template
     *
     * @param array{
     *     to: string,                 收件人
     *     template_id: int,           模板 ID
     *     variables: array<string, mixed>,  渲染变量(对应模板中 {{var_name}} 占位符)
     *     account_id?: int,
     * } $params
     * @return array<string, mixed> {message_id, status}
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
     * @return array<string, mixed> {today_used, today_quota, period_used, period_quota, expires_at}
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
     * @return array<string, mixed> {message_id, status, sent_at, opened_at, clicked_at, error_msg, ...}
     */
    public function getStatus(string $messageId): array
    {
        return $this->client->http->request('GET', "/api/v1/smtp/status/$messageId", [
            'headers' => $this->authHeaders(),
        ]);
    }
}
