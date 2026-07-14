<?php
/**
 * 虚拟信用卡命名空间.
 *
 * 对应平台「虚拟信用卡」模块的 v1 公开接口.
 *
 * 鉴权分两档(凭据 = MPK 商户密钥 api_key / api_secret,与 Account 共用):
 *
 *   1) 只读 / 普通操作 —— X-API-Key + X-Secret-Key 双 header(同 Energy::authHeaders()).
 *      info / bins / cards / transactions / orders / remark
 *
 *   2) 敏感操作(读卡明文 / 开卡 / 充值 / 注销)—— HMAC-SHA256 头签名:
 *      payload = ts . nonce . METHOD . path . rawQuery . body
 *      sig     = hash_hmac('sha256', payload, api_secret)  (小写 hex)
 *      头:X-Key-ID / X-Timestamp / X-Nonce / X-Signature
 *      关键:POST body 先 json_encode 成字符串,对该串签名,再原样作为请求体发送
 *           (Http::request 支持传字符串 body 绕过二次 json_encode,保证签名字节一致).
 *
 * endpoint 一览:
 *
 *   GET  /api/v1/vcard/info                       getInfo               双密钥
 *   GET  /api/v1/vcard/bins                        listBins              双密钥
 *   GET  /api/v1/vcard/cards                       listCards             双密钥
 *   GET  /api/v1/vcard/cards/:id/transactions      getCardTransactions   双密钥
 *   GET  /api/v1/vcard/orders                       listOrders            双密钥
 *   GET  /api/v1/vcard/orders/:id                   getOrder              双密钥
 *   PUT  /api/v1/vcard/cards/:id/remark             updateCardRemark      双密钥
 *   GET  /api/v1/vcard/cards/:id/details            getCardDetails        HMAC 签名
 *   GET  /api/v1/vcard/cards/:id/code               getCardCode           HMAC 签名
 *   POST /api/v1/vcard/cards                         openCard              HMAC 签名
 *   POST /api/v1/vcard/cards/:id/recharge           rechargeCard          HMAC 签名
 *   POST /api/v1/vcard/cards/:id/cancel             cancelCard            HMAC 签名(无 body)
 *
 * 另提供静态方法 verifyWebhook() 校验平台 webhook 回调签名.
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

class VCard
{
    private Client $client;

    public function __construct(Client $c) { $this->client = $c; }

    // ---------- 鉴权 ----------

    /**
     * 双密钥鉴权头(只读 / 普通操作用).
     *
     * @return list<string>
     * @throws NexCoreError 凭据未配置
     */
    private function authHeaders(): array
    {
        $k = $this->client->get('api_key');
        $s = $this->client->get('api_secret');
        if (!$k || !$s) {
            throw new NexCoreError('api_key / api_secret not configured', -1);
        }
        return ["X-API-Key: $k", "X-Secret-Key: $s"];
    }

    /**
     * HMAC 头签名统一发请求(敏感操作用).
     *
     * payload = ts . nonce . METHOD . path . rawQuery . body
     *   - ts:       (string)time()(秒级 unix 时间戳)
     *   - nonce:    bin2hex(random_bytes(8))(一次性随机串)
     *   - METHOD:   HTTP 方法大写
     *   - path:     请求路径(含 :id,不含 query)
     *   - rawQuery: 这些签名端点均无 query,固定 ""
     *   - body:     实际发送的 JSON 字符串(GET / 无 body 为 "")
     *
     * @param string $method
     * @param string $path
     * @param array<string, mixed>|null $body POST body 业务参数;GET / 无 body 传 null
     * @return array<mixed>
     * @throws NexCoreError
     */
    private function signedRequest(string $method, string $path, ?array $body = null): array
    {
        $apiKey    = $this->client->get('api_key');
        $apiSecret = $this->client->get('api_secret');
        if (!$apiKey || !$apiSecret) {
            throw new NexCoreError('api_key / api_secret not configured', -1);
        }

        $method   = strtoupper($method);
        $ts       = (string) time();
        $nonce    = bin2hex(random_bytes(8));
        $rawQuery = '';

        // body 先 json_encode 成字符串,签名与发送都用同一份,保证字节一致
        $bodyStr = '';
        if ($body !== null) {
            $bodyStr = json_encode($body, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            if ($bodyStr === false) {
                throw new NexCoreError('vcard: json_encode body failed', -1);
            }
        }

        $payload = $ts . $nonce . $method . $path . $rawQuery . $bodyStr;
        $sig     = hash_hmac('sha256', $payload, $apiSecret);

        return $this->client->http->request($method, $path, [
            // 传字符串绕过 Http 再次 json_encode;无 body 时传 null 不发请求体
            'body'    => $body === null ? null : $bodyStr,
            'headers' => [
                'X-Key-ID: ' . $apiKey,
                'X-Timestamp: ' . $ts,
                'X-Nonce: ' . $nonce,
                'X-Signature: ' . $sig,
            ],
        ]);
    }

    // ---------- 双密钥(只读 / 普通操作) ----------

    /**
     * 虚拟卡平台信息(可开卡类型 / 费率 / 限额等).
     *
     * GET /api/v1/vcard/info
     *
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function getInfo(): array
    {
        return $this->client->http->request('GET', '/api/v1/vcard/info', [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 列出可用卡 BIN(卡头).
     *
     * GET /api/v1/vcard/bins
     *
     * @return array<string, mixed> {list, ...}
     * @throws NexCoreError
     */
    public function listBins(): array
    {
        return $this->client->http->request('GET', '/api/v1/vcard/bins', [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 列出名下虚拟卡.
     *
     * GET /api/v1/vcard/cards
     *
     * @return array<string, mixed> {list, total, ...}
     * @throws NexCoreError
     */
    public function listCards(): array
    {
        return $this->client->http->request('GET', '/api/v1/vcard/cards', [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 查询单卡交易流水.
     *
     * GET /api/v1/vcard/cards/:id/transactions
     *
     * @param string $cardId 卡 ID
     * @return array<string, mixed> {list, ...}
     * @throws NexCoreError
     */
    public function getCardTransactions(string $cardId): array
    {
        if (!$cardId) {
            throw new NexCoreError('cardId is required', -1);
        }
        return $this->client->http->request('GET', "/api/v1/vcard/cards/$cardId/transactions", [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 列出开卡 / 充值订单.
     *
     * GET /api/v1/vcard/orders
     *
     * @param array{
     *     page?: int,
     *     page_size?: int,
     *     status?: string,
     *     order_type?: string,
     * } $query
     * @return array<string, mixed> {list, total, page, ...}
     * @throws NexCoreError
     */
    public function listOrders(array $query = []): array
    {
        return $this->client->http->request('GET', '/api/v1/vcard/orders', [
            'headers' => $this->authHeaders(),
            'query'   => $query,
        ]);
    }

    /**
     * 查询单笔订单.
     *
     * GET /api/v1/vcard/orders/:id
     *
     * @param string $orderId 订单 ID
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function getOrder(string $orderId): array
    {
        if (!$orderId) {
            throw new NexCoreError('orderId is required', -1);
        }
        return $this->client->http->request('GET', "/api/v1/vcard/orders/$orderId", [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 修改卡备注.
     *
     * PUT /api/v1/vcard/cards/:id/remark
     *
     * @param string $cardId 卡 ID
     * @param string $remark 新备注
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function updateCardRemark(string $cardId, string $remark): array
    {
        if (!$cardId) {
            throw new NexCoreError('cardId is required', -1);
        }
        return $this->client->http->request('PUT', "/api/v1/vcard/cards/$cardId/remark", [
            'headers' => $this->authHeaders(),
            'body'    => ['remark' => $remark],
        ]);
    }

    // ---------- HMAC 签名(敏感操作) ----------

    /**
     * 读取卡明文信息(卡号 / 有效期 / 持卡人等).
     *
     * GET /api/v1/vcard/cards/:id/details
     *
     * @param string $cardId 卡 ID
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function getCardDetails(string $cardId): array
    {
        if (!$cardId) {
            throw new NexCoreError('cardId is required', -1);
        }
        return $this->signedRequest('GET', "/api/v1/vcard/cards/$cardId/details");
    }

    /**
     * 读取卡安全码(CVV).
     *
     * GET /api/v1/vcard/cards/:id/code
     *
     * @param string $cardId 卡 ID
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function getCardCode(string $cardId): array
    {
        if (!$cardId) {
            throw new NexCoreError('cardId is required', -1);
        }
        return $this->signedRequest('GET', "/api/v1/vcard/cards/$cardId/code");
    }

    /**
     * 开卡.
     *
     * POST /api/v1/vcard/cards
     *
     * @param array{
     *     bin_platform_id: int,   卡 BIN 平台 ID(必填,listBins 返回)
     *     amount: float,          开卡金额(必填,> 0)
     * } $params
     * @return array<string, mixed> {order_id, status, total_cost}
     * @throws NexCoreError
     */
    public function openCard(array $params): array
    {
        return $this->signedRequest('POST', '/api/v1/vcard/cards', $params);
    }

    /**
     * 卡充值.
     *
     * POST /api/v1/vcard/cards/:id/recharge
     *
     * @param string $cardId 卡 ID
     * @param array<string, mixed> $params 充值参数 {amount, ...}
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function rechargeCard(string $cardId, array $params): array
    {
        if (!$cardId) {
            throw new NexCoreError('cardId is required', -1);
        }
        return $this->signedRequest('POST', "/api/v1/vcard/cards/$cardId/recharge", $params);
    }

    /**
     * 注销卡(无 body).
     *
     * POST /api/v1/vcard/cards/:id/cancel
     *
     * @param string $cardId 卡 ID
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function cancelCard(string $cardId): array
    {
        if (!$cardId) {
            throw new NexCoreError('cardId is required', -1);
        }
        return $this->signedRequest('POST', "/api/v1/vcard/cards/$cardId/cancel");
    }

    // ---------- Webhook 验签 ----------

    /**
     * 校验平台 webhook 回调签名.
     *
     * 复刻后端签名算法:取所有非空、非 'sign' 字段,按 key 升序(ksort)拼成
     * "k1=v1&k2=v2",用商户 secret 做 HMAC-SHA256(小写 hex),与回调里的
     * sign 字段用 hash_equals 常量时间比较防时序攻击.
     *
     * 注:回调体一般含 sign_ts / nonce 字段,业务方应额外校验 sign_ts 时效
     *     (如 5 分钟内)并对 nonce 做去重,以防重放攻击.本方法只验签名正确性.
     *
     * @param array<string, mixed> $params 回调 JSON 完整解码后的 array(含 sign)
     * @param string $secret 商户密钥(api_secret)
     * @return bool true=签名正确可信;false=签名错误 / 缺失,应拒绝该回调
     */
    public static function verifyWebhook(array $params, string $secret): bool
    {
        $sign = $params['sign'] ?? null;
        if (!$sign || !is_string($sign)) {
            return false;
        }
        unset($params['sign']);
        $params = array_filter($params, fn($v) => $v !== '' && $v !== null);
        ksort($params);

        $pairs = [];
        foreach ($params as $k => $v) {
            $pairs[] = "$k=$v";
        }
        $expected = hash_hmac('sha256', implode('&', $pairs), $secret);
        return hash_equals($expected, $sign);
    }
}
