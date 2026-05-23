<div align="center">

# NexCore Official SDKs

**面向开发者的综合数字基础服务平台 · 官方多语言 SDK**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![v3.0.0](https://img.shields.io/badge/version-3.0.0-blue.svg)](./CHANGELOG.md)
[![Website](https://img.shields.io/badge/website-nexcores.net-2563eb.svg)](https://nexcores.net)
[![Docs](https://img.shields.io/badge/docs-online-22c55e.svg)](https://nexcores.net/docs)

**官方网站** · [nexcores.net](https://nexcores.net) &nbsp;|&nbsp;
**API 文档** · [/docs](https://nexcores.net/docs) &nbsp;|&nbsp;
**易记** · [9188.PRO](https://9188.pro)

</div>

---

## 关于 NexCore

**NexCore** 是面向开发者与中小团队的**综合数字基础服务平台**,把"加密支付、跨链兑换、TRON 能量、虚拟卡、云服务、海外通讯、AI 网关"等高频但难自研的能力统一封装为 API,让一个开发者也能跑完跨境业务全链路.

> 自 **2021 年 7 月** 启动以来,服务范围已经从最初的多链收款扩展到 **9 大业务模块**,目前是国内少数同时提供加密金融、海外通讯、AI 接入完整工具栈的开发者平台之一.

**设计哲学**

- 🔌 **一次接入,全栈可用** — 一个 SDK 实例覆盖所有业务,共享 baseUrl + 凭据池
- 🛡️ **自托管友好** — Webhook HMAC 签名 + 常量时间校验,合规可审计
- 🌍 **多语言对等** — PHP / Python / Node.js / Go 行为完全一致,迁移零成本
- 📜 **永远跟文档对齐** — SDK 字段命名与 [`/docs`](https://nexcores.net/docs) 100% 一致,不需要"翻译表"
- 🪶 **零或最少依赖** — Go / Node 零运行时依赖,Python 仅 `requests`,PHP 仅 `ext-curl`

## 平台业务一览

NexCore 平台共 **9 个业务模块**,本 SDK 当前覆盖**核心 4 个**(其余 5 个见线上文档说明):

<table>
<tr>
<th width="33%">业务</th>
<th width="20%">SDK 覆盖</th>
<th>定位</th>
</tr>
<tr>
<td>🪙 <b>多链收款 / Payment</b></td>
<td>✅ <code>client.payment</code></td>
<td>USDT/USDC/TRX/BTC/ETH 等 6 主链加密货币收款,秒级确认,商户自托管</td>
</tr>
<tr>
<td>💱 <b>汇率 / Exchange</b></td>
<td>✅ <code>client.exchange</code></td>
<td>实时加密 ↔ 法币 / 法币 ↔ 法币 汇率服务,Payment 配套</td>
</tr>
<tr>
<td>⚡ <b>TRON 能量租赁 / Energy</b></td>
<td>✅ <code>client.energy</code></td>
<td>TRC20 转账省 60% gas,即租即用,30 秒到账</td>
</tr>
<tr>
<td>📧 <b>SMTP 聚合 API / Smtp</b></td>
<td>✅ <code>client.smtp</code></td>
<td>模板邮件 + 多账号智能轮发 + 打开/点击全跟踪</td>
</tr>
<tr>
<td>🔄 多链闪兑 / Swap</td>
<td>—</td>
<td>任意币 ↔ 任意币,链上充提,30 分钟自动到账</td>
</tr>
<tr>
<td>💳 虚拟信用卡 / Vcard</td>
<td>—</td>
<td>USDT 充值开卡,海外广告 / AI 订阅秒结算,Visa / Mastercard 全球</td>
</tr>
<tr>
<td>☁️ 云服务 / Cloud</td>
<td>—</td>
<td>域名 / 服务器 / DNS / SSL 一站式</td>
</tr>
<tr>
<td>📱 SMS 接码 + 专用邮箱 / SMS</td>
<td>—</td>
<td>海外平台注册首选,60+ 国家真实运营商号源</td>
</tr>
<tr>
<td>🤖 Astrenix AI / AiApi</td>
<td>—</td>
<td>Claude / OpenAI / Gemini 全代理,兼容官方 SDK(直接用 <code>openai</code> SDK 改 base_url)</td>
</tr>
</table>

> ⚠️ Swap / Vcard / Cloud / SMS / Astrenix AI 当前通过线上文档 + Web 控制台调用.SDK 后续按需求扩展.

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
| 🐘 PHP 7.4+ | [`php/`](./php/) | Composer | `composer require nexcore/sdk` |
| 🐍 Python 3.8+ | [`python/`](./python/) | pip | `pip install nexcore-sdk` |
| 🟢 Node.js 16+ | [`node/`](./node/) | npm | `npm install @nexcore/sdk` |
| 🐹 Go 1.21+ | [`go/`](./go/) | go modules | `go get github.com/DoBestone/nexcore-sdk/go` |

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

## 适用场景

- **跨境电商 / 数字商品** — 多链收款 + 法币汇率 + 自动归集
- **海外广告投流团队** — 虚拟卡批量开卡 / USDT 自动充值
- **AI 应用开发者** — 通过 Astrenix AI 网关同时接 Claude/GPT/Gemini,统一计费
- **海外平台运营** — SMS 接码 + 专用邮箱 + SMTP 聚合发送,完整通讯栈
- **DApp / Web3 应用** — TRON 能量租赁 + 多链收款,链上业务标配

## 版本

当前 SDK 版本:**v3.0.0**(跟主仓库 v3.x 主线一致)

## 反馈

- [GitHub Issues](https://github.com/DoBestone/nexcore-sdk/issues) — 公开提问 / Bug 反馈
- NexCore 用户后台「工单」 — 私下技术支持 / 账号问题
- `security@nexcores.net` — 安全漏洞(详见 [`SECURITY.md`](.github/SECURITY.md))

## License

[MIT](./LICENSE) © 2026 NexCore

---

<div align="center">

**🚀 立即开始**

[官方网站](https://nexcores.net) · [完整 API 文档](https://nexcores.net/docs) · [注册账号](https://nexcores.net/register) · [GitHub Issues](https://github.com/DoBestone/nexcore-sdk/issues)

</div>
