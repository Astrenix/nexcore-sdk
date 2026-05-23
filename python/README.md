# NexCore Python SDK

全能 Python 客户端,覆盖 Payment / Exchange / Energy / SMTP **全部 25 个 v1 公开 endpoint**.

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
    └── smtp.py          SMTP 聚合 API(5 endpoints)
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
    print("需要能量:", est["estimated_energy"])

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
| `get_user_address(user_id, trade_type)` | POST | `/api/v1/pay/get-address` |
| `unbind_address(user_id)` | POST | `/api/v1/pay/unbind-address` |
| `sign(params)` | (工具) | HMAC-SHA256 签名 |
| `verify_notify_sign(payload)` | (工具) | webhook 校验(常量时间) |

### `client.exchange` — 汇率(5 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `get_rate(from_, to)` | GET | `/api/v1/rate` |
| `convert(from_, to, amount)` | POST | `/api/v1/convert` |
| `get_rates(symbols, base)` | GET | `/api/v1/rates` |
| `get_fiat_rates(base)` | GET | `/api/v1/rates/fiat` |
| `get_all_rates(base)` | GET | `/api/v1/rates/all` |

注:`from_` 参数尾下划线是为避开 Python `from` 关键字.

### `client.energy` — TRON 能量租赁(8 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `get_info()` | GET | `/api/v1/energy/info` |
| `get_price(energy, period='1D')` | GET | `/api/v1/energy/price` |
| `estimate_energy(receive_addr)` | GET | `/api/v1/energy/estimate-energy` |
| `create_order(**params)` | POST | `/api/v1/energy/order` |
| `create_onetime_order(**params)` | POST | `/api/v1/energy/order/onetime` |
| `query_order(serial)` | GET | `/api/v1/energy/order/:serial` |
| `list_orders(**filter_)` | GET | `/api/v1/energy/orders` |
| `reclaim_order(serial)` | POST | `/api/v1/energy/order/reclaim` |

### `client.smtp` — SMTP 聚合(5 endpoint)

| Python 方法 | HTTP | endpoint |
|---|---|---|
| `send(to, subject, body, ...)` | POST | `/api/v1/smtp/send` |
| `send_batch(to, subject, body, ...)` | POST | `/api/v1/smtp/send/batch` |
| `send_template(to, template_id, variables, ...)` | POST | `/api/v1/smtp/send/template` |
| `get_quota()` | GET | `/api/v1/smtp/quota` |
| `get_status(message_id)` | GET | `/api/v1/smtp/status/:message_id` |

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
