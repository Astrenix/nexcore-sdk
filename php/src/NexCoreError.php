<?php
/**
 * NexCore SDK 统一异常类型.
 *
 * 所有 SDK 调用失败(网络错误 / HTTP 4xx/5xx / 业务 code != 0)统一抛出本异常.
 * 业务方通过 catch 后访问字段可定位问题.
 */

declare(strict_types=1);

namespace NexCore;

/**
 * NexCoreError 是 SDK 全局异常.
 *
 * 触发场景:
 *   - 网络层失败(连接拒绝 / 超时 / DNS) — code=-1
 *   - HTTP 状态 >= 400(网关错 / 5xx / 4xx 业务限流等) — code 取响应中的 code 字段
 *   - 响应 JSON 解析失败 — code=-1
 *   - 业务层返回 { code: 非 0 } — code 取该值,message 取响应 message
 *
 * 字段:
 *   @property int      $code        平台错误码(0=成功;-1=客户端层错误;其他参见错误码表)
 *   @property string   $message     人类可读错误描述(基类 getMessage())
 *   @property string|null $requestId 服务端日志追踪 ID,通过响应头 X-Trace-Id 透传,用于排查问题时定位日志
 *   @property int|null    $httpStatus 实际 HTTP 状态码;客户端层错误时为 null
 */
class NexCoreError extends \RuntimeException
{
    public int $code;
    public ?string $requestId;
    public ?int $httpStatus;

    public function __construct(string $message, int $code = 0, ?string $requestId = null, ?int $httpStatus = null)
    {
        parent::__construct($message);
        $this->code       = $code;
        $this->requestId  = $requestId;
        $this->httpStatus = $httpStatus;
    }
}
