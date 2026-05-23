<?php
/**
 * NexCore Official PHP SDK
 *
 * 全能客户端,一次配置覆盖 Payment / Energy / SMTP / AI 全部业务。
 *
 * 使用:
 *   $client = new NexCore\Client([
 *       'base_url'         => 'https://your-domain.com',
 *       'payment_app_id'   => 'APP20260412XXXX',
 *       'payment_app_key'  => 'your_app_key_here',
 *       'energy_api_key'   => 'X-API-Key-Value',
 *       'energy_secret_key' => 'X-Secret-Key-Value',
 *       'ai_api_key'       => 'sk-nc-xxx',
 *       'timeout'          => 30,
 *   ]);
 *
 *   // 链收款
 *   $order = $client->payment->createOrder([
 *       'out_order_id' => 'ORDER_' . time(),
 *       'amount'       => '100.00',
 *       'currency'     => 'CNY',
 *       'trade_type'   => 'usdt.trc20',
 *       'call_type'    => 'rotation',
 *   ]);
 *
 *   // 能量租赁
 *   $est = $client->energy->estimateEnergy('TXxxxxxxxxxxxxxxxxxxxxx');
 *
 *   // AI chat
 *   $reply = $client->ai->chat([
 *       ['role' => 'user', 'content' => 'Hello']
 *   ], 'claude-opus-4-7');
 */

namespace NexCore;

class NexCoreError extends \RuntimeException
{
    public int $code;
    public ?string $requestId;
    public ?int $httpStatus;

    public function __construct(string $message, int $code = 0, ?string $requestId = null, ?int $httpStatus = null)
    {
        parent::__construct($message);
        $this->code = $code;
        $this->requestId = $requestId;
        $this->httpStatus = $httpStatus;
    }
}

class Client
{
    private string $baseUrl;
    private array $config;

    public PaymentNamespace $payment;
    public EnergyNamespace $energy;
    public SmtpNamespace $smtp;
    public AiNamespace $ai;

    public function __construct(array $config)
    {
        $this->baseUrl = rtrim($config['base_url'] ?? '', '/');
        $this->config = array_merge([
            'timeout' => 30,
            'verify_ssl' => true,
        ], $config);

        $this->payment = new PaymentNamespace($this);
        $this->energy = new EnergyNamespace($this);
        $this->smtp = new SmtpNamespace($this);
        $this->ai = new AiNamespace($this);
    }

    public function get(string $cfg)
    {
        return $this->config[$cfg] ?? null;
    }

    /**
     * Low-level HTTP request. Sub-namespaces compose params + headers and call this.
     *
     * @return array<mixed> 解析后的 JSON 响应(data 段),失败抛 NexCoreError
     */
    public function request(string $method, string $path, array $opts = []): array
    {
        $url = $this->baseUrl . $path;
        $body = $opts['body'] ?? null;
        $headers = $opts['headers'] ?? [];
        $query = $opts['query'] ?? [];

        if (!empty($query)) {
            $url .= (str_contains($url, '?') ? '&' : '?') . http_build_query($query);
        }

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, strtoupper($method));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $this->config['timeout']);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->config['verify_ssl']);
        curl_setopt($ch, CURLOPT_HEADER, true);

        $defaultHeaders = ['Content-Type: application/json', 'Accept: application/json'];
        $hdrLines = array_merge($defaultHeaders, $headers);
        curl_setopt($ch, CURLOPT_HTTPHEADER, $hdrLines);

        if ($body !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, is_string($body) ? $body : json_encode($body, JSON_UNESCAPED_UNICODE));
        }

        $response = curl_exec($ch);
        if ($response === false) {
            $err = curl_error($ch);
            curl_close($ch);
            throw new NexCoreError("HTTP request failed: $err", -1);
        }
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
        $respHeaders = substr($response, 0, $headerSize);
        $respBody = substr($response, $headerSize);
        curl_close($ch);

        $requestId = null;
        if (preg_match('/^X-Trace-Id:\s*(.+?)\r?$/im', $respHeaders, $m)) {
            $requestId = trim($m[1]);
        }

        $json = json_decode($respBody, true);
        if ($httpCode >= 400 || !is_array($json)) {
            throw new NexCoreError(
                $json['message'] ?? "HTTP $httpCode: " . substr($respBody, 0, 200),
                $json['code'] ?? -1,
                $requestId,
                $httpCode
            );
        }
        if (isset($json['code']) && $json['code'] !== 0) {
            throw new NexCoreError($json['message'] ?? 'unknown error', $json['code'], $requestId, $httpCode);
        }
        return $json['data'] ?? $json;
    }
}

/**
 * 链收款 namespace.
 * 所有请求自动用 payment_app_id + payment_app_key 签名(HMAC-SHA256)。
 */
class PaymentNamespace
{
    private Client $client;
    public function __construct(Client $c) { $this->client = $c; }

    private function sign(array $params): string
    {
        $key = $this->client->get('payment_app_key');
        if (!$key) throw new NexCoreError('payment_app_key not configured', -1);
        unset($params['sign']);
        $params = array_filter($params, fn($v) => $v !== '' && $v !== null);
        ksort($params);
        $pairs = [];
        foreach ($params as $k => $v) { $pairs[] = "$k=$v"; }
        return hash_hmac('sha256', implode('&', $pairs), $key);
    }

    private function signed(array $params): array
    {
        $appId = $this->client->get('payment_app_id');
        if (!$appId) throw new NexCoreError('payment_app_id not configured', -1);
        $params['app_id'] = $appId;
        $params['sign'] = $this->sign($params);
        return $params;
    }

    public function createOrder(array $params): array
    {
        return $this->client->request('POST', '/api/v1/pay/create', ['body' => $this->signed($params)]);
    }
    public function queryOrder(string $outOrderId): array
    {
        return $this->client->request('GET', '/api/v1/pay/query', ['query' => $this->signed(['out_order_id' => $outOrderId])]);
    }
    public function closeOrder(string $outOrderId): array
    {
        return $this->client->request('POST', '/api/v1/pay/close', ['body' => $this->signed(['out_order_id' => $outOrderId])]);
    }
    public function bindAddress(string $userId, string $tradeType): array
    {
        return $this->client->request('POST', '/api/v1/pay/bind-address', ['body' => $this->signed(['user_id' => $userId, 'trade_type' => $tradeType])]);
    }
    public function getAddress(string $userId, string $tradeType): array
    {
        return $this->client->request('GET', '/api/v1/pay/get-address', ['query' => $this->signed(['user_id' => $userId, 'trade_type' => $tradeType])]);
    }
    public function unbindAddress(string $userId): array
    {
        return $this->client->request('POST', '/api/v1/pay/unbind-address', ['body' => $this->signed(['user_id' => $userId])]);
    }
    public function appConfig(): array
    {
        return $this->client->request('GET', '/api/v1/pay/app-config', ['query' => $this->signed([])]);
    }

    /** 给业务方校验 webhook 回调签名用 */
    public function verifyNotifySign(array $notifyPayload): bool
    {
        $sign = $notifyPayload['sign'] ?? null;
        if (!$sign) return false;
        $expected = $this->sign($notifyPayload);
        return hash_equals($expected, $sign);
    }
}

/**
 * 能量租赁 namespace.
 * 自动给 header 加 X-API-Key + X-Secret-Key。
 */
class EnergyNamespace
{
    private Client $client;
    public function __construct(Client $c) { $this->client = $c; }

    private function authHeaders(): array
    {
        $k = $this->client->get('energy_api_key');
        $s = $this->client->get('energy_secret_key');
        if (!$k || !$s) throw new NexCoreError('energy_api_key / energy_secret_key not configured', -1);
        return ["X-API-Key: $k", "X-Secret-Key: $s"];
    }

    public function info(): array
    {
        return $this->client->request('GET', '/api/v1/energy/info', ['headers' => $this->authHeaders()]);
    }
    public function price(int $energy, string $period = '1D'): array
    {
        return $this->client->request('GET', '/api/v1/energy/price', [
            'headers' => $this->authHeaders(),
            'query' => ['energy' => $energy, 'period' => $period],
        ]);
    }
    public function estimateEnergy(string $receiveAddr): array
    {
        return $this->client->request('GET', '/api/v1/energy/estimate-energy', [
            'headers' => $this->authHeaders(),
            'query' => ['receive_addr' => $receiveAddr],
        ]);
    }
    public function createOrder(array $params): array
    {
        return $this->client->request('POST', '/api/v1/energy/order', ['headers' => $this->authHeaders(), 'body' => $params]);
    }
    public function queryOrder(int $orderId): array
    {
        return $this->client->request('GET', "/api/v1/energy/order/$orderId", ['headers' => $this->authHeaders()]);
    }
    public function listOrders(array $filter = []): array
    {
        return $this->client->request('GET', '/api/v1/energy/orders', ['headers' => $this->authHeaders(), 'query' => $filter]);
    }
}

/**
 * SMTP 聚合 API namespace.
 */
class SmtpNamespace
{
    private Client $client;
    public function __construct(Client $c) { $this->client = $c; }

    private function authHeaders(): array
    {
        $k = $this->client->get('smtp_api_key');
        if (!$k) throw new NexCoreError('smtp_api_key not configured', -1);
        return ["X-API-Key: $k"];
    }

    public function sendMail(array $params): array
    {
        return $this->client->request('POST', '/api/v1/smtp/send', ['headers' => $this->authHeaders(), 'body' => $params]);
    }
    public function listAccounts(): array
    {
        return $this->client->request('GET', '/api/v1/smtp/accounts', ['headers' => $this->authHeaders()]);
    }
    public function listTemplates(): array
    {
        return $this->client->request('GET', '/api/v1/smtp/templates', ['headers' => $this->authHeaders()]);
    }
}

/**
 * Astrenix AI namespace(OpenAI 兼容协议).
 */
class AiNamespace
{
    private Client $client;
    public function __construct(Client $c) { $this->client = $c; }

    private function authHeaders(): array
    {
        $k = $this->client->get('ai_api_key');
        if (!$k) throw new NexCoreError('ai_api_key not configured', -1);
        return ["Authorization: Bearer $k"];
    }

    public function chat(array $messages, string $model, array $extra = []): array
    {
        $body = array_merge(['model' => $model, 'messages' => $messages], $extra);
        return $this->client->request('POST', '/v1/chat/completions', ['headers' => $this->authHeaders(), 'body' => $body]);
    }
    public function models(): array
    {
        return $this->client->request('GET', '/v1/models', ['headers' => $this->authHeaders()]);
    }
}
