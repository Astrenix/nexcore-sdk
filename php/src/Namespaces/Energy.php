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
     * GET /api/v1/energy/price?period=1D&energy_amount=65000
     *
     * @param int    $energyAmount 需要的能量值(query key = energy_amount)
     * @param string $period       租期:1H / 1D / 3D / 7D / 30D,默认 1D
     * @return array<string, mixed> {period, energy_amount, price_trx}
     */
    public function getPrice(int $energyAmount, string $period = '1D'): array
    {
        return $this->client->http->request('GET', '/api/v1/energy/price', [
            'headers' => $this->authHeaders(),
            'query'   => ['period' => $period, 'energy_amount' => $energyAmount],
        ]);
    }

    /**
     * 根据接收地址估算 TRC20 转账所需能量.
     *
     * GET /api/v1/energy/estimate-energy?to_address=TXxxxxxxxx
     *
     * @param string $toAddress 收款 TRON 地址(T 开头 Base58,34 位)
     * @return array<string, mixed> {to_address, initialized, suggested_energy}
     */
    public function estimateEnergy(string $toAddress): array
    {
        return $this->client->http->request('GET', '/api/v1/energy/estimate-energy', [
            'headers' => $this->authHeaders(),
            'query'   => ['to_address' => $toAddress],
        ]);
    }

    /**
     * 创建常规租赁订单.
     *
     * POST /api/v1/energy/order
     *
     * @param array{
     *     receive_address: string, 收能量的目标 TRON 地址
     *     energy_amount: int,      能量数(必须 >= minimum_order_energy)
     *     period: string,          1H / 1D / 3D / 7D / 30D
     *     out_trade_no?: string,   商户侧订单号(可选)
     *     remark?: string,         备注(可选)
     * } $params
     * @return array<string, mixed> {serial, price_trx, deducted_usd}
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
     * 注意:与 createOrder 不同,**没有 energy_amount 字段**,能量数由平台按目标地址估算.
     *
     * @param array{
     *     receive_address: string, 收能量的目标 TRON 地址(必填)
     *     period: string,          1H / 1D / 3D / 7D / 30D(必填)
     *     out_trade_no?: string,   商户侧订单号(可选)
     *     remark?: string,         备注(可选)
     * } $params
     * @return array<string, mixed> {serial, price_trx, deducted_usd}
     *                              price_trx 按上游实际结算,预估为上界,多退少不补
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
     * @return array<string, mixed> {serial, receive_address, energy_amount, period, price_trx,
     *                               status, status_msg, out_trade_no, order_type, created_at}
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
     * @param array<string, mixed> $filter {status?, page?, page_size?}
     *                                     status 枚举:-1=全部(默认) / 0=待处理 / 40=成功 / 41=失败
     * @return array<string, mixed> {list, total, page, page_size}
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
     * @return array<string, mixed> {errno, message}
     */
    public function reclaimOrder(array $params): array
    {
        return $this->client->http->request('POST', '/api/v1/energy/order/reclaim', [
            'headers' => $this->authHeaders(),
            'body'    => $params,
        ]);
    }
}
