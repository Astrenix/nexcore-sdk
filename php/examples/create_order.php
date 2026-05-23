<?php
/**
 * NexCore PHP SDK — 创建支付订单(轮播模式)
 *
 * 执行:
 *   php examples/create_order.php
 */

require_once __DIR__ . '/../Client.php';

use NexCore\Client;
use NexCore\NexCoreError;

$client = new Client([
    'base_url'        => getenv('NEXCORE_BASE_URL') ?: 'https://your-domain.com',
    'payment_app_id'  => getenv('NEXCORE_APP_ID')   ?: 'APP20260412XXXX',
    'payment_app_key' => getenv('NEXCORE_APP_KEY')  ?: 'your_app_key_here',
    'timeout'         => 30,
]);

try {
    $order = $client->payment->createOrder([
        'out_order_id' => 'ORDER_' . time(),
        'amount'       => '100.00',           // 必填:法币金额(两位小数 string,避免浮点精度)
        'currency'     => 'CNY',              // 法币:CNY / USD / EUR / JPY / KRW / HKD
        'trade_type'   => 'usdt.trc20',       // 加密币种.链
        'call_type'    => 'rotation',         // rotation=轮播 / 一对一=one_to_one
        'timeout'      => 1800,               // 订单 30 分钟过期
        'subject'      => '会员充值',
        'notify_url'   => 'https://your-domain.com/payment/notify',
        'return_url'   => 'https://your-domain.com/payment/success',
    ]);

    echo "✅ 订单创建成功\n";
    echo "  订单号:    {$order['order_id']}\n";
    echo "  支付地址:  {$order['pay_address']}\n";
    echo "  加密金额:  {$order['crypto_amount']} {$order['crypto_currency']}\n";
    echo "  过期时间:  {$order['expires_at']}\n";

} catch (NexCoreError $e) {
    echo "❌ Error #{$e->code}: {$e->getMessage()}\n";
    if ($e->requestId) echo "  Trace ID: {$e->requestId}\n";
    if ($e->httpStatus) echo "  HTTP: {$e->httpStatus}\n";
    exit(1);
}
