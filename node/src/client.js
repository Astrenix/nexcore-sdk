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

const { NexCoreError } = require('./errors');
const { Http } = require('./http');
const { Payment } = require('./namespaces/payment');
const { Exchange } = require('./namespaces/exchange');
const { Energy } = require('./namespaces/energy');
const { Smtp } = require('./namespaces/smtp');

class Client {
  /**
   * @param {object} config
   * @param {string} config.baseUrl - NexCore 平台基础 URL
   * @param {string} [config.paymentAppId] - 多链收款 / 汇率应用 ID
   * @param {string} [config.paymentAppKey] - 多链收款 / 汇率应用密钥
   * @param {string} [config.energyApiKey] - 能量租赁 X-API-Key
   * @param {string} [config.energySecretKey] - 能量租赁 X-Secret-Key
   * @param {string} [config.smtpApiKey] - SMTP smk_ 前缀 Token
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
  }

  /**
   * 取配置字段(各 namespace 内部使用).
   * @param {string} key
   * @returns {*}
   */
  get(key) {
    return this._cfg[key];
  }
}

Client.VERSION = '3.0.0';

module.exports = { Client };
