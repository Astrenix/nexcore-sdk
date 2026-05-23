/**
 * NexCore Node.js SDK — Astrenix AI 对话(OpenAI 兼容协议)
 *
 * 运行:
 *   node examples/ai_chat.js
 */

const { Client, NexCoreError } = require('../index');

const client = new Client({
  baseUrl: process.env.NEXCORE_BASE_URL || 'https://your-domain.com',
  aiApiKey: process.env.NEXCORE_AI_KEY || 'sk-nc-xxx',
});

(async () => {
  try {
    const reply = await client.ai.chat(
      [
        { role: 'system', content: '你是一个简洁的助手,回答不超过 2 句。' },
        { role: 'user', content: '介绍一下 NexCore' },
      ],
      'claude-opus-4-7',
      { temperature: 0.7, max_tokens: 512 }
    );

    const content = reply.choices?.[0]?.message?.content || '(无响应)';
    console.log(`🤖 Claude:\n${content}\n`);

    const usage = reply.usage || {};
    console.log(`Usage: ${usage.prompt_tokens || 0} → ${usage.completion_tokens || 0} tokens`);

    // 列出可用模型
    console.log('\n可用模型:');
    const models = await client.ai.models();
    for (const m of models.data || []) {
      console.log(`  - ${m.id}`);
    }
  } catch (e) {
    if (e instanceof NexCoreError) {
      console.error(`❌ Error #${e.code}: ${e.message}`);
      if (e.requestId) console.error(`  Trace ID: ${e.requestId}`);
      process.exit(1);
    }
    throw e;
  }
})();
