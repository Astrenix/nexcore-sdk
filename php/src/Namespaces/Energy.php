<?php
/**
 * TRON 能量租赁命名空间.
 *
 * 对应 /docs 文档 "能量租赁" 模块的全部 v1 公开接口.
 *
 * 鉴权:X-API-Key + X-Secret-Key 双 header(注:跟 Payment 的 sign 不同).
 * API Key / Secret Key 在用户后台 "能量租赁 → API 对接" 创建.
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

/**
 * Energy 实现以下 8 个 v1 endpoint(对照 internal/handler/trxx_api.go):
 *
 *   GET  /api/v1/energy/info             getInfo              平台公开信息(可用能量 / 最小最大单 / 阶梯定价)
 *   GET  /api/v1/energy/price            getPrice             指定能量数 + 周期的报价
 *   GET  /api/v1/energy/estimate-energy  estimateEnergy       根据接收地址估算 TRC20 转账所需能量
 *   POST /api/v1/energy/order            createOrder          创建常规租赁订单
 *   POST /api/v1/energy/order/onetime    createOnetimeOrder   创建一次性订单(用完不续)
 *   GET  /api/v1/energy/order/:serial    queryOrder           查询订单(注意是 serial 字符串,不是数字 id)
 *   GET  /api/v1/energy/orders           listOrders           列出所有订单(可按状态过滤)
 *   POST /api/v1/energy/order/reclaim    reclaimOrder         主动回收订单(回收能量到平台,适用部分场景)
 */
class Energy
{
    private Client $client;

    public function __construct(Client $c) { $this->client = $c; }

    /**
     * @return list<string>
     * @throws NexCoreError 凭据未配置
     */
    private function authHeaders(): array
    {
        $k = $this->client->get('energy_api_key');
        $s = $this->client->get('energy_secret_key');
        if (!$k || !$s) {
            throw new NexCoreError('energy_api_key / energy_secret_key not configured', -1);
        }
        return ["X-API-Key: $k", "X-Secret-Key: $s"];
    }

    /**
     * 平台公开信息.
     *
     * GET /api/v1/energy/info
     *
     * @return array<string, mixed> {platform_avail_energy, minimum_order_energy, maximum_order_energy, tiered_pricing, ...}
     */
    public function getInfo(): array
    {
        return $this->client->http->request('GET', '/api/v1/energy/info', [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 获取指定能量数 + 周期的报价.
     *
     * GET /api/v1/energy/price?energy=65000&period=1D
     *
     * @param int    $energy 需要的能量值
     * @param string $period 租期:1H / 6H / 1D / 3D / 1W,默认 1D
     * @return array<string, mixed> {price, period, total, ...}
     */
    public function getPrice(int $energy, string $period = '1D'): array
    {
        return $this->client->http->request('GET', '/api/v1/energy/price', [
            'headers' => $this->authHeaders(),
            'query'   => ['energy' => $energy, 'period' => $period],
        ]);
    }

    /**
     * 根据接收地址估算 TRC20 转账所需能量.
     *
     * GET /api/v1/energy/estimate-energy?receive_addr=TXxxxxxxxx
     *
     * @param string $receiveAddr 收款 TRON 地址(T 开头 Base58)
     * @return array<string, mixed> {estimated_energy, has_usdt_balance, ...}
     */
    public function estimateEnergy(string $receiveAddr): array
    {
        return $this->client->http->request('GET', '/api/v1/energy/estimate-energy', [
            'headers' => $this->authHeaders(),
            'query'   => ['receive_addr' => $receiveAddr],
        ]);
    }

    /**
     * 创建常规租赁订单.
     *
     * POST /api/v1/energy/order
     *
     * @param array{
     *     receive_addr: string,    收能量的目标 TRON 地址
     *     energy: int,             能量数(必须 >= minimum_order_energy)
     *     period: string,          1H / 6H / 1D / 3D / 1W
     *     out_serial?: string,     商户侧订单号(可选,做幂等用)
     * } $params
     * @return array<string, mixed> {serial, status, delegated_at, ...}
     */
    public function createOrder(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/energy/order', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }

    /**
     * 创建一次性订单(用完不续).
     *
     * POST /api/v1/energy/order/onetime
     *
     * 适用场景:用户只做一笔 TRC20 转账,转完即丢能量,不持续占用.
     *
     * @param array<string, mixed> $params 同 createOrder(可不传 period)
     * @return array<string, mixed>
     */
    public function createOnetimeOrder(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/energy/order/onetime', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }

    /**
     * 查询订单状态.
     *
     * GET /api/v1/energy/order/:serial
     *
     * @param string $serial 订单序列号(string,**不是**数字 id)
     * @return array<string, mixed>
     */
    public function queryOrder(string $serial): array
    {
        return $this->client->http->request('GET', "/api/v1/energy/order/$serial", [
            'headers' => $this->authHeaders(),
        ]);
    }

    /**
     * 列出所有订单(可按状态过滤).
     *
     * GET /api/v1/energy/orders
     *
     * @param array<string, mixed> $filter {status?, page?, page_size?, ...}
     * @return array<string, mixed> {list, total, page, ...}
     */
    public function listOrders(array $filter = []): array
    {
        return $this->client->http->request('GET', '/api/v1/energy/orders', [
            'headers' => $this->authHeaders(),
            'query'   => $filter,
        ]);
    }

    /**
     * 主动回收订单(把能量从目标地址收回平台).
     *
     * POST /api/v1/energy/order/reclaim
     *
     * @param array{serial: string} $params {serial}
     * @return array<string, mixed>
     */
    public function reclaimOrder(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/energy/order/reclaim', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }
}
