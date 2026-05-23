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
  payment: PaymentNamespace;
  exchange: ExchangeNamespace;
  energy: EnergyNamespace;
  smtp: SmtpNamespace;
  withdraw: WithdrawNamespace;
}

// ============ Payment ============

export interface PaymentCreateOrderParams {
  out_order_id: string;
  amount: string | number;
  currency: string;
  trade_type: string;
  call_type?: 'rotation' | 'one_to_one';
  user_id?: string;
  timeout?: number;
  subject?: string;
  notify_url?: string;
  return_url?: string;
  [k: string]: unknown;
}

export interface PaymentNamespace {
  sign(params: Record<string, unknown>): string;
  createOrder(params: PaymentCreateOrderParams): Promise<Record<string, unknown>>;
  queryOrder(outOrderId: string): Promise<Record<string, unknown>>;
  closeOrder(outOrderId: string): Promise<Record<string, unknown>>;
  getAppConfig(): Promise<Record<string, unknown>>;
  bindAddress(userId: string, tradeType: string): Promise<Record<string, unknown>>;
  getUserAddress(userId: string, tradeType: string): Promise<Record<string, unknown>>;
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

export interface EnergyCreateOrderParams {
  receive_addr: string;
  energy: number;
  period: '1H' | '6H' | '1D' | '3D' | '1W';
  out_serial?: string;
  [k: string]: unknown;
}

export interface EnergyNamespace {
  getInfo(): Promise<Record<string, unknown>>;
  getPrice(energy: number, period?: string): Promise<Record<string, unknown>>;
  estimateEnergy(receiveAddr: string): Promise<Record<string, unknown>>;
  createOrder(params: EnergyCreateOrderParams): Promise<Record<string, unknown>>;
  createOnetimeOrder(params: EnergyCreateOrderParams): Promise<Record<string, unknown>>;
  queryOrder(serial: string): Promise<Record<string, unknown>>;
  listOrders(filter?: Record<string, unknown>): Promise<Record<string, unknown>>;
  reclaimOrder(serial: string): Promise<Record<string, unknown>>;
}

// ============ SMTP ============

export interface SmtpSendParams {
  to: string;
  subject: string;
  body: string;
  is_html?: boolean;
  account_id?: number;
  reply_to?: string;
}

export interface SmtpSendBatchParams {
  to: string[];
  subject: string;
  body: string;
  is_html?: boolean;
  account_id?: number;
}

export interface SmtpSendTemplateParams {
  to: string;
  template_id: number;
  variables: Record<string, unknown>;
  account_id?: number;
}

export interface SmtpNamespace {
  send(params: SmtpSendParams): Promise<Record<string, unknown>>;
  sendBatch(params: SmtpSendBatchParams): Promise<Record<string, unknown>>;
  sendTemplate(params: SmtpSendTemplateParams): Promise<Record<string, unknown>>;
  getQuota(): Promise<Record<string, unknown>>;
  getStatus(messageId: string): Promise<Record<string, unknown>>;
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
  quoteFee(chain: string, symbol: string, amount?: string): Promise<Record<string, unknown>>;
  verifyCallback(
    method: string,
    path: string,
    timestamp: string,
    nonce: string,
    body: string,
    base64Signature: string
  ): void;
}
