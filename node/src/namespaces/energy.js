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
   * @returns {Promise<object>} {platform_avail_energy, minimum_order_energy, maximum_order_energy, tiered_pricing, ...}
   */
  getInfo() {
    return this._c.http.request('GET', '/api/v1/energy/info', { headers: this._headers() });
  }

  /**
   * 获取指定能量数 + 周期的报价.
   *
   * `GET /api/v1/energy/price?energy=65000&period=1D`
   *
   * @param {number} energy - 能量数
   * @param {string} [period='1D'] - 1H / 6H / 1D / 3D / 1W
   * @returns {Promise<object>}
   */
  getPrice(energy, period = '1D') {
    return this._c.http.request('GET', '/api/v1/energy/price', {
      query: { energy, period },
      headers: this._headers(),
    });
  }

  /**
   * 根据接收地址估算 TRC20 转账所需能量.
   *
   * `GET /api/v1/energy/estimate-energy?receive_addr=TXxxxxxxxx`
   *
   * @param {string} receiveAddr - 收款 TRON 地址(T 开头 Base58)
   * @returns {Promise<object>} {estimated_energy, has_usdt_balance, ...}
   */
  estimateEnergy(receiveAddr) {
    return this._c.http.request('GET', '/api/v1/energy/estimate-energy', {
      query: { receive_addr: receiveAddr },
      headers: this._headers(),
    });
  }

  /**
   * 创建常规租赁订单.
   *
   * `POST /api/v1/energy/order`
   *
   * @param {object} params
   * @param {string} params.receive_addr - 收能量的目标 TRON 地址
   * @param {number} params.energy - 能量数(>= minimum_order_energy)
   * @param {string} params.period - 1H / 6H / 1D / 3D / 1W
   * @param {string} [params.out_serial] - 商户侧订单号(幂等用)
   * @returns {Promise<object>} {serial, status, delegated_at, ...}
   */
  createOrder(params) {
    return this._c.http.request('POST', '/api/v1/energy/order', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 创建一次性订单(用完不续).适用场景:用户只做一笔 TRC20 转账.
   *
   * `POST /api/v1/energy/order/onetime`
   *
   * @param {object} params - 同 createOrder
   * @returns {Promise<object>}
   */
  createOnetimeOrder(params) {
    return this._c.http.request('POST', '/api/v1/energy/order/onetime', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 查询订单状态.
   *
   * `GET /api/v1/energy/order/:serial`
   *
   * @param {string} serial - 订单序列号(string,不是数字 id)
   * @returns {Promise<object>}
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
   * @param {object} [filter] - {status?, page?, page_size?, ...}
   * @returns {Promise<object>}
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
   * @returns {Promise<object>}
   */
  reclaimOrder(serial) {
    return this._c.http.request('POST', '/api/v1/energy/order/reclaim', {
      body: { serial },
      headers: this._headers(),
    });
  }
}

module.exports = { Energy };
