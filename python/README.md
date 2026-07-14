# NexCore Python SDK

全能 Python 客户端,覆盖 Payment / Exchange / Energy / SMTP / Withdraw / Account / VCard **7 大命名空间全部 44 个 v1 公开 endpoint**.

## 环境

- Python 3.8+
- 依赖:`requests`

## 安装

### 方式一:pip(推荐)

```bash
pip install nexcore-sdk
```

(SDK 包发布到 PyPI 后)

### 方式二:直接复制

```bash
cp -r sdk/python/nexcore your-project/
pip install requests
```

## 包结构

```
nexcore/
├── __init__.py          公开 from nexcore import Client, NexCoreError
├── client.py            主客户端入口
├── http.py              底层 HTTP 传输
├── errors.py            统一异常 NexCoreError
└── namespaces/
    ├── payment.py       多链收款(7 endpoints)
    ├── exchange.py      汇率(5 endpoints)
    ├── energy.py        TRON 能量租赁(8 endpoints)
    ├── smtp.py          SMTP 聚合 API(6 endpoints)
    ├── withdraw.py      提币(4 endpoints,RSA 签名)
    ├── account.py       账户(2 endpoints)
    └── vcard.py         虚拟信用卡(12 endpoints)
```

## 用法

```python
from nexcore import Client, NexCoreError

client = Client(
    base_url="https://your-domain.com",
    payment_app_id="APP20260412XXXX",
    payment_app_key="your_app_key_here",
    energy_api_key="energy_api_key_here",
    energy_secret_key="energy_secret_key_here",
    smtp_api_key="smk_xxx",
    timeout=30,
)

try:
    # 创建支付订单
    order = client.payment.create_order(
        out_order_id=f"ORDER_{int(time.time())}",
        amount="100.00",
        currency="CNY",
        trade_type="usdt.trc20",
        call_type="rotation",
        timeout=1800,
    )
    print("支付地址:", order["pay_address"])

    # 查询汇率
    rate = client.exchange.get_rate("USDT", "CNY")
    print(f"USDT/CNY: {rate['rate']}")

    # 估算能量
    est = client.energy.estimate_energy("TXxxxxxxxxxxxxxxxxxxxxx")
    print("建议能量:", est["suggested_energy"])

    # 发送邮件
    mail = client.smtp.send(
        to="user@example.com",
        subject="验证码",
        body="<h1>123456</h1>",
        is_html=True,
    )
    print("消息 ID:", mail["message_id"])

except NexCoreError as e:
    print(f"Error #{e.code}: {e} (trace: {e.request_id})")
```

## API 列表

### `client.payment` — 多链收款(7 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `create_order(**params)` | POST | `/api/v1/pay/create` |
| `query_order(out_order_id)` | GET | `/api/v1/pay/query` |
| `close_order(out_order_id)` | POST | `/api/v1/pay/close` |
| `get_app_config()` | GET | `/api/v1/pay/app-config` |
| `bind_address(user_id, trade_type)` | POST | `/api/v1/pay/bind-address` |
| `get_user_address(user_id)` | POST | `/api/v1/pay/get-address` |
| `unbind_address(user_id)` | POST | `/api/v1/pay/unbind-address` |
| `sign(params)` | (工具) | HMAC-SHA256 签名 |
| `verify_notify_sign(payload)` | (工具) | webhook 校验(常量时间) |

### `client.exchange` — 汇率(5 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `get_rate(from_, to)` | GET | `/api/v1/rate` |
| `convert(from_, to, amount)` | POST | `/api/v1/convert` |
| `get_rates(symbols, base=None)` | GET | `/api/v1/rates` |
| `get_fiat_rates(base)` | GET | `/api/v1/rates/fiat` |
| `get_all_rates(base)` | GET | `/api/v1/rates/all` |

注:`from_` 参数尾下划线是为避开 Python `from` 关键字;`get_rates` 的 `base` 不传时由后端取默认(USDT).

### `client.energy` — TRON 能量租赁(8 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `get_info()` | GET | `/api/v1/energy/info` |
| `get_price(energy_amount, period='1D')` | GET | `/api/v1/energy/price?energy_amount=&period=` |
| `estimate_energy(to_address)` | GET | `/api/v1/energy/estimate-energy?to_address=` |
| `create_order(**params)` | POST | `/api/v1/energy/order` |
| `create_onetime_order(**params)` | POST | `/api/v1/energy/order/onetime` |
| `query_order(serial)` | GET | `/api/v1/energy/order/:serial` |
| `list_orders(**filter_)` | GET | `/api/v1/energy/orders` |
| `reclaim_order(serial)` | POST | `/api/v1/energy/order/reclaim` |

注:租期 `period` 枚举 `1H / 1D / 3D / 7D / 30D`;`create_order` 必填 `receive_address` / `energy_amount` / `period`,可选 `out_trade_no` / `remark`.

### `client.smtp` — SMTP 聚合(6 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `send(to, subject, body, ...)` | POST | `/api/v1/smtp/send` |
| `send_batch(recipients, ...)` | POST | `/api/v1/smtp/send/batch` |
| `send_template(to, template_code, variables, from_name=None)` | POST | `/api/v1/smtp/send/template` |
| `get_quota()` | GET | `/api/v1/smtp/quota` |
| `get_status(message_id)` | GET | `/api/v1/smtp/status/:message_id` |
| `report_inbound(email=None, message_id=None, type=None)` | POST | `/api/v1/smtp/inbound` |

- `send` 可选关键字:`from_name` / `reply_to` / `text_body` / `headers` / `cc` / `bcc` / `attachments` / `account_id` / `send_at`(定时,RFC3339)/ `idempotency_key`(写入 `Idempotency-Key` 幂等头)
- `send_batch` 必填 `recipients` 列表(元素 `{to, variables?, from_name?}`),静态 `subject`+`body` 或 `template_code` 二选一;同样支持 `idempotency_key`
- `get_quota` 返回 `daily_limit/daily_used/daily_remaining` / `monthly_*` / `expire_at`
- `report_inbound` 上报退信/投诉(`email` 与 `message_id` 至少其一,`type` = `bounce` | `complaint`)

### `client.withdraw` — 提币(4 endpoint,RSA-PKCS1v15-SHA256 签名)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `create_withdraw(params)` | POST | `/api/v1/withdraw` |
| `get_withdraw(order_id)` | GET | `/api/v1/withdraw/:id` |
| `get_withdrawable_balance()` | GET | `/api/v1/balance/withdrawable` |
| `quote_fee(chain, symbol, amount)` | GET | `/api/v1/fee/quote`(amount 必填) |
| `sign(...)` / `verify_callback(...)` | (工具) | RSA 签名 / 平台回调验签 |

### `client.account` — 账户(2 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `get_balance()` | GET | `/api/v1/account/balance` |
| `get_deposit_address()` | GET | `/api/v1/account/deposit-address` |

### `client.vcard` — 虚拟信用卡(12 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `get_info()` / `list_bins()` / `list_cards()` | GET | `/api/v1/vcard/*`(读,X-API-Key) |
| `get_card_transactions(card_id)` / `list_orders(**params)` / `get_order(order_id)` | GET | 同上 |
| `update_card_remark(card_id, remark)` | POST | 同上 |
| `get_card_details(card_id)` / `get_card_code(card_id)` | GET | 敏感读(HMAC 头签名) |
| `open_card(**params)` / `recharge_card(card_id, **params)` / `cancel_card(card_id)` | POST | 资金操作(HMAC 头签名) |

## Webhook 签名校验

```python
import flask
app = flask.Flask(__name__)

@app.route("/payment/notify", methods=["POST"])
def notify():
    payload = flask.request.get_json()
    if not client.payment.verify_notify_sign(payload):
        return "invalid sign", 400
    # 处理订单回调... 务必幂等
    return "OK"
```

`verify_notify_sign` 内部用 `hmac.compare_digest`,常量时间比较防时序攻击.

## 异常

`nexcore.NexCoreError`:

- `code` — 平台错误码(0=成功)
- `str(e)` — 错误描述
- `request_id` — 服务端追踪 ID(响应头 `X-Trace-Id`)
- `http_status` — HTTP 状态码

## 示例

见 [`examples/`](./examples/):
- `create_order.py` — 完整下单
- `webhook_flask.py` — Flask 接收回调
