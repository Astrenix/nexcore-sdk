"""NexCore Python SDK — Astrenix AI 对话(OpenAI 兼容协议).

运行:
    python examples/ai_chat.py
"""
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from nexcore import Client, NexCoreError

client = Client(
    base_url=os.getenv("NEXCORE_BASE_URL", "https://your-domain.com"),
    ai_api_key=os.getenv("NEXCORE_AI_KEY", "sk-nc-xxx"),
)

try:
    reply = client.ai.chat(
        messages=[
            {"role": "system", "content": "你是一个简洁的助手,回答不超过 2 句。"},
            {"role": "user",   "content": "介绍一下 NexCore"},
        ],
        model="claude-opus-4-7",
        temperature=0.7,
        max_tokens=512,
    )

    content = reply["choices"][0]["message"]["content"]
    print(f"🤖 Claude:\n{content}\n")

    usage = reply.get("usage", {})
    print(f"Usage: {usage.get('prompt_tokens', 0)} → {usage.get('completion_tokens', 0)} tokens")

    # 列出可用模型
    print("\n可用模型:")
    models = client.ai.models()
    for m in models.get("data", []):
        print(f"  - {m['id']}")

except NexCoreError as e:
    print(f"❌ Error #{e.code}: {e}")
    if e.request_id:
        print(f"  Trace ID: {e.request_id}")
    sys.exit(1)
