'use strict';

/**
 * SMTP 聚合 API 命名空间.
 *
 * 对应 /docs 文档 "SMTP API" 模块的全部 v1 公开接口.
 * 鉴权:Bearer Token — `Authorization: Bearer smk_xxx`.
 *
 * 实现 6 个 v1 endpoint(对照 internal/handler/smtp_api.go + smtp_api_ext.go):
 *   POST /api/v1/smtp/send                 send           发送单封邮件(支持定时 send_at + Idempotency-Key 幂等)
 *   POST /api/v1/smtp/send/batch           sendBatch      批量发送(recipients 逐人变量渲染,支持幂等)
 *   POST /api/v1/smtp/send/template        sendTemplate   按模板 code 渲染发送
 *   GET  /api/v1/smtp/quota                getQuota       查询日/月配额与用量
 *   GET  /api/v1/smtp/status/:message_id   getStatus      查询邮件投递状态
 *   POST /api/v1/smtp/inbound              reportInbound  上报退信/投诉(自动加入抑制名单)
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
   * @param {string} [params.from_name] - 发件人显示名
   * @param {string} [params.reply_to] - 回信地址(Reply-To 头)
   * @param {string} [params.text_body] - 纯文本正文;与 HTML 同时提供时以 multipart/alternative 发送
   * @param {Object<string,string>} [params.headers] - 自定义邮件头(核心头不可覆盖)
   * @param {string[]} [params.cc] - 抄送列表
   * @param {string[]} [params.bcc] - 密送列表(只投递不写头)
   * @param {Array<{filename:string, content_base64:string, content_type:string}>} [params.attachments] - 附件
   * @param {number} [params.account_id] - 指定发信账户 ID(默认自动选号;指定后不故障转移)
   * @param {string} [params.send_at] - 定时发送(RFC3339);晚于当前 30s 以上则排期
   * @param {object} [opts]
   * @param {string} [opts.idempotencyKey] - Idempotency-Key 头;同 key 重试直接返回首次结果,不重复发送/扣配额
   * @returns {Promise<object>} 立即发送:{message_id, status, account_name, used_smtp, account_id, send_duration_ms};
   *   定时分支:{scheduled: true, scheduled_id, send_at}
   */
  send(params, opts = {}) {
    const headers = this._headers();
    if (opts.idempotencyKey) headers['Idempotency-Key'] = opts.idempotencyKey;
    return this._c.http.request('POST', '/api/v1/smtp/send', {
      body: params,
      headers,
    });
  }

  /**
   * 批量发送(recipients 逐收件人独立变量渲染;逐封独立扣配额).
   *
   * `POST /api/v1/smtp/send/batch`
   *
   * 静态模式传 subject + body(支持 `{{var}}` 占位),模板模式传 template_code;二者至少其一.
   * 单次收件人上限 = 订阅的 max_batch_size(默认 10);模板模式需套餐支持模板功能.
   *
   * @param {object} params
   * @param {Array<{to:string, variables?:Object<string,string>, from_name?:string}>} params.recipients - 收件人列表(必填)
   * @param {string} [params.subject] - 静态模式主题
   * @param {string} [params.body] - 静态模式正文,支持 `{{var}}` 替换
   * @param {string} [params.template_code] - 模板模式:模板 code(与 body 二选一)
   * @param {boolean} [params.is_html]
   * @param {string} [params.reply_to]
   * @param {string[]} [params.cc] - 抄送(每封重复)
   * @param {string[]} [params.bcc] - 密送(每封重复)
   * @param {Array<{filename:string, content_base64:string, content_type:string}>} [params.attachments]
   * @param {Object<string,string>} [params.headers] - 自定义邮件头
   * @param {number} [params.account_id]
   * @param {object} [opts]
   * @param {string} [opts.idempotencyKey] - Idempotency-Key 幂等头
   * @returns {Promise<object>} {total, success, failed, results: [{to, status, message_id?, error?}]}
   */
  sendBatch(params, opts = {}) {
    const headers = this._headers();
    if (opts.idempotencyKey) headers['Idempotency-Key'] = opts.idempotencyKey;
    return this._c.http.request('POST', '/api/v1/smtp/send/batch', {
      body: params,
      headers,
    });
  }

  /**
   * 按模板 code 渲染发送单封邮件.
   *
   * `POST /api/v1/smtp/send/template`
   *
   * 模板需要先在用户后台 "SMTP API → 模板管理" 创建.
   *
   * @param {object} params
   * @param {string} params.template_code - 模板 code(必填)
   * @param {string} params.to - 收件邮箱(必填)
   * @param {Object<string,string>} [params.variables] - 渲染变量(对应模板中 `{{var_name}}` 占位符)
   * @param {string} [params.from_name] - 发件人显示名
   * @returns {Promise<object>} {message_id, status, used_smtp}
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
   * @returns {Promise<object>} {daily_limit, daily_used, daily_remaining,
   *   monthly_limit, monthly_used, monthly_remaining, expire_at}
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
   * @returns {Promise<object>} {message_id, status, from_email, to_email, subject, is_html,
   *   account_id, account_name, error_message, smtp_response, send_duration_ms,
   *   opened_at, open_count, clicked_at, click_count, created_at, ...}
   *   status:pending=待处理,sending=发送中,success=成功,failed=失败
   */
  getStatus(messageId) {
    return this._c.http.request('GET', `/api/v1/smtp/status/${messageId}`, {
      headers: this._headers(),
    });
  }

  /**
   * 上报退信/投诉事件(自动把邮箱加入抑制名单并标记对应 send_log).
   *
   * `POST /api/v1/smtp/inbound`
   *
   * @param {object} params - email / message_id 至少提供其一
   * @param {string} [params.email] - 退信/投诉的邮箱
   * @param {string} [params.message_id] - 关联邮件的 message_id
   * @param {string} [params.type] - bounce | complaint
   * @returns {Promise<object>} {ok: true}
   */
  reportInbound(params) {
    return this._c.http.request('POST', '/api/v1/smtp/inbound', {
      body: params,
      headers: this._headers(),
    });
  }
}

module.exports = { Smtp };
