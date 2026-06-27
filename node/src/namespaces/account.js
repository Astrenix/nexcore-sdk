'use strict';

/**
 * 账户命名空间.
 *
 * 对应 /docs 文档 "账户" 模块的 v1 公开接口.
 * 鉴权:X-API-Key + X-Secret-Key 双 header(MPK 商户密钥 — apiKey/apiSecret,
 * account 与 vcard 共用).
 *
 * 实现以下 2 个 v1 endpoint:
 *   GET /api/v1/account/balance          getBalance          查询账户余额
 *   GET /api/v1/account/deposit-address  getDepositAddress   查询充值地址
 */

const { NexCoreError } = require('../errors');

class Account {
  /** @param {import('../client').Client} client */
  constructor(client) { this._c = client; }

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
   * 查询账户余额.
   *
   * `GET /api/v1/account/balance`
   *
   * @returns {Promise<object>}
   */
  getBalance() {
    return this._c.http.request('GET', '/api/v1/account/balance', { headers: this._headers() });
  }

  /**
   * 查询账户充值地址.
   *
   * `GET /api/v1/account/deposit-address`
   *
   * @returns {Promise<object>}
   */
  getDepositAddress() {
    return this._c.http.request('GET', '/api/v1/account/deposit-address', { headers: this._headers() });
  }
}

module.exports = { Account };
