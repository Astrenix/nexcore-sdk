/**
 * NexCore Node.js SDK — 创建支付订单(轮播模式)
 *
 * 运行:
 *   node examples/create_order.js
 */

const { Client, NexCoreError } = require('../index');

const client = new Client({
  baseUrl: process.env.NEXCORE_BASE_URL || 'https://your-domain.com',
  paymentAppId: process.env.NEXCORE_APP_ID || 'APP20260412XXXX',
  paymentAppKey: process.env.NEXCORE_APP_KEY || 'your_app_key_here',
  timeout: 30000,
});

(async () => {
  try {
    const order = await client.payment.createOrder({
      out_order_id: `ORDER_${Date.now()}`,
      amount: '100.00',            // 必填:法币金额(string,两位小数,避免浮点)
      currency: 'CNY',             // CNY / USD / EUR / JPY / KRW / HKD
      trade_type: 'usdt.trc20',    // 加密币种.链
      call_type: 'rotation',       // rotation=轮播 / one_to_one=一对一
      timeout: 1800,
      subject: '会员充值',
      notify_url: 'https://your-domain.com/payment/notify',
      return_url: 'https://your-domain.com/payment/success',
    });

    console.log('✅ 订单创建成功');
    console.log(`  订单号:    ${order.order_id}`);
    console.log(`  支付地址:  ${order.pay_address}`);
    console.log(`  加密金额:  ${order.crypto_amount} ${order.crypto_currency}`);
    console.log(`  过期时间:  ${order.expires_at}`);
  } catch (e) {
    if (e instanceof NexCoreError) {
      console.error(`❌ Error #${e.code}: ${e.message}`);
      if (e.requestId) console.error(`  Trace ID: ${e.requestId}`);
      if (e.httpStatus) console.error(`  HTTP: ${e.httpStatus}`);
      process.exit(1);
    }
    throw e;
  }
})();
