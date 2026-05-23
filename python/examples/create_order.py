"""NexCore Python SDK — 创建支付订单(轮播模式).

运行:
    python examples/create_order.py
"""
import os
import sys
import time

# 让 examples 能 import 仓库根目录的 nexcore.py
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from nexcore import Client, NexCoreError

client = Client(
    base_url=os.getenv('NEXCORE_BASE_URL', 'https://your-domain.com'),
    payment_app_id=os.getenv('NEXCORE_APP_ID', 'APP20260412XXXX'),
    payment_app_key=os.getenv('NEXCORE_APP_KEY', 'your_app_key_here'),
    timeout=30,
)

try:
    order = client.payment.create_order(
        out_order_id=f"ORDER_{int(time.time())}",
        amount="100.00",            # 必填:法币金额(string,两位小数,避免浮点)
        currency="CNY",             # CNY / USD / EUR / JPY / KRW / HKD
        trade_type="usdt.trc20",    # 加密币种.链
        call_type="rotation",       # rotation=轮播 / one_to_one=一对一
        timeout=1800,
        subject="会员充值",
        notify_url="https://your-domain.com/payment/notify",
        return_url="https://your-domain.com/payment/success",
    )

    print(f"✅ 订单创建成功")
    print(f"  订单号:    {order['order_id']}")
    print(f"  支付地址:  {order['pay_address']}")
    print(f"  加密金额:  {order['crypto_amount']} {order['crypto_currency']}")
    print(f"  过期时间:  {order['expires_at']}")

except NexCoreError as e:
    print(f"❌ Error #{e.code}: {e}")
    if e.request_id:
        print(f"  Trace ID: {e.request_id}")
    if e.http_status:
        print(f"  HTTP: {e.http_status}")
    sys.exit(1)
