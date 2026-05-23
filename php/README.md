# NexCore PHP SDK

全能 PHP 客户端,覆盖 Payment / Exchange / Energy / SMTP **全部 25 个 v1 公开 endpoint**.

## 环境

- PHP 7.4+(支持 PHP 8.x)
- ext-curl, ext-json

## 安装

### 方式一:Composer(推荐)

```bash
composer require nexcore/sdk
```

(SDK 包发布到 Packagist 后)

### 方式二:直接复制

```bash
cp -r sdk/php/src your-project/lib/nexcore/
# 然后在代码里 require src/Client.php(及其他依赖文件)
```

## 文件结构

```
src/
├── Client.php             主客户端入口
├── Http.php               底层 HTTP 传输
├── NexCoreError.php       统一异常
└── Namespaces/
    ├── Payment.php        多链收款(7 endpoints)
    ├── Exchange.php       汇率(5 endpoints)
    ├── Energy.php         TRON 能量租赁(8 endpoints)
    └── Smtp.php           SMTP 聚合 API(5 endpoints)
```

## 用法

```php
<?php
require_once __DIR__ . '/vendor/autoload.php';

use NexCore\Client;
use NexCore\NexCoreError;

$client = new Client([
    'base_url'          => 'https://your-domain.com',
    'payment_app_id'    => 'APP20260412XXXX',
    'payment_app_key'   => 'your_app_key_here',
    'energy_api_key'    => 'energy_api_key_here',
    'energy_secret_key' => 'energy_secret_key_here',
    'smtp_api_key'      => 'smk_xxx',
    'timeout'           => 30,
]);

try {
    // 创建支付订单
    $order = $client->payment->createOrder([
        'out_order_id' => 'ORDER_' . time(),
        'amount'       => '100.00',
        'currency'     => 'CNY',
        'trade_type'   => 'usdt.trc20',
        'call_type'    => 'rotation',
        'timeout'      => 1800,
    ]);
    echo "支付地址: {$order['pay_address']}\n";

    // 查询汇率
    $rate = $client->exchange->getRate('USDT', 'CNY');
    echo "USDT/CNY: {$rate['rate']}\n";

    // 估算 TRC20 能量
    $est = $client->energy->estimateEnergy('TXxxxxxxxxxxxxxxxxxxxxx');
    echo "需要能量: {$est['estimated_energy']}\n";

    // 发送邮件
    $mail = $client->smtp->send([
        'to'      => 'user@example.com',
        'subject' => '验证码',
        'body'    => '<h1>123456</h1>',
        'is_html' => true,
    ]);
    echo "消息 ID: {$mail['message_id']}\n";

} catch (NexCoreError $e) {
    echo "Error #{$e->code}: {$e->getMessage()} (trace: {$e->requestId})\n";
}
```

## API 列表

### `$client->payment` — 多链收款(7 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `createOrder($params)` | POST | `/api/v1/pay/create` |
| `queryOrder($outOrderId)` | GET | `/api/v1/pay/query` |
| `closeOrder($outOrderId)` | POST | `/api/v1/pay/close` |
| `getAppConfig()` | GET | `/api/v1/pay/app-config` |
| `bindAddress($userId, $tradeType)` | POST | `/api/v1/pay/bind-address` |
| `getUserAddress($userId, $tradeType)` | POST | `/api/v1/pay/get-address` |
| `unbindAddress($userId)` | POST | `/api/v1/pay/unbind-address` |
| `sign($params)` | (工具) | HMAC-SHA256 签名 |
| `verifyNotifySign($payload)` | (工具) | webhook 校验(常量时间比较) |

### `$client->exchange` — 汇率(5 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getRate($from, $to)` | GET | `/api/v1/rate` |
| `convert($from, $to, $amount)` | POST | `/api/v1/convert` |
| `getRates($symbols, $base)` | GET | `/api/v1/rates` |
| `getFiatRates($base)` | GET | `/api/v1/rates/fiat` |
| `getAllRates($base)` | GET | `/api/v1/rates/all` |

### `$client->energy` — TRON 能量租赁(8 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getInfo()` | GET | `/api/v1/energy/info` |
| `getPrice($energy, $period)` | GET | `/api/v1/energy/price` |
| `estimateEnergy($receiveAddr)` | GET | `/api/v1/energy/estimate-energy` |
| `createOrder($params)` | POST | `/api/v1/energy/order` |
| `createOnetimeOrder($params)` | POST | `/api/v1/energy/order/onetime` |
| `queryOrder($serial)` | GET | `/api/v1/energy/order/:serial` |
| `listOrders($filter)` | GET | `/api/v1/energy/orders` |
| `reclaimOrder($params)` | POST | `/api/v1/energy/order/reclaim` |

### `$client->smtp` — SMTP 聚合(5 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `send($params)` | POST | `/api/v1/smtp/send` |
| `sendBatch($params)` | POST | `/api/v1/smtp/send/batch` |
| `sendTemplate($params)` | POST | `/api/v1/smtp/send/template` |
| `getQuota()` | GET | `/api/v1/smtp/quota` |
| `getStatus($messageId)` | GET | `/api/v1/smtp/status/:message_id` |

## Webhook 签名校验

```php
$payload = json_decode(file_get_contents('php://input'), true);
if (!$client->payment->verifyNotifySign($payload)) {
    http_response_code(400);
    echo 'invalid sign';
    exit;
}
// 处理回调... 同一订单可能重试,务必幂等
http_response_code(200);
```

`verifyNotifySign` 内部用 `hash_equals` 常量时间比较,防时序攻击.

## 异常

`NexCore\NexCoreError`:

- `$code` — 平台错误码(0=成功,-1=客户端层)
- `getMessage()` — 错误描述
- `$requestId` — 服务端追踪 ID(响应头 `X-Trace-Id`)
- `$httpStatus` — HTTP 状态码

## 示例

见 [`examples/`](./examples/):
- `create_order.php` — 完整下单流程
- `webhook.php` — 接收回调
