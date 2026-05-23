'use strict';

/**
 * NexCore SDK 统一异常.
 *
 * 所有 SDK 调用失败(网络错误 / HTTP 4xx-5xx / 业务 code != 0)统一抛本异常.
 * 业务方通过 try/catch 后访问字段可定位问题.
 *
 * @example
 *   try {
 *     await client.payment.createOrder({...});
 *   } catch (e) {
 *     if (e instanceof NexCoreError) {
 *       console.error(e.code, e.message, e.requestId, e.httpStatus);
 *     }
 *   }
 */
class NexCoreError extends Error {
  /**
   * @param {string} message 错误描述
   * @param {number} [code=-1] 平台错误码(0=成功;-1=客户端层;其他参见错误码表)
   * @param {string|null} [requestId=null] 服务端日志追踪 ID(响应头 X-Trace-Id)
   * @param {number|null} [httpStatus=null] HTTP 状态码;客户端层错误时为 null
   */
  constructor(message, code = -1, requestId = null, httpStatus = null) {
    super(message);
    this.name = 'NexCoreError';
    this.code = code;
    this.requestId = requestId;
    this.httpStatus = httpStatus;
  }
}

module.exports = { NexCoreError };
