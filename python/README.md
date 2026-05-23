# NexCore Python SDK

全能 Python 客户端,覆盖 Payment / Energy / SMTP / AI 全部 NexCore 业务。

## 环境

- Python 3.8+
- 依赖:`requests`

## 安装

直接拷贝 `nexcore.py` 到项目:

```bash
cp sdk/python/nexcore.py your-project/
```

依赖:

```bash
pip install requests
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
    ai_api_key="sk-nc-xxx",
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

    # 估算能量
    est = client.energy.estimate_energy("TXxxxxxxxxxxxxxxxxxxxxx")
    print("需要能量:", est["estimated_energy"])

    # AI 对话
    reply = client.ai.chat(
        messages=[{"role": "user", "content": "你好"}],
        model="claude-opus-4-7",
    )
    print(reply["choices"][0]["message"]["content"])

except NexCoreError as e:
    print(f"Error #{e.code}: {e} (trace: {e.request_id})")
```

## 异常

所有错误统一抛 `nexcore.NexCoreError`,字段:

- `code` — 平台错误码(0 = 成功)
- `args[0]` / `str(e)` — 错误描述
- `request_id` — 服务端日志追踪 ID(响应头 `X-Trace-Id`)
- `http_status` — HTTP 状态码

## Webhook 签名校验

```python
import flask

app = flask.Flask(__name__)

@app.route("/payment/notify", methods=["POST"])
def notify():
    payload = flask.request.get_json()
    if not client.payment.verify_notify_sign(payload):
        return "invalid sign", 400
    # 处理订单回调...
    return "OK"
```

## 示例

更多示例见 [`examples/`](./examples/) 目录。
