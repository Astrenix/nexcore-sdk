'use strict';

/**
 * 汇率服务命名空间.
 *
 * 对应 /docs 文档 "多链收款 → 汇率服务" 5 个 endpoint.
 * 走 APIAuth 中间件,用 X-App-Key + X-App-Secret(应用密钥)进行 header 鉴权.
 *
 * 实现 5 个 v1 endpoint(对照 internal/handler/exchange_api.go):
 *   GET  /api/v1/rate          getRate        单对币种汇率
 *   POST /api/v1/convert       convert        金额换算
 *   GET  /api/v1/rates         getRates       批量获取多币种汇率
 *   GET  /api/v1/rates/fiat    getFiatRates   主流法币汇率
 *   GET  /api/v1/rates/all     getAllRates    所有支持币种快照
 */

const { NexCoreError } = require('../errors');

class Exchange {
  /** @param {import('../client').Client} client */
  constructor(client) { this._c = client; }

  /** @private */
  _headers() {
    const appId = this._c.get('paymentAppId');
    const appKey = this._c.get('paymentAppKey');
    if (!appId || !appKey) {
      throw new NexCoreError('paymentAppId / paymentAppKey not configured');
    }
    return { 'X-App-Key': appId, 'X-App-Secret': appKey };
  }

  /**
   * 查询单对币种汇率.
   *
   * `GET /api/v1/rate?from=USDT&to=CNY`
   *
   * @param {string} from - 来源币种(USDT / TRX / ETH / BTC / USD / CNY ...)
   * @param {string} to - 目标币种
   * @returns {Promise<object>} {from, to, rate, inverse, updated_at}
   */
  getRate(from, to) {
    return this._c.http.request('GET', '/api/v1/rate', {
      query: { from, to },
      headers: this._headers(),
    });
  }

  /**
   * 金额换算.
   *
   * `POST /api/v1/convert`
   *
   * @param {string} from
   * @param {string} to
   * @param {string|number} amount - 待换算金额
   * @returns {Promise<object>} {from, to, amount, result, rate, updated_at}(result 为换算结果)
   */
  convert(from, to, amount) {
    return this._c.http.request('POST', '/api/v1/convert', {
      body: { from, to, amount },
      headers: this._headers(),
    });
  }

  /**
   * 批量获取多币种到指定基准币的汇率.
   *
   * `GET /api/v1/rates?symbols=USDT,TRX,ETH&base=USDT`
   *
   * @param {string[]} symbols - 待查询币种代码列表
   * @param {string} [base] - 基准币;不传由后端取默认(USDT)
   * @returns {Promise<object>} {base, rates: {USDT: 7.23, ...}, updated_at}
   */
  getRates(symbols, base) {
    const query = { symbols: symbols.join(',') };
    if (base) query.base = base;
    return this._c.http.request('GET', '/api/v1/rates', {
      query,
      headers: this._headers(),
    });
  }

  /**
   * 主流法币到指定基准法币的汇率.
   *
   * `GET /api/v1/rates/fiat?base=USD`
   *
   * @param {string} [base='USD']
   * @returns {Promise<object>}
   */
  getFiatRates(base = 'USD') {
    return this._c.http.request('GET', '/api/v1/rates/fiat', {
      query: { base },
      headers: this._headers(),
    });
  }

  /**
   * 所有支持币种的汇率快照(加密币 + 法币).
   *
   * `GET /api/v1/rates/all?base=USDT`
   *
   * @param {string} [base='USDT']
   * @returns {Promise<object>}
   */
  getAllRates(base = 'USDT') {
    return this._c.http.request('GET', '/api/v1/rates/all', {
      query: { base },
      headers: this._headers(),
    });
  }
}

module.exports = { Exchange };
