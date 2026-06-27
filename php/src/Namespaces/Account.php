<?php
/**
 * 账户命名空间.
 *
 * 对应平台「账户」模块的 v1 公开接口(余额 / 充值地址查询).
 *
 * 鉴权:X-API-Key + X-Secret-Key 双 header(MPK 商户密钥,与 VCard 命名空间共用).
 * 凭据来自控制台「账号 → API 密钥」,SDK 配置字段 api_key / api_secret.
 *
 *   GET /api/v1/account/balance          getBalance          查询账户余额
 *   GET /api/v1/account/deposit-address  getDepositAddress   查询充值地址
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

class Account
{
    private Client $client;

    public function __construct(Client $c) { $this->client = $c; }

    /**
     * 组装双密钥鉴权头(与 Energy::authHeaders() 同形式,但读 api_key / api_secret).
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
     * 查询账户余额.
     *
     * GET /api/v1/account/balance
     *
     * @return array<string, mixed> {balance, frozen, currency, ...}
     * @throws NexCoreError
     */
    public function getBalance(): array
    {
        return $this->client->http->request('GET', '/api/v1/account/balance', [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 查询充值地址.
     *
     * GET /api/v1/account/deposit-address
     *
     * @return array<string, mixed> {address, chain, ...}
     * @throws NexCoreError
     */
    public function getDepositAddress(): array
    {
        return $this->client->http->request('GET', '/api/v1/account/deposit-address', [
            'headers' => $this->authHeaders(),
        ]);
    }
}
