# Tovanix PHP SDK

Tovanix(原 NexCore)平台全能 PHP 客户端,覆盖 Payment / Exchange / Energy / SMTP / Withdraw / Account / VCard **7 大命名空间全部 44 个 v1 公开 endpoint**.

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
    ├── Smtp.php           SMTP 聚合 API(6 endpoints)
    ├── Withdraw.php       提币(4 endpoints,RSA 签名)
    ├── Account.php        账户(2 endpoints)
    └── VCard.php          虚拟信用卡(12 endpoints)
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
    echo "建议能量: {$est['suggested_energy']}\n";

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
| `getUserAddress($userId)` | POST | `/api/v1/pay/get-address` |
| `unbindAddress($userId)` | POST | `/api/v1/pay/unbind-address` |
| `sign($params)` | (工具) | HMAC-SHA256 签名 |
| `verifyNotifySign($payload)` | (工具) | webhook 校验(常量时间比较) |

### `$client->exchange` — 汇率(5 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getRate($from, $to)` | GET | `/api/v1/rate` |
| `convert($from, $to, $amount)` | POST | `/api/v1/convert` |
| `getRates($symbols, $base = '')` | GET | `/api/v1/rates` |
| `getFiatRates($base)` | GET | `/api/v1/rates/fiat` |
| `getAllRates($base)` | GET | `/api/v1/rates/all` |

注:`getRates` 的 `$base` 不传时由后端取默认(USDT).

### `$client->energy` — TRON 能量租赁(8 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getInfo()` | GET | `/api/v1/energy/info` |
| `getPrice($energyAmount, $period = '1D')` | GET | `/api/v1/energy/price?energy_amount=&period=` |
| `estimateEnergy($toAddress)` | GET | `/api/v1/energy/estimate-energy?to_address=` |
| `createOrder($params)` | POST | `/api/v1/energy/order` |
| `createOnetimeOrder($params)` | POST | `/api/v1/energy/order/onetime` |
| `queryOrder($serial)` | GET | `/api/v1/energy/order/:serial` |
| `listOrders($filter)` | GET | `/api/v1/energy/orders` |
| `reclaimOrder($params)` | POST | `/api/v1/energy/order/reclaim` |

注:租期 `$period` 枚举 `1H / 1D / 3D / 7D / 30D`;`createOrder` 必填 `receive_address` / `energy_amount` / `period`,可选 `out_trade_no` / `remark`.

### `$client->smtp` — SMTP 聚合(6 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `send($params, $idempotencyKey = null)` | POST | `/api/v1/smtp/send` |
| `sendBatch($params, $idempotencyKey = null)` | POST | `/api/v1/smtp/send/batch` |
| `sendTemplate($params)` | POST | `/api/v1/smtp/send/template` |
| `getQuota()` | GET | `/api/v1/smtp/quota` |
| `getStatus($messageId)` | GET | `/api/v1/smtp/status/:message_id` |
| `reportInbound($params)` | POST | `/api/v1/smtp/inbound` |

- `send` 可选字段:`from_name` / `reply_to` / `text_body` / `headers` / `cc` / `bcc` / `attachments` / `account_id` / `send_at`(定时,RFC3339);`$idempotencyKey` 写入 `Idempotency-Key` 幂等头
- `sendBatch` 必填 `recipients` 数组(元素 `{to, variables?, from_name?}`),静态 `subject`+`body` 或 `template_code` 二选一;同样支持 `$idempotencyKey`
- `sendTemplate` 必填 `to` + `template_code`,可选 `variables` / `from_name`
- `getQuota` 返回 `daily_limit/daily_used/daily_remaining` / `monthly_*` / `expire_at`
- `reportInbound` 上报退信/投诉(`email` 与 `message_id` 至少其一,`type` = `bounce` | `complaint`)

### `$client->withdraw` — 提币(4 endpoint,RSA-PKCS1v15-SHA256 签名)

| 方法 | HTTP | endpoint |
|---|---|---|
| `createWithdraw($params)` | POST | `/api/v1/withdraw` |
| `getWithdraw($orderId)` | GET | `/api/v1/withdraw/:id` |
| `getWithdrawableBalance()` | GET | `/api/v1/balance/withdrawable` |
| `quoteFee($chain, $symbol, $amount)` | GET | `/api/v1/fee/quote`(`$amount` 必填) |
| `sign(...)` / `verifyCallback(...)` | (工具) | RSA 签名 / 平台回调验签 |

### `$client->account` — 账户(2 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getBalance()` | GET | `/api/v1/account/balance` |
| `getDepositAddress()` | GET | `/api/v1/account/deposit-address` |

### `$client->vcard` — 虚拟信用卡(12 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getInfo()` / `listBins()` / `listCards()` | GET | `/api/v1/vcard/*`(读,X-API-Key) |
| `getCardTransactions($cardId)` / `listOrders($query)` / `getOrder($orderId)` | GET | 同上 |
| `updateCardRemark($cardId, $remark)` | POST | 同上 |
| `getCardDetails($cardId)` / `getCardCode($cardId)` | GET | 敏感读(HMAC 头签名) |
| `openCard($params)` / `rechargeCard($cardId, $params)` / `cancelCard($cardId)` | POST | 资金操作(HMAC 头签名) |

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
