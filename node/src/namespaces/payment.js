'use strict';

/**
 * 多链收款命名空间.
 *
 * 对应 /docs 文档 "多链收款" 模块的全部 v1 公开接口.
 *
 * 鉴权:HMAC-SHA256 签名 — 所有请求自动追加 app_id + sign 字段.
 * 签名算法:把所有参数按 key 升序拼接成 k1=v1&k2=v2,然后用 app_key 做 HMAC-SHA256.
 *
 * 实现以下 7 个 v1 endpoint(对照 internal/handler/order.go + one_to_one.go):
 *   POST /api/v1/pay/create          createOrder        创建收款订单
 *   GET  /api/v1/pay/query           queryOrder         查询订单状态
 *   POST /api/v1/pay/close           closeOrder         关闭订单
 *   GET  /api/v1/pay/app-config      getAppConfig       查询应用配置
 *   POST /api/v1/pay/bind-address    bindAddress        一对一 — 绑定地址
 *   POST /api/v1/pay/get-address     getUserAddress     一对一 — 查询用户已绑地址
 *   POST /api/v1/pay/unbind-address  unbindAddress      一对一 — 解绑
 *
 * 另提供 verifyNotifySign() 校验 webhook 回调签名(常量时间比较).
 */

const crypto = require('crypto');
const { NexCoreError } = require('../errors');

class Payment {
  /** @param {import('../client').Client} client */
  constructor(client) { this._c = client; }

  /**
   * 计算 HMAC-SHA256 签名.
   *
   * 业务方一般不需要直接调,SDK 内部自动调用.公开出来便于:
   *   - 自行测试签名是否正确(对照 /docs 文档输出)
   *   - 校验回调签名(verifyNotifySign 内部也用)
   *
   * @param {object} params - 待签名参数(会自动过滤 sign 字段和空值,按 key 升序排)
   * @returns {string} 64 字符小写 hex 签名
   * @throws {NexCoreError} paymentAppKey 未配置
   */
  sign(params) {
    const key = this._c.get('paymentAppKey');
    if (!key) throw new NexCoreError('paymentAppKey not configured');
    const filtered = Object.entries(params)
      .filter(([k, v]) => v !== '' && v !== null && v !== undefined && k !== 'sign')
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
    const msg = filtered.map(([k, v]) => `${k}=${v}`).join('&');
    return crypto.createHmac('sha256', key).update(msg).digest('hex');
  }

  /**
   * 自动注入 app_id + 计算 sign,返回签好的参数.
   * @private
   */
  _signed(params) {
    const appId = this._c.get('paymentAppId');
    if (!appId) throw new NexCoreError('paymentAppId not configured');
    const p = { ...params, app_id: appId };
    p.sign = this.sign(p);
    return p;
  }

  // ============ Endpoints ============

  /**
   * 创建收款订单.
   *
   * `POST /api/v1/pay/create`
   *
   * @param {object} params
   * @param {string} params.out_order_id - 商户侧订单号(必须唯一)
   * @param {string|number} params.amount - 法币金额,推荐两位小数 string 避免浮点
   * @param {string} params.currency - 法币 CNY/USD/EUR/JPY/KRW/HKD
   * @param {string} params.trade_type - 加密币种.链,如 "usdt.trc20"
   * @param {string} [params.call_type] - "rotation"(轮播)或 "one_to_one",默认 rotation
   * @param {string} [params.user_id] - 一对一模式必填
   * @param {number} [params.timeout] - 过期秒数,默认 1800
   * @param {string} [params.subject] - 订单描述
   * @param {string} [params.notify_url] - webhook 回调
   * @param {string} [params.return_url] - 成功跳转
   * @returns {Promise<object>} {order_id, pay_address, crypto_amount, crypto_currency, expires_at, ...}
   */
  createOrder(params) {
    return this._c.http.request('POST', '/api/v1/pay/create', { body: this._signed(params) });
  }

  /**
   * 查询订单当前状态.
   *
   * `GET /api/v1/pay/query`
   *
   * @param {string} outOrderId - 商户订单号
   * @returns {Promise<object>}
   */
  queryOrder(outOrderId) {
    return this._c.http.request('GET', '/api/v1/pay/query', { query: this._signed({ out_order_id: outOrderId }) });
  }

  /**
   * 主动关闭订单.
   *
   * `POST /api/v1/pay/close`
   *
   * @param {string} outOrderId
   * @returns {Promise<object>}
   */
  closeOrder(outOrderId) {
    return this._c.http.request('POST', '/api/v1/pay/close', { body: this._signed({ out_order_id: outOrderId }) });
  }

  /**
   * 查询当前应用配置.
   *
   * `GET /api/v1/pay/app-config`
   *
   * @returns {Promise<object>}
   */
  getAppConfig() {
    return this._c.http.request('GET', '/api/v1/pay/app-config', { query: this._signed({}) });
  }

  /**
   * 一对一 — 绑定收款地址.
   *
   * `POST /api/v1/pay/bind-address`
   *
   * @param {string} userId
   * @param {string} tradeType
   * @returns {Promise<object>}
   */
  bindAddress(userId, tradeType) {
    return this._c.http.request('POST', '/api/v1/pay/bind-address', {
      body: this._signed({ user_id: userId, trade_type: tradeType }),
    });
  }

  /**
   * 一对一 — 查询用户已绑定的地址.
   *
   * `POST /api/v1/pay/get-address`(注意:后端是 POST,不是 GET)
   *
   * @param {string} userId
   * @param {string} tradeType
   * @returns {Promise<object>}
   */
  getUserAddress(userId, tradeType) {
    return this._c.http.request('POST', '/api/v1/pay/get-address', {
      body: this._signed({ user_id: userId, trade_type: tradeType }),
    });
  }

  /**
   * 一对一 — 解绑用户地址.
   *
   * `POST /api/v1/pay/unbind-address`
   *
   * @param {string} userId
   * @returns {Promise<object>}
   */
  unbindAddress(userId) {
    return this._c.http.request('POST', '/api/v1/pay/unbind-address', {
      body: this._signed({ user_id: userId }),
    });
  }

  /**
   * 校验 webhook 回调签名(常量时间比较防时序攻击).
   *
   * @param {object} payload - 回调 JSON 完整解码后的对象
   * @returns {boolean} true=签名正确,可信;false=签名错误/缺失
   */
  verifyNotifySign(payload) {
    if (!payload || !payload.sign) return false;
    const expected = this.sign(payload);
    try {
      return crypto.timingSafeEqual(Buffer.from(expected), Buffer.from(payload.sign));
    } catch {
      return false;
    }
  }
}

module.exports = { Payment };
