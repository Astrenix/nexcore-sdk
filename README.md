# NexCore Official SDKs

NexCore 综合数字基础服务平台官方 SDK 集合 — **一个客户端覆盖全部业务模块**.

🌐 **公开仓库:** https://github.com/DoBestone/nexcore-sdk

## 业务覆盖(25 个 v1 公开 endpoint)

| 模块 | 路径前缀 | 鉴权 | endpoints |
|---|---|---|---|
| 多链收款 Payment | `/api/v1/pay/*` | HMAC-SHA256 签名(app_id + app_key) | **7** |
| 汇率 Exchange | `/api/v1/rate` / `/api/v1/rates*` / `/api/v1/convert` | X-App-Key + X-App-Secret | **5** |
| 能量租赁 Energy | `/api/v1/energy/*` | X-API-Key + X-Secret-Key | **8** |
| SMTP 聚合 Smtp | `/api/v1/smtp/*` | Authorization: Bearer smk_xxx | **5** |

所有 endpoint 完全对照 [`/docs`](https://nexcores.net/docs) 在线文档,字段命名与后端 100% 一致.

## 支持语言

| 语言 | 目录 | 包管理 | 安装 |
|---|---|---|---|
| PHP 7.4+ | [`php/`](./php/) | Composer | `composer require nexcore/sdk` |
| Python 3.8+ | [`python/`](./python/) | pip | `pip install nexcore-sdk` |
| Node.js 16+ | [`node/`](./node/) | npm | `npm install @nexcore/sdk` |
| Go 1.21+ | [`go/`](./go/) | go modules | `go get github.com/DoBestone/nexcore-sdk/go` |

(npm/PyPI/Packagist 发布详见 [`PUBLISH.md`](./PUBLISH.md))

## 统一设计

所有语言遵循同样的客户端结构:

```text
Client(config)
  ├─ .payment    多链收款
  │   ├─ createOrder / queryOrder / closeOrder / getAppConfig
  │   ├─ bindAddress / getUserAddress / unbindAddress  (一对一模式)
  │   └─ sign() / verifyNotifySign()                   (签名工具)
  ├─ .exchange   汇率
  │   ├─ getRate / convert
  │   └─ getRates / getFiatRates / getAllRates
  ├─ .energy     TRON 能量租赁
  │   ├─ getInfo / getPrice / estimateEnergy
  │   ├─ createOrder / createOnetimeOrder
  │   └─ queryOrder / listOrders / reclaimOrder
  └─ .smtp       SMTP 聚合
      ├─ send / sendBatch / sendTemplate
      └─ getQuota / getStatus
```

## 文件组织(规范化)

每个语言 SDK 按业务命名空间拆分多文件,便于阅读和维护:

```
sdk/
├── php/
│   ├── composer.json                (PSR-4 autoload)
│   └── src/
│       ├── Client.php
│       ├── Http.php
│       ├── NexCoreError.php
│       └── Namespaces/
│           ├── Payment.php  Exchange.php  Energy.php  Smtp.php
├── python/
│   ├── pyproject.toml               (pip 可安装)
│   └── nexcore/
│       ├── __init__.py / client.py / http.py / errors.py
│       └── namespaces/
│           ├── payment.py  exchange.py  energy.py  smtp.py
├── node/
│   ├── package.json                 (npm 可发布)
│   ├── index.js / index.d.ts
│   └── src/
│       ├── client.js / http.js / errors.js
│       └── namespaces/
│           ├── payment.js  exchange.js  energy.js  smtp.js
└── go/
    ├── go.mod                       (module: github.com/DoBestone/nexcore-sdk/go)
    ├── doc.go / client.go / http.go / errors.go
    └── payment.go  exchange.go  energy.go  smtp.go
```

## 鉴权配置

一个 Client 实例可同时挂多种凭据,各业务自动用对应凭据.

### Python

```python
from nexcore import Client, NexCoreError

client = Client(
    base_url="https://your-domain.com",
    payment_app_id="APP20260412XXXX",
    payment_app_key="your_app_key_here",
    energy_api_key="energy_key",
    energy_secret_key="energy_secret",
    smtp_api_key="smk_xxx",
)

# 创建支付订单
order = client.payment.create_order(
    out_order_id="ORDER_001",
    amount="100.00",
    currency="CNY",
    trade_type="usdt.trc20",
    call_type="rotation",
)

# 估算 TRC20 能量
estimate = client.energy.estimate_energy("TXxxxxxxxxxxxxxxxxxxxxx")

# 查询汇率
rate = client.exchange.get_rate("USDT", "CNY")

# 发送邮件
result = client.smtp.send(to="user@example.com", subject="Hi", body="Hello")
```

### Node.js

```javascript
const { Client, NexCoreError } = require('@nexcore/sdk');

const client = new Client({
  baseUrl: 'https://your-domain.com',
  paymentAppId: 'APP20260412XXXX',
  paymentAppKey: 'your_app_key_here',
  energyApiKey: 'energy_key',
  energySecretKey: 'energy_secret',
  smtpApiKey: 'smk_xxx',
});

const order = await client.payment.createOrder({
  out_order_id: `ORDER_${Date.now()}`,
  amount: '100.00',
  currency: 'CNY',
  trade_type: 'usdt.trc20',
  call_type: 'rotation',
});
```

### PHP

```php
require_once __DIR__ . '/vendor/autoload.php';
use NexCore\Client;

$client = new Client([
    'base_url'        => 'https://your-domain.com',
    'payment_app_id'  => 'APP20260412XXXX',
    'payment_app_key' => 'your_app_key_here',
    // ...
]);

$order = $client->payment->createOrder([
    'out_order_id' => 'ORDER_' . time(),
    'amount'       => '100.00',
    'currency'     => 'CNY',
    'trade_type'   => 'usdt.trc20',
    'call_type'    => 'rotation',
]);
```

### Go

```go
import nexcore "github.com/DoBestone/nexcore-sdk/go"

c := nexcore.NewClient(nexcore.Config{
    BaseURL:       "https://your-domain.com",
    PaymentAppID:  "APP20260412XXXX",
    PaymentAppKey: "your_app_key_here",
})

raw, err := c.Payment.CreateOrder(map[string]any{
    "out_order_id": fmt.Sprintf("ORDER_%d", time.Now().Unix()),
    "amount":       "100.00",
    "currency":     "CNY",
    "trade_type":   "usdt.trc20",
    "call_type":    "rotation",
})
```

## 异常处理

所有 SDK 统一抛出 `NexCoreError`(各语言适配名),包含:

- `code` — 平台返回错误码(0=成功;-1=客户端层错误)
- `message` — 错误信息
- `requestId` / `request_id` — 服务端日志 ID(响应头 `X-Trace-Id`,排查问题用)
- `httpStatus` / `http_status` — HTTP 状态码

## Webhook 签名校验

四个语言都用**常量时间比较**防时序攻击:

- PHP — `hash_equals()`
- Python — `hmac.compare_digest()`
- Node.js — `crypto.timingSafeEqual()`
- Go — `hmac.Equal()`

每个语言的 `client.payment.verifyNotifySign(payload)` 一行调用即完成校验.

## 版本

当前 SDK 版本:**v3.0.0**(跟主仓库 v3.x 主线一致)

## 反馈

[GitHub Issues](https://github.com/DoBestone/nexcore-sdk/issues) 或 NexCore 用户后台 「工单」.

安全漏洞请发送到 `security@nexcores.net`(详见 [`SECURITY.md`](.github/SECURITY.md)).

## License

[MIT](./LICENSE)
