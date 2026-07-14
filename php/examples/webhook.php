<?php
/**
 * Tovanix PHP SDK — Webhook 回调签名校验.
 *
 * 部署:把本文件放到可公开访问的 URL(如 https://your-domain.com/payment/notify),
 *      然后在用户后台 "应用配置" 的 notify_url 填这个 URL.
 *
 * Tovanix 支付成功后会 POST 一个 JSON 到这里,你必须:
 *   1. 校验签名(SDK 提供 verifyNotifySign 一行搞定)
 *   2. 处理订单状态(发货 / 更新 DB),务必幂等
 *   3. 返回 200 OK(否则平台会重试)
 */

require_once __DIR__ . '/../vendor/autoload.php';

use NexCore\Client;

$client = new Client([
    'base_url'        => getenv('NEXCORE_BASE_URL') ?: 'https://your-domain.com',
    'payment_app_id'  => getenv('NEXCORE_APP_ID')   ?: 'APP20260412XXXX',
    'payment_app_key' => getenv('NEXCORE_APP_KEY')  ?: 'your_app_key_here',
]);

// 1. 读 raw body
$raw     = file_get_contents('php://input');
$payload = json_decode($raw, true);

if (!is_array($payload)) {
    http_response_code(400);
    echo 'invalid payload';
    exit;
}

// 2. 校验签名(常量时间比较,防时序攻击)
if (!$client->payment->verifyNotifySign($payload)) {
    http_response_code(400);
    echo 'invalid sign';
    error_log("[nexcore] sign 校验失败: " . substr($raw, 0, 300));
    exit;
}

// 3. 业务处理(示例)
// 同一订单可能因网络重试收到多次回调,务必做幂等(DB 唯一索引 out_order_id 等)
$outOrder = $payload['out_order_id'] ?? '';
$status   = (int) ($payload['status'] ?? 0);
$amount   = $payload['amount']        ?? '';
$txHash   = $payload['tx_hash']       ?? '';

// 状态枚举:1=已支付  2=待支付  3=已关闭  4=已退款
if ($status === 1) {
    error_log("[nexcore] 订单已支付: {$outOrder} = {$amount} (tx: {$txHash})");
    // TODO: DB 查 out_order_id,判断是否已发货,未发货才发货
}

// 4. 必须返回 200
http_response_code(200);
echo 'OK';
