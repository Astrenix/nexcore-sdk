<?php
/**
 * 汇率服务命名空间.
 *
 * 对应 /docs 文档 "多链收款 → 汇率服务" 5 个 endpoint.
 * 走 APIAuth 中间件,用 X-App-Key + X-App-Secret(应用密钥)进行 header 鉴权.
 *
 * 注:汇率接口跟 Payment 业务接口虽然在文档同一模块下,但鉴权机制不同 —
 *   Payment 用 ?sign= query 签名,Exchange 用 header.
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

/**
 * Exchange 实现以下 5 个 v1 endpoint(对照 internal/handler/exchange_api.go):
 *
 *   GET  /api/v1/rate          getRate        单对币种汇率
 *   POST /api/v1/convert       convert        金额换算
 *   GET  /api/v1/rates         getRates       批量获取多币种到指定基准币的汇率
 *   GET  /api/v1/rates/fiat    getFiatRates   主流法币到指定基准法币的汇率
 *   GET  /api/v1/rates/all     getAllRates    所有支持币种的汇率快照
 */
class Exchange
{
    private Client $client;

    public function __construct(Client $c) { $this->client = $c; }

    /**
     * 构造 APIAuth header.汇率接口用应用密钥(payment_app_id + payment_app_key).
     *
     * @return list<string>
     * @throws NexCoreError 凭据未配置
     */
    private function authHeaders(): array
    {
        $appId  = $this->client->get('payment_app_id');
        $appKey = $this->client->get('payment_app_key');
        if (!$appId || !$appKey) {
            throw new NexCoreError('payment_app_id / payment_app_key not configured', -1);
        }
        return [
            "X-App-Key: $appId",
            "X-App-Secret: $appKey",
        ];
    }

    /**
     * 查询单对币种汇率.
     *
     * GET /api/v1/rate?from=USDT&to=CNY
     *
     * @param string $from 来源币种代码(USDT / TRX / ETH / BTC / USD / CNY ...)
     * @param string $to   目标币种代码
     * @return array<string, mixed> {from, to, rate, updated_at}
     */
    public function getRate(string $from, string $to): array
    {
        return $this->client->http->request('GET', '/api/v1/rate', [
            'headers' => $this->authHeaders(),
            'query'   => ['from' => $from, 'to' => $to],
        ]);
    }

    /**
     * 金额换算.
     *
     * POST /api/v1/convert
     * Body: {from, to, amount}
     *
     * @param string       $from   来源币种
     * @param string       $to     目标币种
     * @param string|float $amount 待换算金额
     * @return array<string, mixed> {from_amount, to_amount, rate}
     */
    public function convert(string $from, string $to, $amount): array
    {
        return $this->client->http->request('POST', '/api/v1/convert', [
            'headers' => $this->authHeaders(),
            'body'    => ['from' => $from, 'to' => $to, 'amount' => $amount],
        ]);
    }

    /**
     * 批量获取多币种到指定基准币的汇率.
     *
     * GET /api/v1/rates?symbols=USDT,TRX,ETH&base=CNY
     *
     * @param list<string> $symbols 待查询的币种代码列表
     * @param string       $base    基准币(报价单位),默认 CNY
     * @return array<string, mixed> {base, rates: {USDT: 7.23, TRX: 0.85, ...}, updated_at}
     */
    public function getRates(array $symbols, string $base = 'CNY'): array
    {
        return $this->client->http->request('GET', '/api/v1/rates', [
            'headers' => $this->authHeaders(),
            'query'   => ['symbols' => implode(',', $symbols), 'base' => $base],
        ]);
    }

    /**
     * 主流法币到指定基准法币的汇率.
     *
     * GET /api/v1/rates/fiat?base=USD
     *
     * @param string $base 基准法币,默认 USD
     * @return array<string, mixed>
     */
    public function getFiatRates(string $base = 'USD'): array
    {
        return $this->client->http->request('GET', '/api/v1/rates/fiat', [
            'headers' => $this->authHeaders(),
            'query'   => ['base' => $base],
        ]);
    }

    /**
     * 所有支持币种的汇率快照(加密币 + 法币).
     *
     * GET /api/v1/rates/all?base=USDT
     *
     * @param string $base 基准币,默认 USDT
     * @return array<string, mixed>
     */
    public function getAllRates(string $base = 'USDT'): array
    {
        return $this->client->http->request('GET', '/api/v1/rates/all', [
            'headers' => $this->authHeaders(),
            'query'   => ['base' => $base],
        ]);
    }
}
