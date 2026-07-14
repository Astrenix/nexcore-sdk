<?php
/**
 * Tovanix(原 NexCore)Official PHP SDK 主客户端.
 *
 * 一次配置覆盖 Tovanix 平台全部 v1 公开接口,业务按 namespace 划分:
 *
 *   $client->payment   — 多链收款(HMAC-SHA256 签名)
 *   $client->exchange  — 汇率(X-App-Key + X-App-Secret header)
 *   $client->energy    — TRON 能量租赁(X-API-Key + X-Secret-Key)
 *   $client->smtp      — SMTP 聚合(Bearer Token)
 *   $client->account   — 账户余额 / 充值地址(MPK 双密钥)
 *   $client->vcard     — 虚拟信用卡(MPK 双密钥 + 敏感操作 HMAC 签名)
 *
 * 用法:
 *
 *   require_once __DIR__ . '/vendor/autoload.php';  // composer autoload
 *   use NexCore\Client;
 *
 *   $client = new Client([
 *       'base_url'          => 'https://your-domain.com',
 *       'payment_app_id'    => 'APP20260412XXXX',
 *       'payment_app_key'   => 'your_app_key_here',
 *       'energy_api_key'    => 'energy_key',
 *       'energy_secret_key' => 'energy_secret',
 *       'smtp_api_key'      => 'smk_xxx',
 *       'timeout'           => 30,
 *   ]);
 *
 *   $order = $client->payment->createOrder([...]);
 *
 * 所有错误统一抛 NexCore\NexCoreError(含 code / requestId / httpStatus).
 */

declare(strict_types=1);

namespace NexCore;

use NexCore\Namespaces\Payment;
use NexCore\Namespaces\Exchange;
use NexCore\Namespaces\Energy;
use NexCore\Namespaces\Smtp;
use NexCore\Namespaces\Withdraw;
use NexCore\Namespaces\Account;
use NexCore\Namespaces\VCard;

/**
 * Client 是 Tovanix SDK 的入口.
 *
 * 内部组合 Http 传输层 + 4 个业务 namespace.
 * 构造时把整个 config 数组保存,各 namespace 通过 get() 取需要的字段.
 */
class Client
{
    /** SDK 版本号(跟主仓库 v3.x.x 同步) */
    public const VERSION = '3.3.0';

    /** @var array<string, mixed> 完整配置 */
    private array $config;

    /** @var Http 底层 HTTP 传输 */
    public Http $http;

    /** @var Payment 多链收款命名空间 */
    public Payment $payment;

    /** @var Exchange 汇率命名空间 */
    public Exchange $exchange;

    /** @var Energy TRON 能量租赁命名空间 */
    public Energy $energy;

    /** @var Smtp SMTP 聚合命名空间 */
    public Smtp $smtp;

    /** @var Withdraw 多链收款 · 提币端命名空间(RSA-2048 签名) */
    public Withdraw $withdraw;

    /** @var Account 账户命名空间(MPK 双密钥) */
    public Account $account;

    /** @var VCard 虚拟信用卡命名空间(MPK 双密钥 + HMAC 签名) */
    public VCard $vcard;

    /**
     * @param array{
     *     base_url: string,
     *     payment_app_id?: string,
     *     payment_app_key?: string,
     *     energy_api_key?: string,
     *     energy_secret_key?: string,
     *     smtp_api_key?: string,
     *     withdraw_api_key?: string,
     *     withdraw_private_key_pem?: string,
     *     withdraw_platform_public_key_pem?: string,
     *     api_key?: string,
     *     api_secret?: string,
     *     timeout?: int,
     *     verify_ssl?: bool,
     *     user_agent?: string,
     * } $config
     */
    public function __construct(array $config)
    {
        if (empty($config['base_url'])) {
            throw new NexCoreError('base_url is required', -1);
        }
        $this->config = array_merge([
            'timeout'    => 30,
            'verify_ssl' => true,
            'user_agent' => 'NexCore-PHP-SDK/' . self::VERSION,
        ], $config);

        $this->http     = new Http(
            $this->config['base_url'],
            (int) $this->config['timeout'],
            (bool) $this->config['verify_ssl'],
            (string) $this->config['user_agent']
        );

        $this->payment  = new Payment($this);
        $this->exchange = new Exchange($this);
        $this->energy   = new Energy($this);
        $this->smtp     = new Smtp($this);
        $this->withdraw = new Withdraw($this);
        $this->account  = new Account($this);
        $this->vcard    = new VCard($this);
    }

    /**
     * 取配置字段.各 namespace 用来获取自身需要的 API key / secret.
     *
     * @param string $key 配置字段名,如 'payment_app_key'
     * @return mixed|null
     */
    public function get(string $key)
    {
        return $this->config[$key] ?? null;
    }
}
