<?php
/**
 * NexCore PHP SDK — Astrenix AI 对话(OpenAI 兼容协议)
 *
 * 执行:
 *   php examples/ai_chat.php
 */

require_once __DIR__ . '/../Client.php';

use NexCore\Client;
use NexCore\NexCoreError;

$client = new Client([
    'base_url'   => getenv('NEXCORE_BASE_URL') ?: 'https://your-domain.com',
    'ai_api_key' => getenv('NEXCORE_AI_KEY')   ?: 'sk-nc-xxx',
]);

try {
    // 普通对话
    $reply = $client->ai->chat(
        [
            ['role' => 'system', 'content' => '你是一个简洁的助手,回答不超过 2 句。'],
            ['role' => 'user',   'content' => '介绍一下 NexCore'],
        ],
        'claude-opus-4-7',
        ['temperature' => 0.7, 'max_tokens' => 512]
    );

    $content = $reply['choices'][0]['message']['content'] ?? '(无响应)';
    echo "🤖 Claude:\n{$content}\n\n";
    echo "Usage: {$reply['usage']['prompt_tokens']} → {$reply['usage']['completion_tokens']} tokens\n";

    // 列出可用模型
    echo "\n可用模型:\n";
    $models = $client->ai->models();
    foreach (($models['data'] ?? []) as $m) {
        echo "  - {$m['id']}\n";
    }

} catch (NexCoreError $e) {
    echo "❌ Error #{$e->code}: {$e->getMessage()}\n";
    if ($e->requestId) echo "  Trace ID: {$e->requestId}\n";
    exit(1);
}
