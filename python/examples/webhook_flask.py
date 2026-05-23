"""NexCore Python SDK — Webhook 回调签名校验(Flask 示例).

部署:
    pip install flask requests
    gunicorn -b 0.0.0.0:8000 webhook_flask:app

然后在 NexCore 用户后台「应用配置」的 notify_url 填上你的 URL。

NexCore 支付成功后会 POST JSON 到这里,本示例:
    1. 校验签名(SDK 一行搞定,内部用 hmac.compare_digest 常量时间比较)
    2. 业务处理(发货 / 更新 DB,务必幂等)
    3. 返回 200 OK(否则平台会重试)
"""
import os
import sys
import logging

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from flask import Flask, request, jsonify
from nexcore import Client

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)
log = logging.getLogger("nexcore-webhook")

client = Client(
    base_url=os.getenv("NEXCORE_BASE_URL", "https://your-domain.com"),
    payment_app_id=os.getenv("NEXCORE_APP_ID", "APP20260412XXXX"),
    payment_app_key=os.getenv("NEXCORE_APP_KEY", "your_app_key_here"),
)


@app.route("/payment/notify", methods=["POST"])
def payment_notify():
    payload = request.get_json(silent=True)
    if not isinstance(payload, dict):
        return "invalid payload", 400

    # 1. 校验签名(常量时间比较,防时序攻击)
    if not client.payment.verify_notify_sign(payload):
        log.warning("[nexcore] sign 校验失败: %s", str(payload)[:300])
        return "invalid sign", 400

    # 2. 业务处理(示例)
    # 同一订单可能因网络重试收到多次回调,务必做幂等(DB 唯一索引 out_order_id 等)
    order_id = payload.get("order_id", "")
    out_order = payload.get("out_order_id", "")
    status = int(payload.get("status", 0))
    amount = payload.get("amount", "")
    tx_hash = payload.get("tx_hash", "")

    # 状态:1=已支付  2=待支付  3=已关闭  4=已退款
    if status == 1:
        log.info("[nexcore] 订单已支付: %s = %s (tx: %s)", out_order, amount, tx_hash)
        # TODO: DB 查 out_order_id,判断是否已发货,未发货才发货

    # 3. 必须返回 200
    return "OK", 200


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000, debug=False)
