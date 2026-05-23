/**
 * NexCore Official Node.js SDK
 *
 * 全能客户端,一次配置覆盖 Payment / Energy / SMTP / AI 全部业务。
 *
 * 用法:
 *   const { Client, NexCoreError } = require('./index');
 *
 *   const client = new Client({
 *     baseUrl: 'https://your-domain.com',
 *     paymentAppId: 'APP20260412XXXX',
 *     paymentAppKey: 'your_app_key_here',
 *     energyApiKey: 'energy_api_key_here',
 *     energySecretKey: 'energy_secret_key_here',
 *     aiApiKey: 'sk-nc-xxx',
 *   });
 *
 *   const order = await client.payment.createOrder({
 *     out_order_id: `ORDER_${Date.now()}`,
 *     amount: '100.00',
 *     currency: 'CNY',
 *     trade_type: 'usdt.trc20',
 *     call_type: 'rotation',
 *   });
 *
 *   const reply = await client.ai.chat(
 *     [{ role: 'user', content: 'Hello' }],
 *     'claude-opus-4-7'
 *   );
 */

'use strict';

const crypto = require('crypto');
const https = require('https');
const http = require('http');
const { URL } = require('url');

class NexCoreError extends Error {
  constructor(message, code = -1, requestId = null, httpStatus = null) {
    super(message);
    this.name = 'NexCoreError';
    this.code = code;
    this.requestId = requestId;
    this.httpStatus = httpStatus;
  }
}

class Client {
  /**
   * @param {object} config
   * @param {string} config.baseUrl
   * @param {string} [config.paymentAppId]
   * @param {string} [config.paymentAppKey]
   * @param {string} [config.energyApiKey]
   * @param {string} [config.energySecretKey]
   * @param {string} [config.smtpApiKey]
   * @param {string} [config.aiApiKey]
   * @param {number} [config.timeout=30000]
   */
  constructor(config) {
    if (!config || !config.baseUrl) {
      throw new NexCoreError('baseUrl is required');
    }
    this.baseUrl = String(config.baseUrl).replace(/\/+$/, '');
    this._cfg = config;
    this._timeout = config.timeout || 30000;

    this.payment = new PaymentNamespace(this);
    this.energy = new EnergyNamespace(this);
    this.smtp = new SmtpNamespace(this);
    this.ai = new AiNamespace(this);
  }

  async _request(method, path, { body, query, headers } = {}) {
    const url = new URL(this.baseUrl + path);
    if (query && typeof query === 'object') {
      for (const [k, v] of Object.entries(query)) {
        if (v !== undefined && v !== null && v !== '') url.searchParams.set(k, v);
      }
    }

    const h = { Accept: 'application/json', ...(headers || {}) };
    let payload;
    if (body !== undefined && body !== null) {
      payload = typeof body === 'string' ? body : JSON.stringify(body);
      h['Content-Type'] = 'application/json';
      h['Content-Length'] = Buffer.byteLength(payload);
    }

    const lib = url.protocol === 'https:' ? https : http;
    const opts = {
      method: method.toUpperCase(),
      hostname: url.hostname,
      port: url.port || (url.protocol === 'https:' ? 443 : 80),
      path: url.pathname + url.search,
      headers: h,
      timeout: this._timeout,
    };

    return new Promise((resolve, reject) => {
      const req = lib.request(opts, (res) => {
        const chunks = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => {
          const text = Buffer.concat(chunks).toString('utf8');
          const requestId = res.headers['x-trace-id'] || null;
          let json;
          try { json = JSON.parse(text); }
          catch {
            return reject(new NexCoreError(`HTTP ${res.statusCode}: ${text.slice(0, 200)}`, -1, requestId, res.statusCode));
          }
          if (res.statusCode >= 400 || (json.code !== undefined && json.code !== 0)) {
            return reject(new NexCoreError(json.message || `HTTP ${res.statusCode}`, json.code ?? -1, requestId, res.statusCode));
          }
          resolve(json.data !== undefined ? json.data : json);
        });
      });
      req.on('timeout', () => { req.destroy(); reject(new NexCoreError('request timeout', -1)); });
      req.on('error', (e) => reject(new NexCoreError(`HTTP request failed: ${e.message}`, -1)));
      if (payload) req.write(payload);
      req.end();
    });
  }

  _need(key) {
    if (!this._cfg[key]) throw new NexCoreError(`${key} not configured`);
    return this._cfg[key];
  }
}

/**
 * 链收款 — HMAC-SHA256 签名,?sign= 模式.
 */
class PaymentNamespace {
  constructor(client) { this._c = client; }

  _sign(params) {
    const key = this._c._need('paymentAppKey');
    const filtered = Object.entries(params)
      .filter(([k, v]) => v !== '' && v !== null && v !== undefined && k !== 'sign')
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
    const msg = filtered.map(([k, v]) => `${k}=${v}`).join('&');
    return crypto.createHmac('sha256', key).update(msg).digest('hex');
  }

  _signed(params) {
    const appId = this._c._need('paymentAppId');
    const p = { ...params, app_id: appId };
    p.sign = this._sign(p);
    return p;
  }

  createOrder(params) {
    return this._c._request('POST', '/api/v1/pay/create', { body: this._signed(params) });
  }
  queryOrder(outOrderId) {
    return this._c._request('GET', '/api/v1/pay/query', { query: this._signed({ out_order_id: outOrderId }) });
  }
  closeOrder(outOrderId) {
    return this._c._request('POST', '/api/v1/pay/close', { body: this._signed({ out_order_id: outOrderId }) });
  }
  bindAddress(userId, tradeType) {
    return this._c._request('POST', '/api/v1/pay/bind-address', { body: this._signed({ user_id: userId, trade_type: tradeType }) });
  }
  getAddress(userId, tradeType) {
    return this._c._request('GET', '/api/v1/pay/get-address', { query: this._signed({ user_id: userId, trade_type: tradeType }) });
  }
  unbindAddress(userId) {
    return this._c._request('POST', '/api/v1/pay/unbind-address', { body: this._signed({ user_id: userId }) });
  }
  appConfig() {
    return this._c._request('GET', '/api/v1/pay/app-config', { query: this._signed({}) });
  }

  /** 校验 webhook 回调签名 */
  verifyNotifySign(payload) {
    if (!payload || !payload.sign) return false;
    const expected = this._sign(payload);
    try {
      return crypto.timingSafeEqual(Buffer.from(expected), Buffer.from(payload.sign));
    } catch { return false; }
  }
}

/**
 * 能量租赁 — X-API-Key + X-Secret-Key 双 header.
 */
class EnergyNamespace {
  constructor(client) { this._c = client; }
  _headers() {
    return {
      'X-API-Key': this._c._need('energyApiKey'),
      'X-Secret-Key': this._c._need('energySecretKey'),
    };
  }
  info()         { return this._c._request('GET', '/api/v1/energy/info', { headers: this._headers() }); }
  price(energy, period = '1D') {
    return this._c._request('GET', '/api/v1/energy/price', {
      query: { energy, period }, headers: this._headers(),
    });
  }
  estimateEnergy(receiveAddr) {
    return this._c._request('GET', '/api/v1/energy/estimate-energy', {
      query: { receive_addr: receiveAddr }, headers: this._headers(),
    });
  }
  createOrder(params) {
    return this._c._request('POST', '/api/v1/energy/order', { body: params, headers: this._headers() });
  }
  queryOrder(orderId) {
    return this._c._request('GET', `/api/v1/energy/order/${orderId}`, { headers: this._headers() });
  }
  listOrders(filter = {}) {
    return this._c._request('GET', '/api/v1/energy/orders', { query: filter, headers: this._headers() });
  }
}

/**
 * SMTP 聚合 API — X-API-Key header.
 */
class SmtpNamespace {
  constructor(client) { this._c = client; }
  _headers() { return { 'X-API-Key': this._c._need('smtpApiKey') }; }
  sendMail(params)      { return this._c._request('POST', '/api/v1/smtp/send', { body: params, headers: this._headers() }); }
  listAccounts()        { return this._c._request('GET', '/api/v1/smtp/accounts', { headers: this._headers() }); }
  listTemplates()       { return this._c._request('GET', '/api/v1/smtp/templates', { headers: this._headers() }); }
}

/**
 * Astrenix AI(OpenAI 兼容协议).
 */
class AiNamespace {
  constructor(client) { this._c = client; }
  _headers() { return { Authorization: `Bearer ${this._c._need('aiApiKey')}` }; }
  chat(messages, model, extra = {}) {
    return this._c._request('POST', '/v1/chat/completions', {
      body: { model, messages, ...extra }, headers: this._headers(),
    });
  }
  models() { return this._c._request('GET', '/v1/models', { headers: this._headers() }); }
}

module.exports = { Client, NexCoreError };
