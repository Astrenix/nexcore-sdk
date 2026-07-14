/**
 * NexCore Official Node.js SDK — Type definitions.
 */

export interface ClientConfig {
  baseUrl: string;
  paymentAppId?: string;
  paymentAppKey?: string;
  energyApiKey?: string;
  energySecretKey?: string;
  smtpApiKey?: string;
  apiKey?: string;
  apiSecret?: string;
  withdrawApiKey?: string;
  withdrawPrivateKeyPem?: string;
  withdrawPlatformPublicKeyPem?: string;
  timeout?: number;
  userAgent?: string;
}

export class NexCoreError extends Error {
  code: number;
  requestId: string | null;
  httpStatus: number | null;
  constructor(message: string, code?: number, requestId?: string | null, httpStatus?: number | null);
}

export class Client {
  constructor(config: ClientConfig);
  static readonly VERSION: string;
  static verifyWebhook(params: Record<string, unknown>, secret: string): boolean;
  payment: PaymentNamespace;
  exchange: ExchangeNamespace;
  energy: EnergyNamespace;
  smtp: SmtpNamespace;
  withdraw: WithdrawNamespace;
  account: AccountNamespace;
  vcard: VCardNamespace;
}

// ============ Payment ============

export interface PaymentCreateOrderParams {
  out_order_id: string;
  amount: string | number;
  currency: string;
  trade_type: string;
  /** 必填;本接口只支持 rotation,一对一模式走 bindAddress/getUserAddress */
  call_type: 'rotation';
  /** 用户标识(一对一模式必填) */
  out_user_id?: string;
  timeout?: number;
  subject?: string;
  notify_url?: string;
  return_url?: string;
  order_type?: 'normal' | 'recharge';
  [k: string]: unknown;
}

export interface PaymentNamespace {
  sign(params: Record<string, unknown>): string;
  createOrder(params: PaymentCreateOrderParams): Promise<Record<string, unknown>>;
  queryOrder(outOrderId: string): Promise<Record<string, unknown>>;
  closeOrder(outOrderId: string): Promise<Record<string, unknown>>;
  getAppConfig(): Promise<Record<string, unknown>>;
  bindAddress(userId: string, tradeType: string): Promise<Record<string, unknown>>;
  /** 签名只含 app_id + user_id(与 bindAddress 不同,无 trade_type) */
  getUserAddress(userId: string): Promise<Record<string, unknown>>;
  unbindAddress(userId: string): Promise<Record<string, unknown>>;
  verifyNotifySign(payload: Record<string, unknown>): boolean;
}

// ============ Exchange ============

export interface ExchangeNamespace {
  getRate(from: string, to: string): Promise<Record<string, unknown>>;
  convert(from: string, to: string, amount: string | number): Promise<Record<string, unknown>>;
  getRates(symbols: string[], base?: string): Promise<Record<string, unknown>>;
  getFiatRates(base?: string): Promise<Record<string, unknown>>;
  getAllRates(base?: string): Promise<Record<string, unknown>>;
}

// ============ Energy ============

export type EnergyPeriod = '1H' | '1D' | '3D' | '7D' | '30D';

export interface EnergyCreateOrderParams {
  receive_address: string;
  energy_amount: number;
  period: EnergyPeriod;
  out_trade_no?: string;
  remark?: string;
  [k: string]: unknown;
}

export interface EnergyOnetimeOrderParams {
  receive_address: string;
  period: EnergyPeriod;
  out_trade_no?: string;
  remark?: string;
  [k: string]: unknown;
}

export interface EnergyNamespace {
  getInfo(): Promise<Record<string, unknown>>;
  getPrice(energyAmount: number, period?: EnergyPeriod): Promise<Record<string, unknown>>;
  estimateEnergy(toAddress: string): Promise<Record<string, unknown>>;
  createOrder(params: EnergyCreateOrderParams): Promise<Record<string, unknown>>;
  createOnetimeOrder(params: EnergyOnetimeOrderParams): Promise<Record<string, unknown>>;
  queryOrder(serial: string): Promise<Record<string, unknown>>;
  listOrders(filter?: { page?: number; page_size?: number; status?: -1 | 0 | 40 | 41 }): Promise<Record<string, unknown>>;
  reclaimOrder(serial: string): Promise<Record<string, unknown>>;
}

// ============ SMTP ============

export interface SmtpAttachment {
  filename: string;
  content_base64: string;
  content_type: string;
}

export interface SmtpSendParams {
  to: string;
  subject: string;
  body: string;
  is_html?: boolean;
  from_name?: string;
  reply_to?: string;
  /** 纯文本正文;与 HTML 同时提供时以 multipart/alternative 发送 */
  text_body?: string;
  /** 自定义邮件头(核心头不可覆盖) */
  headers?: Record<string, string>;
  cc?: string[];
  bcc?: string[];
  attachments?: SmtpAttachment[];
  account_id?: number;
  /** 定时发送(RFC3339);晚于当前 30s 以上则排期 */
  send_at?: string;
}

export interface SmtpBatchRecipient {
  to: string;
  variables?: Record<string, string>;
  from_name?: string;
}

export interface SmtpSendBatchParams {
  /** 收件人列表(必填);上限 = 订阅 max_batch_size(默认 10) */
  recipients: SmtpBatchRecipient[];
  /** 静态模式主题 */
  subject?: string;
  /** 静态模式正文,支持 {{var}} 替换;与 template_code 二选一 */
  body?: string;
  /** 模板模式:模板 code;与 body 二选一 */
  template_code?: string;
  is_html?: boolean;
  reply_to?: string;
  cc?: string[];
  bcc?: string[];
  attachments?: SmtpAttachment[];
  headers?: Record<string, string>;
  account_id?: number;
}

export interface SmtpSendTemplateParams {
  template_code: string;
  to: string;
  variables?: Record<string, string>;
  from_name?: string;
}

export interface SmtpSendOptions {
  /** Idempotency-Key 头;同 key 重试直接返回首次结果,不重复发送/扣配额 */
  idempotencyKey?: string;
}

export interface SmtpNamespace {
  send(params: SmtpSendParams, opts?: SmtpSendOptions): Promise<Record<string, unknown>>;
  sendBatch(params: SmtpSendBatchParams, opts?: SmtpSendOptions): Promise<Record<string, unknown>>;
  sendTemplate(params: SmtpSendTemplateParams): Promise<Record<string, unknown>>;
  getQuota(): Promise<Record<string, unknown>>;
  getStatus(messageId: string): Promise<Record<string, unknown>>;
  reportInbound(params: { email?: string; message_id?: string; type?: 'bounce' | 'complaint' }): Promise<Record<string, unknown>>;
}

// ============ Withdraw (多链收款 · 提币端,RSA-2048) ============

export interface WithdrawCreateParams {
  chain: 'tron' | 'eth' | 'bsc' | 'polygon' | 'arbitrum' | 'btc' | string;
  symbol: string;
  amount: string;
  to_address: string;
  memo?: string;
  callback_url?: string;
  request_id?: string;
  [k: string]: unknown;
}

export interface WithdrawNamespace {
  sign(method: string, path: string, timestamp: string, nonce: string, body: string): string;
  createWithdraw(params: WithdrawCreateParams): Promise<Record<string, unknown>>;
  getWithdraw(orderId: string): Promise<Record<string, unknown>>;
  getWithdrawableBalance(): Promise<Record<string, unknown>>;
  quoteFee(chain: string, symbol: string, amount: string): Promise<Record<string, unknown>>;
  verifyCallback(
    method: string,
    path: string,
    timestamp: string,
    nonce: string,
    body: string,
    base64Signature: string
  ): void;
}

// ============ Account (MPK apiKey + apiSecret 双密钥) ============

export interface AccountNamespace {
  getBalance(): Promise<Record<string, unknown>>;
  getDepositAddress(): Promise<Record<string, unknown>>;
}

// ============ VCard (虚拟信用卡 · 双密钥读 + HMAC 头签名) ============

export interface VCardListOrdersQuery {
  page?: number;
  page_size?: number;
  status?: string;
  order_type?: string;
  [k: string]: unknown;
}

export interface VCardNamespace {
  // 双密钥 header(只读)
  getInfo(): Promise<Record<string, unknown>>;
  listBins(): Promise<Record<string, unknown>>;
  listCards(): Promise<Record<string, unknown>>;
  getCardTransactions(cardId: string | number): Promise<Record<string, unknown>>;
  listOrders(query?: VCardListOrdersQuery): Promise<Record<string, unknown>>;
  getOrder(orderId: string | number): Promise<Record<string, unknown>>;
  updateCardRemark(cardId: string | number, remark: string): Promise<Record<string, unknown>>;
  // HMAC 头签名(敏感 / 写)
  sign(ts: string, nonce: string, method: string, path: string, rawQuery: string, body: string): string;
  getCardDetails(cardId: string | number): Promise<Record<string, unknown>>;
  getCardCode(cardId: string | number): Promise<Record<string, unknown>>;
  openCard(params: Record<string, unknown>): Promise<Record<string, unknown>>;
  rechargeCard(cardId: string | number, params: Record<string, unknown>): Promise<Record<string, unknown>>;
  cancelCard(cardId: string | number): Promise<Record<string, unknown>>;
}
