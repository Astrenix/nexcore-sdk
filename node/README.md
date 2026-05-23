# NexCore Node.js SDK

全能 Node.js 客户端,覆盖 Payment / Energy / SMTP / AI 全部 NexCore 业务。**零运行时依赖**(仅 Node.js 标准库)。

## 环境

- Node.js 16+
- 内置 TypeScript 类型(`index.d.ts`)

## 安装

```bash
# 从 npm 安装(SDK 发布到 npm 后)
npm install @nexcore/sdk

# 或直接复制
cp sdk/node/index.js sdk/node/index.d.ts sdk/node/package.json your-project/lib/nexcore/
```

## 用法

```javascript
const { Client, NexCoreError } = require('@nexcore/sdk');

const client = new Client({
  baseUrl: 'https://your-domain.com',
  paymentAppId: 'APP20260412XXXX',
  paymentAppKey: 'your_app_key_here',
  energyApiKey: 'energy_api_key_here',
  energySecretKey: 'energy_secret_key_here',
  aiApiKey: 'sk-nc-xxx',
  timeout: 30000,
});

(async () => {
  try {
    // 创建支付订单
    const order = await client.payment.createOrder({
      out_order_id: `ORDER_${Date.now()}`,
      amount: '100.00',
      currency: 'CNY',
      trade_type: 'usdt.trc20',
      call_type: 'rotation',
      timeout: 1800,
    });
    console.log('支付地址:', order.pay_address);

    // 估算能量
    const est = await client.energy.estimateEnergy('TXxxxxxxxxxxxxxxxxxxxxx');
    console.log('需要能量:', est.estimated_energy);

    // AI 对话
    const reply = await client.ai.chat(
      [{ role: 'user', content: '你好' }],
      'claude-opus-4-7'
    );
    console.log(reply.choices[0].message.content);
  } catch (e) {
    if (e instanceof NexCoreError) {
      console.error(`Error #${e.code}: ${e.message} (trace: ${e.requestId})`);
    } else {
      throw e;
    }
  }
})();
```

## TypeScript / ESM

```typescript
import { Client, NexCoreError } from '@nexcore/sdk';

const client = new Client({ baseUrl: 'https://your-domain.com', /* ... */ });
const order = await client.payment.createOrder({ /* ... */ });
```

## 异常

所有错误统一抛 `NexCoreError`,字段:

- `code` — 平台错误码(0 = 成功)
- `message` — 错误描述
- `requestId` — 服务端日志追踪 ID(响应头 `X-Trace-Id`)
- `httpStatus` — HTTP 状态码

## Webhook 签名校验

```javascript
const express = require('express');
const app = express();
app.use(express.json());

app.post('/payment/notify', (req, res) => {
  if (!client.payment.verifyNotifySign(req.body)) {
    return res.status(400).send('invalid sign');
  }
  // 处理回调...
  res.send('OK');
});
```

签名校验用 `crypto.timingSafeEqual`,常量时间比较,防时序攻击。

## 示例

更多示例见 [`examples/`](./examples/) 目录。
