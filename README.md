# NexCore Official SDKs

NexCore 综合数字基础服务平台官方 SDK 集合,**一个客户端覆盖全部业务模块**。

## 业务覆盖

| 模块 | 路径前缀 | 认证 |
|---|---|---|
| 链收款 Payment | `/api/v1/pay/*` | HMAC-SHA256(app_id + app_key) |
| 能量租赁 Energy | `/api/v1/energy/*` | X-API-Key + X-Secret-Key |
| SMTP 聚合 API | `/api/v1/smtp/*` | X-API-Key |
| Astrenix AI(LLM) | `/v1/chat/completions` 等 | `Authorization: Bearer sk-nc-xxx` |
| 多链闪兑 Swap | `/api/v1/swap/*` | HMAC-SHA256 |
| 虚拟信用卡 Vcard | `/api/v1/vcard/*` | HMAC-SHA256 |

## 支持语言

| 语言 | 目录 | 入口 |
|---|---|---|
| PHP 7.4+ | [`php/`](./php/) | `require 'php/Client.php'` |
| Python 3.8+ | [`python/`](./python/) | `from nexcore import Client` |
| Node.js 16+ | [`node/`](./node/) | `const { Client } = require('./node/client')` |
| Go 1.21+ | [`go/`](./go/) | `import "nexcore-sdk"` |

## 统一设计

所有语言遵循同样的设计:

```text
Client(base_url, [auth options])
  ├─ .payment     ← 链收款 namespace
  │   ├─ .createOrder(params)
  │   ├─ .queryOrder(out_order_id)
  │   ├─ .closeOrder(out_order_id)
  │   ├─ .bindAddress(user_id, trade_type)
  │   ├─ .getAddress(user_id, trade_type)
  │   ├─ .unbindAddress(user_id)
  │   └─ .appConfig()
  ├─ .energy      ← 能量租赁
  │   ├─ .info()
  │   ├─ .price(energy, period)
  │   ├─ .estimateEnergy(receive_addr)
  │   ├─ .createOrder(params)
  │   ├─ .queryOrder(order_id)
  │   └─ .listOrders(filter)
  ├─ .smtp        ← SMTP 聚合
  │   ├─ .sendMail(params)
  │   ├─ .createTemplate(params)
  │   └─ .listAccounts()
  └─ .ai          ← Astrenix AI(OpenAI 兼容)
      ├─ .chat(messages, model)
      ├─ .completions(prompt, model)
      └─ .models()
```

## 鉴权配置

支持同一 Client 实例同时挂多种凭据 — 不同业务自动用对应凭据。

```python
# Python 示例
from nexcore import Client

client = Client(
    base_url="https://your-domain.com",
    payment_app_id="APP20260412XXXX",
    payment_app_key="your_app_key",
    energy_api_key="X-API-Key-Value",
    energy_secret_key="X-Secret-Key-Value",
    ai_api_key="sk-nc-xxx",
)

# 创建支付订单
order = client.payment.create_order(out_order_id="ORDER_001", amount=100, currency="CNY", trade_type="usdt.trc20")

# 估算 TRC20 能量
estimate = client.energy.estimate_energy(receive_addr="TXxxxxxxxxxxxxxxxxxxxxx")

# 跟 LLM 聊
reply = client.ai.chat(messages=[{"role": "user", "content": "Hello"}], model="claude-opus-4-7")
```

## 错误处理

所有 SDK 统一抛出 `NexCoreError`(各语言适配名),包含:
- `code` — 平台返回错误码(0 = 成功)
- `message` — 错误信息
- `request_id` — 服务端日志 ID(排查问题用)
- `http_status` — HTTP 状态码

## 版本

当前 SDK 版本:**v3.0**(跟 NexCore 后端 API 一致)

完整 API 字段说明请参考在线文档:[https://your-domain.com/docs](/docs)

## 反馈

发现 SDK Bug 或缺接口请通过 NexCore 用户后台「工单」反馈,或在 GitHub Issues 提交。
