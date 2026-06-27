'use strict';

/**
 * NexCore Official Node.js SDK 主客户端.
 *
 * 一次配置覆盖 NexCore 平台全部 v1 公开接口,业务按 namespace 划分:
 *
 *   client.payment   — 多链收款(HMAC-SHA256 签名)
 *   client.exchange  — 汇率(X-App-Key + X-App-Secret header)
 *   client.energy    — TRON 能量租赁(X-API-Key + X-Secret-Key)
 *   client.smtp      — SMTP 聚合(Bearer Token)
 *
 * @example
 *   const { Client, NexCoreError } = require('@nexcore/sdk');
 *
 *   const client = new Client({
 *     baseUrl: 'https://your-domain.com',
 *     paymentAppId: 'APP20260412XXXX',
 *     paymentAppKey: 'your_app_key_here',
 *     energyApiKey: 'energy_key',
 *     energySecretKey: 'energy_secret',
 *     smtpApiKey: 'smk_xxx',
 *   });
 *
 *   const order = await client.payment.createOrder({
 *     out_order_id: `ORDER_${Date.now()}`,
 *     amount: '100.00',
 *     currency: 'CNY',
 *     trade_type: 'usdt.trc20',
 *     call_type: 'rotation',
 *   });
 */

const crypto = require('crypto');

const { NexCoreError } = require('./errors');
const { Http } = require('./http');
const { Payment } = require('./namespaces/payment');
const { Exchange } = require('./namespaces/exchange');
const { Energy } = require('./namespaces/energy');
const { Smtp } = require('./namespaces/smtp');
const { Withdraw } = require('./namespaces/withdraw');
const { Account } = require('./namespaces/account');
const { VCard } = require('./namespaces/vcard');

class Client {
  /**
   * @param {object} config
   * @param {string} config.baseUrl - NexCore 平台基础 URL
   * @param {string} [config.paymentAppId] - 多链收款 / 汇率应用 ID
   * @param {string} [config.paymentAppKey] - 多链收款 / 汇率应用密钥
   * @param {string} [config.energyApiKey] - 能量租赁 X-API-Key
   * @param {string} [config.energySecretKey] - 能量租赁 X-Secret-Key
   * @param {string} [config.smtpApiKey] - SMTP smk_ 前缀 Token
   * @param {string} [config.apiKey] - MPK 商户 API Key(account / vcard 共用,X-API-Key / X-Key-ID)
   * @param {string} [config.apiSecret] - MPK 商户 API Secret(account / vcard 共用,X-Secret-Key / HMAC 签名密钥)
   * @param {string} [config.withdrawApiKey] - 提币 X-API-Key(账户级)
   * @param {string} [config.withdrawPrivateKeyPem] - 对接方 RSA 私钥 PEM(提币请求签名)
   * @param {string} [config.withdrawPlatformPublicKeyPem] - 平台 RSA 公钥 PEM(可选,回调验签用)
   * @param {number} [config.timeout=30000] - HTTP 超时毫秒
   * @param {string} [config.userAgent] - 自定义 User-Agent
   */
  constructor(config) {
    if (!config || !config.baseUrl) {
      throw new NexCoreError('baseUrl is required');
    }
    this._cfg = config;

    this.http = new Http({
      baseUrl: config.baseUrl,
      timeout: config.timeout || 30000,
      userAgent: config.userAgent || `NexCore-Node-SDK/${Client.VERSION}`,
    });

    this.payment = new Payment(this);
    this.exchange = new Exchange(this);
    this.energy = new Energy(this);
    this.smtp = new Smtp(this);
    this.withdraw = new Withdraw(this);
    this.account = new Account(this);
    this.vcard = new VCard(this);
  }

  /**
   * 取配置字段(各 namespace 内部使用).
   * @param {string} key
   * @returns {*}
   */
  get(key) {
    return this._cfg[key];
  }

  /**
   * 校验平台 webhook 回调签名(复刻后端签名算法,常量时间比较防时序攻击).
   *
   * 算法:取 params 中非空(`'' / null / undefined`)且非 `sign` 字段,
   * 按 key 升序拼成 `k1=v1&k2=v2`,用 secret 做 HMAC-SHA256 hex,
   * 与 `params.sign` 用 `crypto.timingSafeEqual` 比较.
   *
   * 回调体里的 `sign_ts` / `nonce` 字段参与签名,业务方应另行校验时间窗口与
   * nonce 去重以防重放攻击(本方法只验签名正确性,不做重放检查).
   *
   * @example
   *   const ok = Client.verifyWebhook(req.body, apiSecret);
   *   if (!ok) { res.statusCode = 401; return res.end(); }
   *
   * @param {object} params - 回调 JSON 完整解码后的对象(含 sign 字段)
   * @param {string} secret - 验签密钥(MPK apiSecret 等)
   * @returns {boolean} true=签名正确;false=签名错误 / 缺失 / 参数非法
   */
  static verifyWebhook(params, secret) {
    if (!params || typeof params !== 'object' || !params.sign || !secret) return false;
    const filtered = Object.entries(params)
      .filter(([k, v]) => v !== '' && v !== null && v !== undefined && k !== 'sign')
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
    const msg = filtered.map(([k, v]) => `${k}=${v}`).join('&');
    const expected = crypto.createHmac('sha256', secret).update(msg).digest('hex');
    try {
      return crypto.timingSafeEqual(Buffer.from(expected), Buffer.from(String(params.sign)));
    } catch {
      return false;
    }
  }
}

Client.VERSION = '3.2.0';

module.exports = { Client };
