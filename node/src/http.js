'use strict';

/**
 * NexCore SDK 底层 HTTP 传输层.
 *
 * 仅依赖 Node.js 内置 http/https 模块,零外部运行时依赖.
 * 业务命名空间(payment / exchange / energy / smtp)调用本类发请求,
 * 不直接接触 http,保持各 namespace 关注业务逻辑.
 */

const http = require('http');
const https = require('https');
const { URL } = require('url');

const { NexCoreError } = require('./errors');

class Http {
  /**
   * @param {object} opts
   * @param {string} opts.baseUrl
   * @param {number} [opts.timeout=30000]
   * @param {string} [opts.userAgent]
   */
  constructor({ baseUrl, timeout = 30000, userAgent = 'NexCore-Node-SDK/3.0.0' }) {
    if (!baseUrl) throw new NexCoreError('baseUrl is required');
    this.baseUrl = String(baseUrl).replace(/\/+$/, '');
    this.timeout = timeout;
    this.userAgent = userAgent;
  }

  /**
   * 发送 HTTP 请求.
   *
   * @param {string} method - HTTP 方法 GET / POST / PUT / DELETE
   * @param {string} path - 以 / 开头的路径,如 "/api/v1/pay/create"
   * @param {object} [opts]
   * @param {object|string} [opts.body] - JSON body(自动序列化)
   * @param {Object<string,string|number>} [opts.query] - query 参数
   * @param {Object<string,string>} [opts.headers] - 额外 header
   * @returns {Promise<object>} 响应中的 data 段(已解 {code,message,data} envelope)
   * @throws {NexCoreError}
   */
  request(method, path, opts = {}) {
    const { body, query, headers } = opts;
    const url = new URL(this.baseUrl + path);

    if (query && typeof query === 'object') {
      for (const [k, v] of Object.entries(query)) {
        if (v !== undefined && v !== null && v !== '') {
          url.searchParams.set(k, String(v));
        }
      }
    }

    const h = {
      Accept: 'application/json',
      'User-Agent': this.userAgent,
      ...(headers || {}),
    };

    let payload;
    if (body !== undefined && body !== null) {
      payload = typeof body === 'string' ? body : JSON.stringify(body);
      h['Content-Type'] = 'application/json';
      h['Content-Length'] = Buffer.byteLength(payload);
    }

    const lib = url.protocol === 'https:' ? https : http;
    const reqOpts = {
      method: method.toUpperCase(),
      hostname: url.hostname,
      port: url.port || (url.protocol === 'https:' ? 443 : 80),
      path: url.pathname + url.search,
      headers: h,
      timeout: this.timeout,
    };

    return new Promise((resolve, reject) => {
      const req = lib.request(reqOpts, (res) => {
        const chunks = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => {
          const text = Buffer.concat(chunks).toString('utf8');
          const requestId = res.headers['x-trace-id'] || null;
          let json;
          try {
            json = JSON.parse(text);
          } catch {
            return reject(new NexCoreError(
              `HTTP ${res.statusCode}: ${text.slice(0, 200)}`,
              -1, requestId, res.statusCode,
            ));
          }
          if (res.statusCode >= 400 || (json.code !== undefined && json.code !== 0)) {
            return reject(new NexCoreError(
              json.message || `HTTP ${res.statusCode}`,
              json.code ?? -1, requestId, res.statusCode,
            ));
          }
          resolve(json.data !== undefined ? json.data : json);
        });
      });
      req.on('timeout', () => {
        req.destroy();
        reject(new NexCoreError('request timeout', -1));
      });
      req.on('error', (e) => reject(new NexCoreError(`HTTP request failed: ${e.message}`, -1)));
      if (payload) req.write(payload);
      req.end();
    });
  }
}

module.exports = { Http };
