package nexcore

import (
	"errors"
	"fmt"
)

// Error is the unified SDK error type.
//
// 所有 SDK 调用失败(网络错误 / HTTP 4xx-5xx / 业务 code != 0)统一返回 *Error.
// 业务方用 errors.As 或 nexcore.AsError 提取详细字段.
type Error struct {
	// Message 人类可读错误描述
	Message string
	// Code 平台错误码(0=成功;-1=客户端层错误;其他参见错误码表)
	Code int
	// RequestID 服务端日志追踪 ID,通过响应头 X-Trace-Id 透传.排查问题时给后端工单提供本值.
	RequestID string
	// HTTPStatus 实际 HTTP 状态码;客户端层错误时为 0.
	HTTPStatus int
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("nexcore: %s (code=%d, http=%d, trace=%s)",
		e.Message, e.Code, e.HTTPStatus, e.RequestID)
}

// AsError 把通用 error 转成 *Error,方便业务层拿 Code / RequestID 等字段.
// 不是 SDK 抛出的错误返回 nil.
//
//	if err := client.Payment.CreateOrder(...); err != nil {
//	    if ne := nexcore.AsError(err); ne != nil {
//	        log.Printf("Code=%d, Trace=%s", ne.Code, ne.RequestID)
//	    }
//	}
func AsError(err error) *Error {
	var ne *Error
	if errors.As(err, &ne) {
		return ne
	}
	return nil
}
