/**
 * Tovanix Node.js SDK — Webhook 回调签名校验(Express 示例)
 *
 * 部署:
 *   npm install express
 *   node examples/webhook_express.js
 *
 * 然后在 Tovanix 用户后台「应用配置」的 notify_url 填你的 URL。
 *
 * Tovanix 支付成功后会 POST JSON 到这里,本示例:
 *   1. 校验签名(SDK 一行搞定,内部用 crypto.timingSafeEqual 常量时间比较)
 *   2. 业务处理(发货 / 更新 DB,务必幂等)
 *   3. 返回 200 OK(否则平台会重试)
 */

const express = require('express');
const { Client } = require('../index');

const app = express();
app.use(express.json());

const client = new Client({
  baseUrl: process.env.NEXCORE_BASE_URL || 'https://your-domain.com',
  paymentAppId: process.env.NEXCORE_APP_ID || 'APP20260412XXXX',
  paymentAppKey: process.env.NEXCORE_APP_KEY || 'your_app_key_here',
});

app.post('/payment/notify', (req, res) => {
  const payload = req.body;
  if (!payload || typeof payload !== 'object') {
    return res.status(400).send('invalid payload');
  }

  // 1. 校验签名(常量时间比较,防时序攻击)
  if (!client.payment.verifyNotifySign(payload)) {
    console.warn('[nexcore] sign 校验失败:', JSON.stringify(payload).slice(0, 300));
    return res.status(400).send('invalid sign');
  }

  // 2. 业务处理(示例)
  // 同一订单可能因网络重试收到多次回调,务必做幂等(DB 唯一索引 out_order_id 等)
  const { order_id: orderId, out_order_id: outOrder, status, amount, tx_hash: txHash } = payload;

  // 状态:1=已支付  2=待支付  3=已关闭  4=已退款
  if (Number(status) === 1) {
    console.log(`[nexcore] 订单已支付: ${outOrder} = ${amount} (tx: ${txHash})`);
    // TODO: DB 查 out_order_id,判断是否已发货,未发货才发货
  }

  // 3. 必须返回 200
  res.status(200).send('OK');
});

const port = process.env.PORT || 8000;
app.listen(port, () => {
  console.log(`nexcore webhook listening on :${port}/payment/notify`);
});
