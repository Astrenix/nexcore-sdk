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
  aiApiKey?: string;
  timeout?: number;
}

export class NexCoreError extends Error {
  code: number;
  requestId: string | null;
  httpStatus: number | null;
  constructor(message: string, code?: number, requestId?: string | null, httpStatus?: number | null);
}

export class Client {
  constructor(config: ClientConfig);
  baseUrl: string;
  payment: PaymentNamespace;
  energy: EnergyNamespace;
  smtp: SmtpNamespace;
  ai: AiNamespace;
}

export interface PaymentCreateOrderParams {
  out_order_id: string;
  amount: string | number;
  currency: string;
  trade_type: string;
  call_type?: 'rotation' | 'one_to_one';
  user_id?: string;
  timeout?: number;
  notify_url?: string;
  return_url?: string;
  subject?: string;
  [k: string]: any;
}

export interface PaymentNamespace {
  createOrder(params: PaymentCreateOrderParams): Promise<any>;
  queryOrder(outOrderId: string): Promise<any>;
  closeOrder(outOrderId: string): Promise<any>;
  bindAddress(userId: string, tradeType: string): Promise<any>;
  getAddress(userId: string, tradeType: string): Promise<any>;
  unbindAddress(userId: string): Promise<any>;
  appConfig(): Promise<any>;
  verifyNotifySign(payload: Record<string, any>): boolean;
}

export interface EnergyNamespace {
  info(): Promise<any>;
  price(energy: number, period?: string): Promise<any>;
  estimateEnergy(receiveAddr: string): Promise<any>;
  createOrder(params: Record<string, any>): Promise<any>;
  queryOrder(orderId: number | string): Promise<any>;
  listOrders(filter?: Record<string, any>): Promise<any>;
}

export interface SmtpNamespace {
  sendMail(params: Record<string, any>): Promise<any>;
  listAccounts(): Promise<any>;
  listTemplates(): Promise<any>;
}

export interface AiNamespace {
  chat(
    messages: Array<{ role: string; content: string }>,
    model: string,
    extra?: Record<string, any>,
  ): Promise<any>;
  models(): Promise<any>;
}
