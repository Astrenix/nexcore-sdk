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
 * 字段访问(继承 \Exception):
 *   - $e->getCode()    平台错误码(0=成功;-1=客户端层错误)— int
 *   - $e->getMessage() 错误描述 — string
 *   - $e->requestId    服务端 X-Trace-Id — ?string
 *   - $e->httpStatus   实际 HTTP 状态码 — ?int
 *
 * 注:`code` / `message` 通过父类 getter 访问,不重声明字段避免 PHP 严格类型冲突.
 */
class NexCoreError extends \RuntimeException
{
    public ?string $requestId;
    public ?int $httpStatus;

    public function __construct(string $message, int $code = 0, ?string $requestId = null, ?int $httpStatus = null)
    {
        parent::__construct($message, $code);
        $this->requestId  = $requestId;
        $this->httpStatus = $httpStatus;
    }

    /**
     * 兼容老代码 $e->code 访问(部分 examples 直接读字段).
     * 推荐使用 $e->getCode() 标准 PHP 风格.
     */
    public function __get(string $name)
    {
        if ($name === 'code') {
            return $this->getCode();
        }
        if ($name === 'message') {
            return $this->getMessage();
        }
        return null;
    }
}
