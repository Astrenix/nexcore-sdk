<?php
/**
 * 提币 API namespace — 多链收款业务的资金出库端.
 *
 * 鉴权:RSA-PKCS1v15-SHA256 签名 + 4 个请求头
 *
 *   X-API-Key            账户级 API Key(控制台「账号 → API 密钥」)
 *   X-Timestamp          unix ms,与服务器时差 ≤ 60s
 *   X-Nonce              一次性 nonce(uuid v4),5 分钟内不可重复
 *   X-Withdraw-Signature RSA-PKCS1v15-SHA256(caller_private_key, signString),Base64
 *
 * signString = METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + BODY
 * 其中 BODY 为 HTTP body 原文(JSON 字符串原样,GET 请求为空字符串).
 *
 * 对应 /docs 文档 "提币 API" 章节的 4 个 endpoint:
 *
 *   POST /api/v1/withdraw                 createWithdraw            发起提币
 *   GET  /api/v1/withdraw/:id             getWithdraw               查询单笔状态
 *   GET  /api/v1/balance/withdrawable     getWithdrawableBalance    查询可提余额
 *   GET  /api/v1/fee/quote                quoteFee                  费用预估
 *
 * 另提供 verifyCallback() 校验平台回调签名(用平台公钥).
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

class Withdraw
{
    private Client $client;

    /** @var resource|\OpenSSLAsymmetricKey|null 缓存的对接方私钥 */
    private $privKey = null;

    /** @var resource|\OpenSSLAsymmetricKey|null 缓存的平台公钥 */
    private $platformPub = null;

    public function __construct(Client $client)
    {
        $this->client = $client;
    }

    // ---------- 内部:密钥懒解析 ----------

    private function getPrivKey()
    {
        if ($this->privKey !== null) {
            return $this->privKey;
        }
        $pem = $this->client->get('withdraw_private_key_pem');
        if (!$pem) {
            throw new NexCoreError('withdraw_private_key_pem not configured', -1);
        }
        if (!extension_loaded('openssl')) {
            throw new NexCoreError('提币 API 需要 ext-openssl,请先启用 PHP 的 openssl 扩展', -1);
        }
        $key = openssl_pkey_get_private($pem);
        if ($key === false) {
            throw new NexCoreError('withdraw: invalid private key PEM: ' . openssl_error_string(), -1);
        }
        $details = openssl_pkey_get_details($key);
        if (!$details || ($details['type'] ?? null) !== OPENSSL_KEYTYPE_RSA) {
            throw new NexCoreError('withdraw: configured private key is not RSA', -1);
        }
        $this->privKey = $key;
        return $key;
    }

    private function getPlatformPub()
    {
        if ($this->platformPub !== null) {
            return $this->platformPub;
        }
        $pem = $this->client->get('withdraw_platform_public_key_pem');
        if (!$pem) {
            throw new NexCoreError('withdraw_platform_public_key_pem not configured', -1);
        }
        if (!extension_loaded('openssl')) {
            throw new NexCoreError('提币 API 需要 ext-openssl,请先启用 PHP 的 openssl 扩展', -1);
        }
        $key = openssl_pkey_get_public($pem);
        if ($key === false) {
            throw new NexCoreError('withdraw: invalid platform public key PEM: ' . openssl_error_string(), -1);
        }
        $details = openssl_pkey_get_details($key);
        if (!$details || ($details['type'] ?? null) !== OPENSSL_KEYTYPE_RSA) {
            throw new NexCoreError('withdraw: platform key is not RSA', -1);
        }
        $this->platformPub = $key;
        return $key;
    }

    // ---------- 签名 ----------

    /**
     * 计算请求的 RSA-PKCS1v15-SHA256 签名(Base64).
     *
     * 业务方一般不需要直接调,SDK 内部 do() 时自动调用.
     * 公开出来便于测试 / 自实现非标场景(比如 curl 调试).
     */
    public function sign(string $method, string $path, string $timestamp, string $nonce, string $body): string
    {
        $priv = $this->getPrivKey();
        $signString = strtoupper($method) . "\n" . $path . "\n" . $timestamp . "\n" . $nonce . "\n" . $body;
        $sig = '';
        if (!openssl_sign($signString, $sig, $priv, OPENSSL_ALGO_SHA256)) {
            throw new NexCoreError('withdraw: openssl_sign failed: ' . openssl_error_string(), -1);
        }
        return base64_encode($sig);
    }

    /**
     * 生成 uuid v4(不依赖 ext-uuid).
     */
    private function newNonce(): string
    {
        $b = random_bytes(16);
        $b[6] = chr(ord($b[6]) & 0x0f | 0x40);
        $b[8] = chr(ord($b[8]) & 0x3f | 0x80);
        return vsprintf('%s%s-%s-%s-%s-%s%s%s', str_split(bin2hex($b), 4));
    }

    /**
     * 内部统一发请求 — 自动加 4 个鉴权头.
     *
     * @param string $method
     * @param string $path
     * @param array<string, mixed>|null $body 业务参数(会 json_encode 一次,签名和发送都用同一份字符串)
     * @param array<string, mixed> $query
     * @return array<mixed>
     */
    private function do(string $method, string $path, ?array $body = null, array $query = []): array
    {
        $apiKey = $this->client->get('withdraw_api_key');
        if (!$apiKey) {
            throw new NexCoreError('withdraw_api_key not configured', -1);
        }
        $timestamp = (string) intval(microtime(true) * 1000);
        $nonce = $this->newNonce();
        $bodyStr = '';
        if ($body !== null) {
            $bodyStr = json_encode($body, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            if ($bodyStr === false) {
                throw new NexCoreError('withdraw: json_encode body failed', -1);
            }
        }
        $sig = $this->sign($method, $path, $timestamp, $nonce, $bodyStr);
        return $this->client->http->request($method, $path, [
            'body'    => $body === null ? null : $bodyStr,  // 传字符串绕过 Http 再次 encode
            'query'   => $query,
            'headers' => [
                'X-API-Key: ' . $apiKey,
                'X-Timestamp: ' . $timestamp,
                'X-Nonce: ' . $nonce,
                'X-Withdraw-Signature: ' . $sig,
            ],
        ]);
    }

    // ---------- 公开 endpoint ----------

    /**
     * 发起提币 — POST /api/v1/withdraw.
     *
     * 下单后状态为 pending,等延迟到期由 worker 自动广播.
     * 期间可在控制台暂停 / 加速 / 取消.
     *
     * @param array{
     *     chain: string,
     *     symbol: string,
     *     amount: string,
     *     to_address: string,
     *     memo?: string,
     *     callback_url?: string,
     *     request_id?: string,
     * } $params
     * @return array<string, mixed>
     */
    public function createWithdraw(array $params): array
    {
        return $this->do('POST', '/api/v1/withdraw', $params);
    }

    /**
     * 查询单笔提币状态 — GET /api/v1/withdraw/:id.
     */
    public function getWithdraw(string $orderId): array
    {
        if (!$orderId) {
            throw new NexCoreError('orderId is required', -1);
        }
        return $this->do('GET', '/api/v1/withdraw/' . $orderId);
    }

    /**
     * 查询可提余额 — GET /api/v1/balance/withdrawable.
     *
     * 返回该账户在每条链 × 每种资产下的「已归集待提现」余额.
     * 只有这部分可用于 API 提币.
     */
    public function getWithdrawableBalance(): array
    {
        return $this->do('GET', '/api/v1/balance/withdrawable');
    }

    /**
     * 费用预估 — GET /api/v1/fee/quote.
     *
     * @return array{chain: string, symbol: string, amount?: string, fee_amount: string, fee_asset: string}
     */
    public function quoteFee(string $chain, string $symbol, ?string $amount = null): array
    {
        if (!$chain || !$symbol) {
            throw new NexCoreError('chain and symbol are required', -1);
        }
        $q = ['chain' => $chain, 'symbol' => $symbol];
        if ($amount !== null && $amount !== '') {
            $q['amount'] = $amount;
        }
        return $this->do('GET', '/api/v1/fee/quote', null, $q);
    }

    // ---------- 回调验签 ----------

    /**
     * 验证平台回调签名(对接方收到 webhook 时调用).
     *
     * 用法:
     *   $sig   = $_SERVER['HTTP_X_PLATFORM_SIGNATURE'] ?? '';
     *   $ts    = $_SERVER['HTTP_X_TIMESTAMP'] ?? '';
     *   $nonce = $_SERVER['HTTP_X_NONCE'] ?? '';
     *   $body  = file_get_contents('php://input');  // 原始 body 字符串,不要 re-encode
     *   try {
     *       $client->withdraw->verifyCallback($_SERVER['REQUEST_METHOD'], $_SERVER['REQUEST_URI'], $ts, $nonce, $body, $sig);
     *   } catch (NexCoreError $e) {
     *       http_response_code(401);
     *       exit;
     *   }
     *
     * 验签算法与请求方向一致:RSA-PKCS1v15-SHA256(platform_public_key, signString).
     *
     * @throws NexCoreError 验签失败抛出
     */
    public function verifyCallback(
        string $method,
        string $path,
        string $timestamp,
        string $nonce,
        string $body,
        string $base64Signature
    ): void {
        $pub = $this->getPlatformPub();
        $sig = base64_decode($base64Signature, true);
        if ($sig === false) {
            throw new NexCoreError('withdraw: bad signature base64', -1);
        }
        $signString = strtoupper($method) . "\n" . $path . "\n" . $timestamp . "\n" . $nonce . "\n" . $body;
        $result = openssl_verify($signString, $sig, $pub, OPENSSL_ALGO_SHA256);
        if ($result !== 1) {
            throw new NexCoreError('withdraw: signature verify failed', -1);
        }
    }
}
