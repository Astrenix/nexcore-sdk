<?php
/**
 * NexCore SDK 底层 HTTP 传输层.
 *
 * 仅依赖 ext-curl,不依赖 Guzzle / Symfony HTTP Client.
 * 业务命名空间(Payment / Exchange / Energy / Smtp)调用本类发请求,
 * 不直接接触 curl,保持各 namespace 关注业务逻辑.
 */

declare(strict_types=1);

namespace NexCore;

/**
 * Http 封装 curl 调用,处理:
 *   - URL + query 拼接
 *   - JSON body 编码
 *   - 公共 header 注入(Content-Type / Accept / User-Agent)
 *   - 响应解包(自动解 envelope 拿 .data)
 *   - 错误统一抛 NexCoreError(含 X-Trace-Id)
 */
class Http
{
    private string $baseUrl;
    private int $timeout;
    private bool $verifySsl;
    private string $userAgent;

    public function __construct(string $baseUrl, int $timeout = 30, bool $verifySsl = true, string $userAgent = '')
    {
        $this->baseUrl   = rtrim($baseUrl, '/');
        $this->timeout   = $timeout;
        $this->verifySsl = $verifySsl;
        $this->userAgent = $userAgent ?: ('NexCore-PHP-SDK/' . Client::VERSION);
    }

    /**
     * 发送 HTTP 请求.
     *
     * @param string $method HTTP 方法,GET / POST / PUT / DELETE
     * @param string $path  路径,以 / 开头,如 "/api/v1/pay/create"
     * @param array{
     *     body?: array|string|null,
     *     query?: array<string, mixed>,
     *     headers?: list<string>,
     * } $opts 选项:
     *     - body:    JSON 序列化后作为请求体,可传字符串绕过序列化
     *     - query:   query string 参数,自动 url-encode
     *     - headers: 额外 header 行,格式 "Header-Name: value"
     * @return array<mixed> 业务响应的 data 段(自动解 { code, message, data } envelope)
     * @throws NexCoreError 网络错误 / HTTP 4xx-5xx / 业务 code != 0
     */
    public function request(string $method, string $path, array $opts = []): array
    {
        $url   = $this->baseUrl . $path;
        $body  = $opts['body'] ?? null;
        $hdrs  = $opts['headers'] ?? [];
        $query = $opts['query'] ?? [];

        if (!empty($query)) {
            $url .= (str_contains($url, '?') ? '&' : '?') . http_build_query($query);
        }

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, strtoupper($method));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $this->timeout);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->verifySsl);
        curl_setopt($ch, CURLOPT_HEADER, true);

        $defaults = [
            'Content-Type: application/json',
            'Accept: application/json',
            'User-Agent: ' . $this->userAgent,
        ];
        curl_setopt($ch, CURLOPT_HTTPHEADER, array_merge($defaults, $hdrs));

        if ($body !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS,
                is_string($body) ? $body : json_encode($body, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));
        }

        $resp = curl_exec($ch);
        if ($resp === false) {
            $err = curl_error($ch);
            curl_close($ch);
            throw new NexCoreError("HTTP request failed: $err", -1);
        }

        $status  = (int) curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $hdrSize = (int) curl_getinfo($ch, CURLINFO_HEADER_SIZE);
        $rawHdr  = substr($resp, 0, $hdrSize);
        $rawBody = substr($resp, $hdrSize);
        curl_close($ch);

        // 提取 X-Trace-Id(服务端日志追踪用,排查问题时给后端工单提供本值)
        $traceId = null;
        if (preg_match('/^X-Trace-Id:\s*(.+?)\r?$/im', $rawHdr, $m)) {
            $traceId = trim($m[1]);
        }

        $json = json_decode($rawBody, true);

        if ($status >= 400 || !is_array($json)) {
            throw new NexCoreError(
                $json['message'] ?? "HTTP $status: " . substr($rawBody, 0, 200),
                $json['code'] ?? -1,
                $traceId,
                $status
            );
        }
        if (isset($json['code']) && $json['code'] !== 0) {
            throw new NexCoreError(
                $json['message'] ?? 'unknown error',
                (int) $json['code'],
                $traceId,
                $status
            );
        }

        return $json['data'] ?? $json;
    }
}
