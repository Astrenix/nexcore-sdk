'use strict';

/**
 * SMTP 聚合 API 命名空间.
 *
 * 对应 /docs 文档 "SMTP API" 模块的全部 v1 公开接口.
 * 鉴权:Bearer Token — `Authorization: Bearer smk_xxx`.
 *
 * 实现 5 个 v1 endpoint(对照 internal/handler/smtp_api.go + smtp_api_ext.go):
 *   POST /api/v1/smtp/send                 send           发送单封邮件
 *   POST /api/v1/smtp/send/batch           sendBatch      批量发送
 *   POST /api/v1/smtp/send/template        sendTemplate   按模板渲染发送
 *   GET  /api/v1/smtp/quota                getQuota       查询本期配额与用量
 *   GET  /api/v1/smtp/status/:message_id   getStatus      查询邮件投递状态
 */

const { NexCoreError } = require('../errors');

class Smtp {
  /** @param {import('../client').Client} client */
  constructor(client) { this._c = client; }

  /** @private */
  _headers() {
    const k = this._c.get('smtpApiKey');
    if (!k) throw new NexCoreError('smtpApiKey not configured');
    return { Authorization: `Bearer ${k}` };
  }

  /**
   * 发送单封邮件.
   *
   * `POST /api/v1/smtp/send`
   *
   * @param {object} params
   * @param {string} params.to - 收件人邮箱
   * @param {string} params.subject - 邮件主题
   * @param {string} params.body - 正文(纯文本或 HTML)
   * @param {boolean} [params.is_html=false] - body 是否为 HTML
   * @param {number} [params.account_id] - 指定发信账户 ID(默认自动选最优)
   * @param {string} [params.reply_to] - 回信地址
   * @returns {Promise<object>} {message_id, status}
   */
  send(params) {
    return this._c.http.request('POST', '/api/v1/smtp/send', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 批量发送(同主题/正文,多收件人).
   *
   * `POST /api/v1/smtp/send/batch`
   *
   * @param {object} params
   * @param {string[]} params.to - 收件人邮箱列表
   * @param {string} params.subject
   * @param {string} params.body
   * @param {boolean} [params.is_html]
   * @param {number} [params.account_id]
   * @returns {Promise<object>} {message_ids, total, accepted}
   */
  sendBatch(params) {
    return this._c.http.request('POST', '/api/v1/smtp/send/batch', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 按模板渲染发送.
   *
   * `POST /api/v1/smtp/send/template`
   *
   * 模板需要先在用户后台 "SMTP API → 模板管理" 创建.
   *
   * @param {object} params
   * @param {string} params.to
   * @param {number} params.template_id - 模板 ID
   * @param {object} params.variables - 渲染变量(对应模板中 `{{var_name}}` 占位符)
   * @param {number} [params.account_id]
   * @returns {Promise<object>}
   */
  sendTemplate(params) {
    return this._c.http.request('POST', '/api/v1/smtp/send/template', {
      body: params,
      headers: this._headers(),
    });
  }

  /**
   * 查询当前订阅期内的配额与已用量.
   *
   * `GET /api/v1/smtp/quota`
   *
   * @returns {Promise<object>} {today_used, today_quota, period_used, period_quota, expires_at}
   */
  getQuota() {
    return this._c.http.request('GET', '/api/v1/smtp/quota', {
      headers: this._headers(),
    });
  }

  /**
   * 查询指定邮件的投递状态.
   *
   * `GET /api/v1/smtp/status/:message_id`
   *
   * @param {string} messageId - send / sendBatch / sendTemplate 返回的 message_id
   * @returns {Promise<object>} {message_id, status, sent_at, opened_at, clicked_at, error_msg, ...}
   */
  getStatus(messageId) {
    return this._c.http.request('GET', `/api/v1/smtp/status/${messageId}`, {
      headers: this._headers(),
    });
  }
}

module.exports = { Smtp };
