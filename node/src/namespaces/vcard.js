'use strict';

/**
 * 虚拟信用卡命名空间.
 *
 * 对应 /docs 文档 "虚拟信用卡" 模块的 v1 公开接口.
 * 共用 MPK 商户密钥(apiKey / apiSecret,与 account 命名空间一致).
 *
 * 两类鉴权:
 *
 *  ① 双密钥 header(只读查询类):X-API-Key + X-Secret-Key
 *       GET /api/v1/vcard/info                       getInfo                卡服务信息
 *       GET /api/v1/vcard/bins                        listBins               可开卡 BIN 列表
 *       GET /api/v1/vcard/cards                       listCards              我的卡列表
 *       GET /api/v1/vcard/cards/{id}/transactions     getCardTransactions    单卡交易流水
 *       GET /api/v1/vcard/orders                      listOrders             订单列表
 *       GET /api/v1/vcard/orders/{id}                 getOrder               单笔订单
 *       PUT /api/v1/vcard/cards/{id}/remark           updateCardRemark       改卡备注
 *
 *  ② HMAC 头签名(敏感 / 写操作类):X-Key-ID + X-Timestamp + X-Nonce + X-Signature
 *       GET  /api/v1/vcard/cards/{id}/details         getCardDetails         卡敏感详情(卡号等)
 *       GET  /api/v1/vcard/cards/{id}/code            getCardCode            CVV / 安全码
 *       POST /api/v1/vcard/cards                      openCard               开卡
 *       POST /api/v1/vcard/cards/{id}/recharge        rechargeCard           充值
 *       POST /api/v1/vcard/cards/{id}/cancel          cancelCard             销卡(无 body)
 *
 * 签名算法(与后端字节级一致):
 *   ts      = 当前 unix 秒字符串
 *   nonce   = crypto.randomBytes(8).toString('hex')
 *   method  = 大写
 *   path    = 请求路径(含 id,不含 query)
 *   rawQuery= ""(签名接口均无 query)
 *   body    = 实际发送的 JSON 字符串(GET / 无 body 为 "")
 *   payload = ts + nonce + method + path + rawQuery + body
 *   sig     = HMAC-SHA256(apiSecret, payload) 的 hex
 *
 * 关键:POST 先 JSON.stringify 成字符串,对该字符串签名,再把"同一字符串"原样作为
 * 请求 body 发送(http.js 对 string body 不二次序列化),保证签名字节与发送字节一致.
 */

const crypto = require('crypto');
const { NexCoreError } = require('../errors');

class VCard {
  /** @param {import('../client').Client} client */
  constructor(client) { this._c = client; }

  // ---------- ① 双密钥 header(只读查询) ----------

  /** @private 双密钥 header(复用 energy 的模式,密钥用 apiKey/apiSecret) */
  _headers() {
    const k = this._c.get('apiKey');
    const s = this._c.get('apiSecret');
    if (!k || !s) {
      throw new NexCoreError('apiKey / apiSecret not configured');
    }
    return { 'X-API-Key': k, 'X-Secret-Key': s };
  }

  /**
   * 卡服务信息.`GET /api/v1/vcard/info`
   * @returns {Promise<object>}
   */
  getInfo() {
    return this._c.http.request('GET', '/api/v1/vcard/info', { headers: this._headers() });
  }

  /**
   * 可开卡 BIN 列表.`GET /api/v1/vcard/bins`
   * @returns {Promise<object>}
   */
  listBins() {
    return this._c.http.request('GET', '/api/v1/vcard/bins', { headers: this._headers() });
  }

  /**
   * 我的卡列表.`GET /api/v1/vcard/cards`
   * @returns {Promise<object>}
   */
  listCards() {
    return this._c.http.request('GET', '/api/v1/vcard/cards', { headers: this._headers() });
  }

  /**
   * 单卡交易流水.`GET /api/v1/vcard/cards/{id}/transactions`
   * @param {string|number} cardId
   * @returns {Promise<object>}
   */
  getCardTransactions(cardId) {
    if (!cardId && cardId !== 0) throw new NexCoreError('cardId is required');
    return this._c.http.request('GET', `/api/v1/vcard/cards/${cardId}/transactions`, {
      headers: this._headers(),
    });
  }

  /**
   * 订单列表.`GET /api/v1/vcard/orders`
   * @param {object} [query] - {page?, page_size?, status?, order_type?}
   * @returns {Promise<object>}
   */
  listOrders(query = {}) {
    return this._c.http.request('GET', '/api/v1/vcard/orders', {
      query,
      headers: this._headers(),
    });
  }

  /**
   * 单笔订单.`GET /api/v1/vcard/orders/{id}`
   * @param {string|number} orderId
   * @returns {Promise<object>}
   */
  getOrder(orderId) {
    if (!orderId && orderId !== 0) throw new NexCoreError('orderId is required');
    return this._c.http.request('GET', `/api/v1/vcard/orders/${orderId}`, {
      headers: this._headers(),
    });
  }

  /**
   * 修改卡备注.`PUT /api/v1/vcard/cards/{id}/remark`
   * @param {string|number} cardId
   * @param {string} remark
   * @returns {Promise<object>}
   */
  updateCardRemark(cardId, remark) {
    if (!cardId && cardId !== 0) throw new NexCoreError('cardId is required');
    return this._c.http.request('PUT', `/api/v1/vcard/cards/${cardId}/remark`, {
      body: { remark },
      headers: this._headers(),
    });
  }

  // ---------- ② HMAC 头签名(敏感 / 写操作) ----------

  /**
   * 计算 HMAC-SHA256 头签名(hex).
   *
   * 业务方一般无需直接调,_signedRequest 内部自动调用;公开便于测试 / 自实现调试.
   *
   * @param {string} ts - unix 秒字符串
   * @param {string} nonce - 一次性随机串
   * @param {string} method - HTTP 方法
   * @param {string} path - 请求路径(含 id,不含 query)
   * @param {string} rawQuery - query 原文(签名接口为空)
   * @param {string} body - 实际发送的 body 字符串(GET / 无 body 为 "")
   * @returns {string} 64 字符小写 hex 签名
   */
  sign(ts, nonce, method, path, rawQuery, body) {
    const secret = this._c.get('apiSecret');
    if (!secret) throw new NexCoreError('apiSecret not configured');
    const payload = `${ts}${nonce}${String(method).toUpperCase()}${path}${rawQuery}${body}`;
    return crypto.createHmac('sha256', secret).update(payload).digest('hex');
  }

  /**
   * 内部统一发签名请求 — 自动加 4 个鉴权头.
   *
   * @private
   * @param {string} method - 大写 GET / POST
   * @param {string} path - 含 id、不含 query
   * @param {object} [bodyObj] - POST body 对象;无 body(如 cancelCard / GET)传 undefined
   * @returns {Promise<object>}
   */
  _signedRequest(method, path, bodyObj) {
    const keyId = this._c.get('apiKey');
    const secret = this._c.get('apiSecret');
    if (!keyId || !secret) {
      throw new NexCoreError('apiKey / apiSecret not configured');
    }
    const ts = String(Math.floor(Date.now() / 1000));
    const nonce = crypto.randomBytes(8).toString('hex');

    // body:先序列化成字符串,签名与发送共用同一字符串(http.js 对 string 不再 stringify).
    let bodyStr = '';
    let bodyToSend;
    if (bodyObj !== undefined && bodyObj !== null) {
      bodyStr = JSON.stringify(bodyObj);
      bodyToSend = bodyStr;
    }
    const rawQuery = ''; // 签名接口均无 query
    const sig = this.sign(ts, nonce, method, path, rawQuery, bodyStr);

    return this._c.http.request(method, path, {
      body: bodyToSend,
      headers: {
        'X-Key-ID': keyId,
        'X-Timestamp': ts,
        'X-Nonce': nonce,
        'X-Signature': sig,
      },
    });
  }

  /**
   * 卡敏感详情(完整卡号等).`GET /api/v1/vcard/cards/{id}/details`
   * @param {string|number} cardId
   * @returns {Promise<object>}
   */
  getCardDetails(cardId) {
    if (!cardId && cardId !== 0) throw new NexCoreError('cardId is required');
    return this._signedRequest('GET', `/api/v1/vcard/cards/${cardId}/details`);
  }

  /**
   * 卡安全码(CVV).`GET /api/v1/vcard/cards/{id}/code`
   * @param {string|number} cardId
   * @returns {Promise<object>}
   */
  getCardCode(cardId) {
    if (!cardId && cardId !== 0) throw new NexCoreError('cardId is required');
    return this._signedRequest('GET', `/api/v1/vcard/cards/${cardId}/code`);
  }

  /**
   * 开卡.`POST /api/v1/vcard/cards`
   * @param {object} params
   * @param {number} params.bin_platform_id - 卡段 platform_id(listBins 返回,必填)
   * @param {number} params.amount - 开卡充值金额(必填,>0)
   * @returns {Promise<object>} {order_id, status, total_cost}
   */
  openCard(params) {
    return this._signedRequest('POST', '/api/v1/vcard/cards', params || {});
  }

  /**
   * 卡充值.`POST /api/v1/vcard/cards/{id}/recharge`
   * @param {string|number} cardId
   * @param {object} params - 充值参数(amount 等)
   * @returns {Promise<object>}
   */
  rechargeCard(cardId, params) {
    if (!cardId && cardId !== 0) throw new NexCoreError('cardId is required');
    return this._signedRequest('POST', `/api/v1/vcard/cards/${cardId}/recharge`, params || {});
  }

  /**
   * 销卡(无 body).`POST /api/v1/vcard/cards/{id}/cancel`
   * @param {string|number} cardId
   * @returns {Promise<object>}
   */
  cancelCard(cardId) {
    if (!cardId && cardId !== 0) throw new NexCoreError('cardId is required');
    return this._signedRequest('POST', `/api/v1/vcard/cards/${cardId}/cancel`);
  }
}

module.exports = { VCard };
