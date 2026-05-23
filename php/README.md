# NexCore PHP SDK

全能 PHP 客户端,覆盖 Payment / Energy / SMTP / AI 全部 NexCore 业务。

## 环境

- PHP 7.4+(支持 PHP 8.x)
- ext-curl, ext-json(标准发行版都内置)

## 安装

直接拷贝 `Client.php` 到项目:

```bash
cp sdk/php/Client.php your-project/lib/nexcore/
```

或克隆整个目录,在代码里 `require 'sdk/php/Client.php'`。

## 用法

```php
<?php
require_once __DIR__ . '/path/to/Client.php';

use NexCore\Client;
use NexCore\NexCoreError;

$client = new Client([
    'base_url'          => 'https://your-domain.com',
    'payment_app_id'    => 'APP20260412XXXX',
    'payment_app_key'   => 'your_app_key_here',
    'energy_api_key'    => 'energy_api_key_here',
    'energy_secret_key' => 'energy_secret_key_here',
    'ai_api_key'        => 'sk-nc-xxx',
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

    // 估算能量
    $est = $client->energy->estimateEnergy('TXxxxxxxxxxxxxxxxxxxxxx');
    echo "需要能量: {$est['estimated_energy']}\n";

    // AI 对话
    $reply = $client->ai->chat(
        [['role' => 'user', 'content' => '你好']],
        'claude-opus-4-7'
    );
    echo $reply['choices'][0]['message']['content'] . "\n";

} catch (NexCoreError $e) {
    echo "Error #{$e->code}: {$e->getMessage()} (trace: {$e->requestId})\n";
}
```

## 异常

所有错误统一抛 `NexCore\NexCoreError`,字段:

- `$code` — 平台错误码(0 = 成功,其他参见错误码表)
- `getMessage()` — 错误描述
- `$requestId` — 服务端日志追踪 ID(响应头 `X-Trace-Id`)
- `$httpStatus` — HTTP 状态码

## Webhook 签名校验

```php
$payload = json_decode(file_get_contents('php://input'), true);
if (!$client->payment->verifyNotifySign($payload)) {
    http_response_code(400);
    echo 'invalid sign';
    exit;
}
// 处理回调...
```

## 示例

更多示例见 [`examples/`](./examples/) 目录。
