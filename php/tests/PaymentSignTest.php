<?php
/**
 * NexCore PHP SDK — Payment 签名算法测试.
 *
 * 跨语言一致性 fixture(PHP / Python / Node / Go 4 个 SDK 跑出同样结果):
 *   key      = "test-key-abc-123"
 *   params   = { app_id, amount, currency, out_order_id, trade_type }
 *   expected = 44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72
 *
 * 不依赖真实后端 / 不发 HTTP 请求,纯算法验证.
 *
 * 运行:
 *   php tests/PaymentSignTest.php
 */

declare(strict_types=1);

require_once __DIR__ . '/../src/NexCoreError.php';
require_once __DIR__ . '/../src/Http.php';
require_once __DIR__ . '/../src/Namespaces/Payment.php';
require_once __DIR__ . '/../src/Namespaces/Exchange.php';
require_once __DIR__ . '/../src/Namespaces/Energy.php';
require_once __DIR__ . '/../src/Namespaces/Smtp.php';
require_once __DIR__ . '/../src/Namespaces/Withdraw.php';
require_once __DIR__ . '/../src/Client.php';

use NexCore\Client;

// 全局测试计数
$passed = 0;
$failed = 0;

function assertEq($actual, $expected, string $name): void {
    global $passed, $failed;
    if ($actual === $expected) {
        echo "  ✓ $name\n";
        $passed++;
    } else {
        echo "  ✗ $name\n    expected: " . var_export($expected, true) . "\n    actual:   " . var_export($actual, true) . "\n";
        $failed++;
    }
}

echo "=== Payment::sign() — 跨语言 fixture ===\n";

$client = new Client([
    'base_url'        => 'https://example.com',
    'payment_app_id'  => 'APP_TEST',
    'payment_app_key' => 'test-key-abc-123',
]);

$params = [
    'app_id'       => 'APP_TEST',
    'amount'       => '100.00',
    'currency'     => 'CNY',
    'out_order_id' => 'ORDER_001',
    'trade_type'   => 'usdt.trc20',
];
$sign = $client->payment->sign($params);
assertEq($sign, '44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72', 'canonical fixture sign');

echo "\n=== Payment::sign() — 空值过滤 ===\n";
$params2 = $params;
$params2['empty_field'] = '';
$params2['null_field']  = null;
assertEq($client->payment->sign($params2), $sign, '空字符串和 null 被过滤,签名跟基础 fixture 一致');

echo "\n=== Payment::sign() — sign 字段自身被过滤 ===\n";
$params3 = $params;
$params3['sign'] = 'tampered-value';
assertEq($client->payment->sign($params3), $sign, 'sign 字段不参与签名');

echo "\n=== Payment::sign() — key 错误时返回不同签名 ===\n";
$client2 = new Client([
    'base_url'        => 'https://example.com',
    'payment_app_id'  => 'APP_TEST',
    'payment_app_key' => 'wrong-key',
]);
$wrongSign = $client2->payment->sign($params);
if ($wrongSign !== $sign) {
    echo "  ✓ 不同 key 产生不同签名\n";
    $passed++;
} else {
    echo "  ✗ 不同 key 应该产生不同签名\n";
    $failed++;
}

echo "\n=== verifyNotifySign() ===\n";
$signedPayload = array_merge($params, ['sign' => $sign]);
assertEq($client->payment->verifyNotifySign($signedPayload), true, '正确签名通过');

$tamperedPayload = array_merge($params, ['sign' => '0000000000000000000000000000000000000000000000000000000000000000']);
assertEq($client->payment->verifyNotifySign($tamperedPayload), false, '伪造签名拒绝');

assertEq($client->payment->verifyNotifySign($params), false, '缺 sign 字段拒绝');

echo "\n=== 结果 ===\n";
echo "  ✓ $passed passed";
if ($failed > 0) {
    echo " · ✗ $failed failed\n";
    exit(1);
}
echo "\n";
exit(0);
