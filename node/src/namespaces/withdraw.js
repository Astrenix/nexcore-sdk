'use strict';

/**
 * 提币 API namespace — 多链收款业务的资金出库端.
 *
 * 鉴权:RSA-PKCS1v15-SHA256 签名 + 4 个请求头
 *
 *   X-API-Key            账户级 API Key(控制台「账号 → API 密钥」)
 *   X-Timestamp          unix ms,与服务器时差 ≤ 60s
 *   X-Nonce              一次性 nonce(uuid v4),5 分钟内不可重复
 *   X-Withdraw-Signature RSA-PKCS1v15-SHA256(caller_private_key, signString),Base64
 *
 * signString = METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + BODY
 * 其中 BODY 为 HTTP body 原文(JSON 字符串原样,GET 请求为空字符串).
 *
 * 对应 /docs 文档 "提币 API" 章节的 4 个 endpoint:
 *
 *   POST /api/v1/withdraw                 createWithdraw            发起提币
 *   GET  /api/v1/withdraw/:id             getWithdraw               查询单笔状态
 *   GET  /api/v1/balance/withdrawable     getWithdrawableBalance    查询可提余额
 *   GET  /api/v1/fee/quote                quoteFee                  费用预估
 *
 * 另提供 verifyCallback() 校验平台回调签名(用平台公钥).
 */

const crypto = require('crypto');
const { NexCoreError } = require('../errors');

class Withdraw {
  constructor(client) {
    this._c = client;
    this._privKeyObj = null;       // cached crypto.KeyObject
    this._platformPubObj = null;
  }

  // ---------- 内部:密钥懒解析 ----------

  _getPrivKey() {
    if (this._privKeyObj) return this._privKeyObj;
    const pem = this._c.get('withdrawPrivateKeyPem');
    if (!pem) throw new NexCoreError('withdrawPrivateKeyPem not configured');
    try {
      this._privKeyObj = crypto.createPrivateKey(pem);
    } catch (e) {
      throw new NexCoreError(`withdraw: invalid private key PEM: ${e.message}`);
    }
    if (this._privKeyObj.asymmetricKeyType !== 'rsa') {
      throw new NexCoreError('withdraw: configured private key is not RSA');
    }
    return this._privKeyObj;
  }

  _getPlatformPub() {
    if (this._platformPubObj) return this._platformPubObj;
    const pem = this._c.get('withdrawPlatformPublicKeyPem');
    if (!pem) throw new NexCoreError('withdrawPlatformPublicKeyPem not configured');
    try {
      this._platformPubObj = crypto.createPublicKey(pem);
    } catch (e) {
      throw new NexCoreError(`withdraw: invalid platform public key PEM: ${e.message}`);
    }
    if (this._platformPubObj.asymmetricKeyType !== 'rsa') {
      throw new NexCoreError('withdraw: platform key is not RSA');
    }
    return this._platformPubObj;
  }

  // ---------- 签名 ----------

  /**
   * 计算请求的 RSA-PKCS1v15-SHA256 签名(Base64).
   *
   * 业务方一般不需要直接调,SDK 内部 _do 时自动调用.
   * 公开出来便于测试 / 自实现非标场景(比如 curl 调试).
   *
   * @param {string} method
   * @param {string} path
   * @param {string} timestamp
   * @param {string} nonce
   * @param {string} body - HTTP body 原文(GET 请求传空字符串)
   * @returns {string} base64 sig
   */
  sign(method, path, timestamp, nonce, body) {
    const signString = `${String(method).toUpperCase()}\n${path}\n${timestamp}\n${nonce}\n${body}`;
    const priv = this._getPrivKey();
    const signer = crypto.createSign('RSA-SHA256');
    signer.update(signString, 'utf8');
    signer.end();
    return signer.sign(priv).toString('base64');
  }

  /**
   * 内部统一发请求 — 自动加 4 个鉴权头.
   *
   * @param {string} method
   * @param {string} path
   * @param {object} [body] - 业务参数对象(会 JSON.stringify;一次产出,签名和发送都用它)
   * @param {object} [query]
   * @returns {Promise<object>}
   */
  async _do(method, path, body, query) {
    const apiKey = this._c.get('withdrawApiKey');
    if (!apiKey) throw new NexCoreError('withdrawApiKey not configured');
    const timestamp = String(Date.now());
    const nonce = (typeof crypto.randomUUID === 'function')
      ? crypto.randomUUID()
      : _legacyUUID();
    let bodyStr = '';
    let bodyToSend;
    if (body !== undefined && body !== null) {
      bodyStr = JSON.stringify(body);
      bodyToSend = bodyStr; // http.js 接受 string,会原样发送
    }
    const sig = this.sign(method, path, timestamp, nonce, bodyStr);
    return this._c.http.request(method, path, {
      body: bodyToSend,
      query,
      headers: {
        'X-API-Key': apiKey,
        'X-Timestamp': timestamp,
        'X-Nonce': nonce,
        'X-Withdraw-Signature': sig,
      },
    });
  }

  // ---------- 公开 endpoint ----------

  /**
   * 发起提币 — POST /api/v1/withdraw
   *
   * 下单后状态为 pending,等延迟到期由 worker 自动广播.
   * 期间可在控制台暂停 / 加速 / 取消.
   *
   * @param {object} params - 必填 chain/symbol/amount/to_address;可选 memo/callback_url/request_id
   * @returns {Promise<object>} {order_id, status, amount, fee, fee_mode, delayed_until}
   */
  createWithdraw(params) {
    return this._do('POST', '/api/v1/withdraw', params);
  }

  /**
   * 查询单笔提币状态 — GET /api/v1/withdraw/:id
   * @param {string} orderId
   * @returns {Promise<object>}
   */
  getWithdraw(orderId) {
    if (!orderId) throw new NexCoreError('orderId is required');
    return this._do('GET', `/api/v1/withdraw/${orderId}`);
  }

  /**
   * 查询可提余额 — GET /api/v1/balance/withdrawable
   *
   * 返回该账户在每条链 × 每种资产下的「已归集待提现」余额.
   * 只有这部分可用于 API 提币.
   *
   * @returns {Promise<object>} {tron: {USDT: "...", ...}, eth: {...}, ...}
   */
  getWithdrawableBalance() {
    return this._do('GET', '/api/v1/balance/withdrawable');
  }

  /**
   * 费用预估 — GET /api/v1/fee/quote
   *
   * @param {string} chain - tron / eth / bsc / polygon / arbitrum / btc
   * @param {string} symbol - USDT / TRX / ETH 等
   * @param {string} [amount] - 提币金额(可选)
   * @returns {Promise<object>} {chain, symbol, amount, fee_amount, fee_asset}
   */
  quoteFee(chain, symbol, amount) {
    if (!chain || !symbol) throw new NexCoreError('chain and symbol are required');
    const q = { chain, symbol };
    if (amount) q.amount = amount;
    return this._do('GET', '/api/v1/fee/quote', undefined, q);
  }

  // ---------- 回调验签 ----------

  /**
   * 验证平台回调签名(对接方收到 webhook 时调用).
   *
   * 用法:
   *   const sig = req.headers['x-platform-signature'];
   *   const ts  = req.headers['x-timestamp'];
   *   const nonce = req.headers['x-nonce'];
   *   const body = rawBodyBuffer.toString('utf8'); // 必须是 raw body 字符串,不要 re-stringify
   *   try {
   *     client.withdraw.verifyCallback(req.method, req.url, ts, nonce, body, sig);
   *   } catch (e) {
   *     res.statusCode = 401; res.end();
   *   }
   *
   * 验签算法与请求方向一致:RSA-PKCS1v15-SHA256(platform_public_key, signString).
   *
   * @param {string} method
   * @param {string} path
   * @param {string} timestamp
   * @param {string} nonce
   * @param {string} body - HTTP body 原文
   * @param {string} base64Signature
   * @throws {NexCoreError} 验签失败
   */
  verifyCallback(method, path, timestamp, nonce, body, base64Signature) {
    const pub = this._getPlatformPub();
    const signString = `${String(method).toUpperCase()}\n${path}\n${timestamp}\n${nonce}\n${body}`;
    const verifier = crypto.createVerify('RSA-SHA256');
    verifier.update(signString, 'utf8');
    verifier.end();
    let ok = false;
    try {
      ok = verifier.verify(pub, base64Signature, 'base64');
    } catch (e) {
      throw new NexCoreError(`withdraw: signature verify error: ${e.message}`);
    }
    if (!ok) throw new NexCoreError('withdraw: signature verify failed');
  }
}

// 老 Node(< 14.17)无 crypto.randomUUID 时的后备实现
function _legacyUUID() {
  const b = crypto.randomBytes(16);
  b[6] = (b[6] & 0x0f) | 0x40;
  b[8] = (b[8] & 0x3f) | 0x80;
  const hex = b.toString('hex');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

module.exports = { Withdraw };
