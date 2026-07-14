# Tovanix Node.js SDK

Tovanix(原 NexCore)平台全能 Node.js 客户端,覆盖 Payment / Exchange / Energy / SMTP / Withdraw / Account / VCard **7 大命名空间全部 44 个 v1 公开 endpoint**.

**零运行时依赖**(仅 Node.js 标准库 `http`/`https`/`crypto`).

## 环境

- Node.js 16+
- 内置 TypeScript 类型(`index.d.ts`)

## 安装

### 方式一:npm(推荐)

```bash
npm install @nexcore/sdk
# or
pnpm add @nexcore/sdk
# or
yarn add @nexcore/sdk
```

(SDK 包发布到 npm 后)

### 方式二:直接复制

```bash
cp -r sdk/node/{index.js,index.d.ts,package.json,src} your-project/lib/nexcore/
```

## 文件结构

```
index.js                公开入口(re-export Client + NexCoreError)
index.d.ts              TypeScript 类型定义
package.json
src/
├── client.js           主客户端
├── http.js             底层 HTTP 传输
├── errors.js           统一异常 NexCoreError
└── namespaces/
    ├── payment.js      多链收款(7 endpoints)
    ├── exchange.js     汇率(5 endpoints)
    ├── energy.js       TRON 能量租赁(8 endpoints)
    ├── smtp.js         SMTP 聚合 API(6 endpoints)
    ├── withdraw.js     提币(4 endpoints,RSA 签名)
    ├── account.js      账户(2 endpoints)
    └── vcard.js        虚拟信用卡(12 endpoints)
```

## 用法

### CommonJS

```javascript
const { Client, NexCoreError } = require('@nexcore/sdk');

const client = new Client({
  baseUrl: 'https://your-domain.com',
  paymentAppId: 'APP20260412XXXX',
  paymentAppKey: 'your_app_key_here',
  energyApiKey: 'energy_api_key_here',
  energySecretKey: 'energy_secret_key_here',
  smtpApiKey: 'smk_xxx',
  timeout: 30000,
});

(async () => {
  try {
    const order = await client.payment.createOrder({
      out_order_id: `ORDER_${Date.now()}`,
      amount: '100.00',
      currency: 'CNY',
      trade_type: 'usdt.trc20',
      call_type: 'rotation',
    });
    console.log('支付地址:', order.pay_address);

    const rate = await client.exchange.getRate('USDT', 'CNY');
    console.log('USDT/CNY:', rate.rate);

    const est = await client.energy.estimateEnergy('TXxxxxxxxxxxxxxxxxxxxxx');
    console.log('建议能量:', est.suggested_energy);

    const mail = await client.smtp.send({
      to: 'user@example.com',
      subject: '验证码',
      body: '<h1>123456</h1>',
      is_html: true,
    });
    console.log('消息 ID:', mail.message_id);
  } catch (e) {
    if (e instanceof NexCoreError) {
      console.error(`Error #${e.code}: ${e.message} (trace: ${e.requestId})`);
    }
  }
})();
```

### TypeScript / ESM

```typescript
import { Client, NexCoreError } from '@nexcore/sdk';

const client = new Client({ baseUrl: 'https://your-domain.com', /* ... */ });
const order = await client.payment.createOrder({ /* ... */ });
```

完整类型定义在 [`index.d.ts`](./index.d.ts).

## API 列表

### `client.payment` — 多链收款(7 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `createOrder(params)` | POST | `/api/v1/pay/create` |
| `queryOrder(outOrderId)` | GET | `/api/v1/pay/query` |
| `closeOrder(outOrderId)` | POST | `/api/v1/pay/close` |
| `getAppConfig()` | GET | `/api/v1/pay/app-config` |
| `bindAddress(userId, tradeType)` | POST | `/api/v1/pay/bind-address` |
| `getUserAddress(userId)` | POST | `/api/v1/pay/get-address` |
| `unbindAddress(userId)` | POST | `/api/v1/pay/unbind-address` |
| `sign(params)` | (工具) | HMAC-SHA256 签名 |
| `verifyNotifySign(payload)` | (工具) | webhook 校验(常量时间) |

### `client.exchange` — 汇率(5 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getRate(from, to)` | GET | `/api/v1/rate` |
| `convert(from, to, amount)` | POST | `/api/v1/convert` |
| `getRates(symbols, base?)` | GET | `/api/v1/rates` |
| `getFiatRates(base)` | GET | `/api/v1/rates/fiat` |
| `getAllRates(base)` | GET | `/api/v1/rates/all` |

注:`getRates` 的 `base` 可不传,由后端取默认(USDT).

### `client.energy` — TRON 能量租赁(8 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getInfo()` | GET | `/api/v1/energy/info` |
| `getPrice(energyAmount, period)` | GET | `/api/v1/energy/price?energy_amount=&period=` |
| `estimateEnergy(toAddress)` | GET | `/api/v1/energy/estimate-energy?to_address=` |
| `createOrder(params)` | POST | `/api/v1/energy/order` |
| `createOnetimeOrder(params)` | POST | `/api/v1/energy/order/onetime` |
| `queryOrder(serial)` | GET | `/api/v1/energy/order/:serial` |
| `listOrders(filter)` | GET | `/api/v1/energy/orders` |
| `reclaimOrder(serial)` | POST | `/api/v1/energy/order/reclaim` |

注:租期 `period` 枚举 `1H / 1D / 3D / 7D / 30D`;`createOrder` 必填 `receive_address` / `energy_amount` / `period`,可选 `out_trade_no` / `remark`.

### `client.smtp` — SMTP 聚合(6 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `send(params, opts?)` | POST | `/api/v1/smtp/send` |
| `sendBatch(params, opts?)` | POST | `/api/v1/smtp/send/batch` |
| `sendTemplate(params)` | POST | `/api/v1/smtp/send/template` |
| `getQuota()` | GET | `/api/v1/smtp/quota` |
| `getStatus(messageId)` | GET | `/api/v1/smtp/status/:message_id` |
| `reportInbound(params)` | POST | `/api/v1/smtp/inbound` |

- `send` 可选字段:`from_name` / `reply_to` / `text_body` / `headers` / `cc` / `bcc` / `attachments` / `account_id` / `send_at`(定时,RFC3339);`opts.idempotencyKey` 写入 `Idempotency-Key` 幂等头
- `sendBatch` 必填 `recipients` 数组(元素 `{to, variables?, from_name?}`),静态 `subject`+`body` 或 `template_code` 二选一;同样支持 `opts.idempotencyKey`
- `sendTemplate` 必填 `to` + `template_code`,可选 `variables` / `from_name`
- `getQuota` 返回 `daily_limit/daily_used/daily_remaining` / `monthly_*` / `expire_at`
- `reportInbound` 上报退信/投诉(`email` 与 `message_id` 至少其一,`type` = `bounce` | `complaint`)

### `client.withdraw` — 提币(4 endpoint,RSA-PKCS1v15-SHA256 签名)

| 方法 | HTTP | endpoint |
|---|---|---|
| `createWithdraw(params)` | POST | `/api/v1/withdraw` |
| `getWithdraw(orderId)` | GET | `/api/v1/withdraw/:id` |
| `getWithdrawableBalance()` | GET | `/api/v1/balance/withdrawable` |
| `quoteFee(chain, symbol, amount)` | GET | `/api/v1/fee/quote`(amount 必填) |
| `sign(...)` / `verifyCallback(...)` | (工具) | RSA 签名 / 平台回调验签 |

### `client.account` — 账户(2 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getBalance()` | GET | `/api/v1/account/balance` |
| `getDepositAddress()` | GET | `/api/v1/account/deposit-address` |

### `client.vcard` — 虚拟信用卡(12 endpoint)

| 方法 | HTTP | endpoint |
|---|---|---|
| `getInfo()` / `listBins()` / `listCards()` | GET | `/api/v1/vcard/*`(读,X-API-Key) |
| `getCardTransactions(cardId)` / `listOrders(query)` / `getOrder(orderId)` | GET | 同上 |
| `updateCardRemark(cardId, remark)` | POST | 同上 |
| `getCardDetails(cardId)` / `getCardCode(cardId)` | GET | 敏感读(HMAC 头签名) |
| `openCard(params)` / `rechargeCard(cardId, params)` / `cancelCard(cardId)` | POST | 资金操作(HMAC 头签名) |

## Webhook 签名校验

```javascript
const express = require('express');
const app = express();
app.use(express.json());

app.post('/payment/notify', (req, res) => {
  if (!client.payment.verifyNotifySign(req.body)) {
    return res.status(400).send('invalid sign');
  }
  // 处理回调... 务必幂等
  res.send('OK');
});
```

`verifyNotifySign` 内部用 `crypto.timingSafeEqual`,常量时间比较防时序攻击.

## 异常

`NexCoreError`:

- `code` — 平台错误码(0=成功)
- `message` — 错误描述
- `requestId` — 服务端追踪 ID(响应头 `X-Trace-Id`)
- `httpStatus` — HTTP 状态码

## 示例

见 [`examples/`](./examples/):
- `create_order.js` — 完整下单
- `webhook_express.js` — Express 接收回调
