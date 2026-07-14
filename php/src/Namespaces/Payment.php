<?php
/**
 * 多链收款命名空间.
 *
 * 对应 /docs 文档 "多链收款" 模块的全部 v1 公开接口.
 *
 * 鉴权:HMAC-SHA256 签名 — 所有请求自动追加 app_id + sign 字段.
 * 签名算法:把所有参数按 key 升序拼接成 k1=v1&k2=v2,然后用 app_key 做 HMAC-SHA256.
 */

declare(strict_types=1);

namespace NexCore\Namespaces;

use NexCore\Client;
use NexCore\NexCoreError;

/**
 * Payment 实现以下 7 个 v1 endpoint(对照 internal/handler/order.go + one_to_one.go):
 *
 *   POST /api/v1/pay/create          createOrder        创建收款订单
 *   GET  /api/v1/pay/query           queryOrder         查询订单状态
 *   POST /api/v1/pay/close           closeOrder         关闭订单
 *   GET  /api/v1/pay/app-config      getAppConfig       查询当前应用配置
 *   POST /api/v1/pay/bind-address    bindAddress        一对一 — 绑定收款地址
 *   POST /api/v1/pay/get-address     getUserAddress     一对一 — 查询用户已绑地址
 *   POST /api/v1/pay/unbind-address  unbindAddress      一对一 — 解绑地址
 *
 * 另提供 verifyNotifySign() 校验 webhook 回调签名(常量时间比较防时序攻击).
 */
class Payment
{
    private Client $client;

    public function __construct(Client $c) { $this->client = $c; }

    /**
     * 计算 HMAC-SHA256 签名.
     *
     * 业务方一般不需要直接调,SDK 内部自动调用.公开出来便于:
     *   - 自行测试签名是否正确(对照 /docs 文档输出)
     *   - 校验回调签名(verifyNotifySign 内部也用)
     *
     * @param array<string, mixed> $params 待签名参数(会自动过滤 sign 字段和空值,按 key 升序排)
     * @return string 64 字符小写 hex 签名
     * @throws NexCoreError payment_app_key 未配置
     */
    public function sign(array $params): string
    {
        $key = $this->client->get('payment_app_key');
        if (!$key) {
            throw new NexCoreError('payment_app_key not configured', -1);
        }
        unset($params['sign']);
        $params = array_filter($params, fn($v) => $v !== '' && $v !== null);
        ksort($params);

        $pairs = [];
        foreach ($params as $k => $v) {
            $pairs[] = "$k=$v";
        }
        return hash_hmac('sha256', implode('&', $pairs), $key);
    }

    /**
     * 自动注入 app_id + 计算 sign,返回签好的参数.
     *
     * @param array<string, mixed> $params 业务参数
     * @return array<string, mixed> 业务参数 + app_id + sign
     */
    private function signed(array $params): array
    {
        $appId = $this->client->get('payment_app_id');
        if (!$appId) {
            throw new NexCoreError('payment_app_id not configured', -1);
        }
        $params['app_id'] = $appId;
        $params['sign']   = $this->sign($params);
        return $params;
    }

    /**
     * 创建收款订单.
     *
     * POST /api/v1/pay/create
     *
     * 签名注意(对齐后端 internal/service/order.go 验签逻辑):
     *   - amount 入签时按两位小数归一(如 "100.5" → 签 "100.50"),发送 body 保持原值
     *   - timeout 恒入签:未传时按 "0" 参与签名(body 不发该字段)
     *
     * @param array{
     *     out_order_id: string,       商户侧订单号(必须唯一)
     *     amount: string|float,       法币金额,两位小数 string 避免浮点误差
     *     currency: string,           法币:CNY/USD/EUR/JPY/KRW/HKD
     *     trade_type: string,         加密币种.链,如 usdt.trc20 / trx / eth
     *     call_type: string,          必填.本接口只支持 rotation(轮播);一对一模式请走 bindAddress
     *     out_user_id?: string,       商户侧用户 ID(一对一模式)
     *     timeout?: int,              订单过期秒数,默认 1800
     *     subject?: string,           订单描述
     *     notify_url?: string,        webhook 回调 URL
     *     return_url?: string,        支付成功后跳转 URL
     * } $params
     * @return array<string, mixed> 返回 {order_id, out_order_id, amount, currency, crypto_amount,
     *                              crypto_currency, crypto_symbol, pay_address, pay_url, qrcode_url,
     *                              payment_page_url, status, created_at, expired_at}
     * @throws NexCoreError
     */
    public function createOrder(array $params): array
    {
        // 签名参数与发送 body 分离:后端验签时 amount 用两位小数归一、timeout 恒入签(未传按 "0")
        $signParams = $params;
        if (isset($signParams['amount'])) {
            $signParams['amount'] = number_format((float) $signParams['amount'], 2, '.', '');
        }
        if (!isset($signParams['timeout']) || $signParams['timeout'] === '' || $signParams['timeout'] === null) {
            $signParams['timeout'] = '0';
        }
        $signed = $this->signed($signParams);

        // 发送 body 保持用户原值(timeout 不传就不发),只附加签好的 app_id + sign
        $body            = $params;
        $body['app_id']  = $signed['app_id'];
        $body['sign']    = $signed['sign'];

        return $this->client->http->request('POST', '/api/v1/pay/create', [
            'body' => $body,
        ]);
    }

    /**
     * 查询订单当前状态.
     *
     * GET /api/v1/pay/query
     *
     * @param string $outOrderId 商户订单号
     * @return array<string, mixed> 返回 {order_id, status, amount, paid_at, tx_hash, ...}
     * @throws NexCoreError
     */
    public function queryOrder(string $outOrderId): array
    {
        return $this->client->http->request('GET', '/api/v1/pay/query', [
            'query' => $this->signed(['out_order_id' => $outOrderId]),
        ]);
    }

    /**
     * 主动关闭订单.
     *
     * POST /api/v1/pay/close
     *
     * @param string $outOrderId 商户订单号
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function closeOrder(string $outOrderId): array
    {
        return $this->client->http->request('POST', '/api/v1/pay/close', [
            'body' => $this->signed(['out_order_id' => $outOrderId]),
        ]);
    }

    /**
     * 查询应用当前配置(启用的币种 / 支付模式 / 回调 URL 等).
     *
     * GET /api/v1/pay/app-config
     *
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function getAppConfig(): array
    {
        return $this->client->http->request('GET', '/api/v1/pay/app-config', [
            'query' => $this->signed([]),
        ]);
    }

    /**
     * 一对一模式 — 给用户绑定一个固定收款地址.
     *
     * POST /api/v1/pay/bind-address
     *
     * @param string $userId    用户 ID(商户侧)
     * @param string $tradeType 加密币种.链
     * @return array<string, mixed> 返回 {user_id, address, chain, bind_at}
     * @throws NexCoreError
     */
    public function bindAddress(string $userId, string $tradeType): array
    {
        return $this->client->http->request('POST', '/api/v1/pay/bind-address', [
            'body' => $this->signed(['user_id' => $userId, 'trade_type' => $tradeType]),
        ]);
    }

    /**
     * 一对一模式 — 查询用户已绑定的地址.
     *
     * POST /api/v1/pay/get-address
     * (注意:后端是 POST,不是 GET)
     *
     * 签名注意:本接口签名只含 app_id + user_id(**没有 trade_type**),与 bindAddress 不同.
     *
     * @param string $userId 用户 ID
     * @return array<string, mixed> {user_id, address, chain, status, bind_at}
     * @throws NexCoreError
     */
    public function getUserAddress(string $userId): array
    {
        return $this->client->http->request('POST', '/api/v1/pay/get-address', [
            'body' => $this->signed(['user_id' => $userId]),
        ]);
    }

    /**
     * 一对一模式 — 解绑用户地址.
     *
     * POST /api/v1/pay/unbind-address
     *
     * @param string $userId 用户 ID
     * @return array<string, mixed>
     * @throws NexCoreError
     */
    public function unbindAddress(string $userId): array
    {
        return $this->client->http->request('POST', '/api/v1/pay/unbind-address', [
            'body' => $this->signed(['user_id' => $userId]),
        ]);
    }

    /**
     * 校验 webhook 回调签名.
     *
     * NexCore 平台通过 notify_url 推送 JSON 通知时会带 sign 字段.
     * 本方法用常量时间比较(hash_equals)防止时序攻击.
     *
     * @param array<string, mixed> $payload 回调 JSON 完整解码后的 array
     * @return bool true=签名正确,可信;false=签名错误/缺失,应拒绝该回调
     */
    public function verifyNotifySign(array $payload): bool
    {
        $sign = $payload['sign'] ?? null;
        if (!$sign) {
            return false;
        }
        $expected = $this->sign($payload);
        return hash_equals($expected, $sign);
    }
}
