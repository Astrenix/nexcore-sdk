'use strict';

/**
 * TRON 能量租赁命名空间.
 *
 * 对应 /docs 文档 "能量租赁" 模块的全部 v1 公开接口.
 * 鉴权:X-API-Key + X-Secret-Key 双 header.
 *
 * 实现 8 个 v1 endpoint(对照 internal/handler/trxx_api.go):
 *   GET  /api/v1/energy/info             getInfo              平台信息
 *   GET  /api/v1/energy/price            getPrice             报价
 *   GET  /api/v1/energy/estimate-energy  estimateEnergy       估算所需能量
 *   POST /api/v1/energy/order            createOrder          创建常规订单
 *   POST /api/v1/energy/order/onetime    createOnetimeOrder   一次性订单
 *   GET  /api/v1/energy/order/:serial    queryOrder           查询(serial 字符串)
 *   GET  /api/v1/energy/orders           listOrders           订单列表
 *   POST /api/v1/energy/order/reclaim    reclaimOrder         主动回收
 */

const { NexCoreError } = require('../errors');

class Energy {
  /** @param {import('../client').Client} client */
  constructor(client) { this._c = client; }

  /** @private */
  _headers() {
    const k = this._c.get('energyApiKey');
    const s = this._c.get('energySecretKey');
    if (!k || !s) {
      throw new NexCoreError('energyApiKey / energySecretKey not configured');
    }
    return { 'X-API-Key': k, 'X-Secret-Key': s };
  }

  /**
   * 平台公开信息.
   *
   * `GET /api/v1/energy/info`
   *
   * @returns {Promise<object>} {platform_avail_energy, minimum_order_energy, maximum_order_energy,
   *   tiered_pricing: [{period, price}], min_energy_amount, default_energy, default_period}
   */
  getInfo() {
    return this._c.http.request('GET', '/api/v1/energy/info', { headers: this._headers() });
  }

  /**
   * 获取指定能量数 + 周期的报价.
   *
   * `GET /api/v1/energy/price?period=1D&energy_amount=65000`
   *
   * @param {number} energyAmount - 能量数
   * @param {string} [period='1D'] - 1H / 1D / 3D / 7D / 30D
   * @returns {Promise<object>} {period, energy_amount, price_trx}(price_trx 为终价,已含 API 加价)
   */
  getPrice(energyAmount, period = '1D') {
    return this._c.http.request('GET', '/api/v1/energy/price', {
      query: { period, energy_amount: energyAmount },
      headers: this._headers(),
    });
  }

  /**
   * 根据目标地址估算 TRC20 转账所需能量.
   *
   * `GET /api/v1/energy/estimate-energy?to_address=TXxxxxxxxx`
   *
   * @param {string} toAddress - 目标 TRON 地址(T 开头,34 位)
   * @returns {Promise<object>} {to_address, initialized, suggested_energy}
   *   initialized=false 表示地址未持有 USDT,首笔转账消耗更多能量
   */
  estimateEnergy(toAddress) {
    return this._c.http.request('GET', '/api/v1/energy/estimate-energy', {
      query: { to_address: toAddress },
      headers: this._headers(),
    });
  }

  /**
   * 创建常规租赁订单.
   *
   * `POST /api/v1/energy/order`
   *
   * @param {object} params
   * @param {string} params.receive_address - 接收能量的 TRON 地址
   * @param {number} params.energy_amount - 能量数(>= minimum_order_energy)
   * @param {string} params.period - 1H / 1D / 3D / 7D / 30D
   * @param {string} [params.out_trade_no] - 商户侧自定义订单号
   * @param {string} [params.remark] - 备注
   * @returns {Promise<object>} {serial, price_trx, deducted_usd}
   */
  createOrder(params) {
    return this._c.http.request('POST', '/api/v1/energy/order', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 单笔能量下单(笔数策略,系统按策略自动分配能量数).
   *
   * `POST /api/v1/energy/order/onetime`
   *
   * @param {object} params
   * @param {string} params.receive_address - 接收能量的 TRON 地址
   * @param {string} params.period - 1H / 1D / 3D / 7D / 30D
   * @param {string} [params.out_trade_no] - 商户侧自定义订单号
   * @param {string} [params.remark] - 备注
   * @returns {Promise<object>} {serial, price_trx, deducted_usd}(price_trx 按上游实际结算,多退少不补)
   */
  createOnetimeOrder(params) {
    return this._c.http.request('POST', '/api/v1/energy/order/onetime', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 查询订单状态(会先向上游同步一次最新状态).
   *
   * `GET /api/v1/energy/order/:serial`
   *
   * @param {string} serial - 订单序列号(string,不是数字 id)
   * @returns {Promise<object>} {serial, receive_address, energy_amount, period, price_trx,
   *   status, status_msg, out_trade_no, order_type, created_at}
   *   status:0=待处理/处理中,40=成功,41=失败
   */
  queryOrder(serial) {
    return this._c.http.request('GET', `/api/v1/energy/order/${serial}`, {
      headers: this._headers(),
    });
  }

  /**
   * 列出所有订单(可按状态过滤).
   *
   * `GET /api/v1/energy/orders`
   *
   * @param {object} [filter]
   * @param {number} [filter.page=1]
   * @param {number} [filter.page_size=20] - 上限 100
   * @param {number} [filter.status=-1] - -1=全部,0=待处理/处理中,40=成功,41=失败
   * @returns {Promise<object>} {list, total, page, page_size}
   */
  listOrders(filter = {}) {
    return this._c.http.request('GET', '/api/v1/energy/orders', {
      query: filter,
      headers: this._headers(),
    });
  }

  /**
   * 主动回收订单.
   *
   * `POST /api/v1/energy/order/reclaim`
   *
   * @param {string} serial - 订单序列号
   * @returns {Promise<object>} {errno, message}(errno=0 回收成功)
   */
  reclaimOrder(serial) {
    return this._c.http.request('POST', '/api/v1/energy/order/reclaim', {
      body: { serial },
      headers: this._headers(),
    });
  }
}

module.exports = { Energy };
