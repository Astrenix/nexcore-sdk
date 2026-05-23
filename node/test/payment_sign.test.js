'use strict';

/**
 * NexCore Node.js SDK — Payment 签名算法测试.
 *
 * 跨语言一致性 fixture(PHP / Python / Node / Go 4 个 SDK 跑出同样结果):
 *   key      = "test-key-abc-123"
 *   params   = { app_id, amount, currency, out_order_id, trade_type }
 *   expected = 44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72
 *
 * 不依赖真实后端 / 不发 HTTP 请求,纯算法验证.
 *
 * 用 Node.js 内置 node:test runner(Node 18+).无需 mocha / jest.
 *
 * 运行:
 *   node --test test/payment_sign.test.js
 */

const test = require('node:test');
const assert = require('node:assert/strict');
const { Client } = require('../index');

const CANONICAL_KEY = 'test-key-abc-123';
const CANONICAL_PARAMS = {
  app_id: 'APP_TEST',
  amount: '100.00',
  currency: 'CNY',
  out_order_id: 'ORDER_001',
  trade_type: 'usdt.trc20',
};
const CANONICAL_SIGN = '44486081415cc5eb5a8b6625c0420ce3812285a44f19d8a1d48dae8ad83edd72';

function newClient(key = CANONICAL_KEY) {
  return new Client({
    baseUrl: 'https://example.com',
    paymentAppId: 'APP_TEST',
    paymentAppKey: key,
  });
}

test('Payment.sign — canonical fixture (cross-language parity)', () => {
  const client = newClient();
  const sign = client.payment.sign(CANONICAL_PARAMS);
  assert.equal(sign, CANONICAL_SIGN);
});

test('Payment.sign — empty / null values are filtered', () => {
  const client = newClient();
  const params = { ...CANONICAL_PARAMS, empty_field: '', null_field: null, undef_field: undefined };
  assert.equal(client.payment.sign(params), CANONICAL_SIGN);
});

test('Payment.sign — `sign` field itself is excluded', () => {
  const client = newClient();
  const params = { ...CANONICAL_PARAMS, sign: 'tampered-value' };
  assert.equal(client.payment.sign(params), CANONICAL_SIGN);
});

test('Payment.sign — wrong key produces different sign', () => {
  const client = newClient('wrong-key');
  const wrongSign = client.payment.sign(CANONICAL_PARAMS);
  assert.notEqual(wrongSign, CANONICAL_SIGN);
  assert.equal(wrongSign.length, 64);
});

test('Payment.verifyNotifySign — valid sign accepted', () => {
  const client = newClient();
  const payload = { ...CANONICAL_PARAMS, sign: CANONICAL_SIGN };
  assert.equal(client.payment.verifyNotifySign(payload), true);
});

test('Payment.verifyNotifySign — tampered sign rejected', () => {
  const client = newClient();
  const payload = { ...CANONICAL_PARAMS, sign: '0'.repeat(64) };
  assert.equal(client.payment.verifyNotifySign(payload), false);
});

test('Payment.verifyNotifySign — missing sign rejected', () => {
  const client = newClient();
  assert.equal(client.payment.verifyNotifySign(CANONICAL_PARAMS), false);
});

test('Payment.verifyNotifySign — null payload rejected', () => {
  const client = newClient();
  assert.equal(client.payment.verifyNotifySign(null), false);
});
