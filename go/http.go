package nexcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// httpTransport 封装底层 HTTP 调用.
// 各业务命名空间通过 Client.transport.do() 发请求,不直接接触 net/http.
type httpTransport struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// requestOpts 是 do() 的可选参数.
type requestOpts struct {
	Body    any                    // 自动 JSON 序列化为 body
	Query   map[string]any         // query 参数(自动 url-encode + 过滤空值)
	Headers map[string]string      // 额外 header
}

// do sends an HTTP request and returns the response data segment (envelope unwrapped).
//
// 自动处理:
//   - URL + query 拼接
//   - JSON body 编码
//   - 公共 header 注入(Content-Type / Accept / User-Agent)
//   - 响应解包(自动从 {code, message, data} envelope 取 data)
//   - 错误统一返回 *Error(含 X-Trace-Id)
//
// 业务方拿到 json.RawMessage 后自行 Unmarshal 到具体 struct.
func (t *httpTransport) do(method, path string, opts *requestOpts) (json.RawMessage, error) {
	if opts == nil {
		opts = &requestOpts{}
	}

	urlStr := t.baseURL + path

	// query 拼接
	if len(opts.Query) > 0 {
		q := url.Values{}
		for k, v := range opts.Query {
			s := fmt.Sprint(v)
			if s == "" || s == "<nil>" {
				continue
			}
			q.Set(k, s)
		}
		if enc := q.Encode(); enc != "" {
			if strings.Contains(urlStr, "?") {
				urlStr += "&" + enc
			} else {
				urlStr += "?" + enc
			}
		}
	}

	// body 编码
	var body io.Reader
	if opts.Body != nil {
		b, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, &Error{Message: "marshal body: " + err.Error(), Code: -1}
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(strings.ToUpper(method), urlStr, body)
	if err != nil {
		return nil, &Error{Message: "build request: " + err.Error(), Code: -1}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", t.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Message: "http: " + err.Error(), Code: -1}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	traceID := resp.Header.Get("X-Trace-Id")

	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if jsonErr := json.Unmarshal(raw, &env); jsonErr != nil {
		return nil, &Error{
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200)),
			Code:       -1,
			RequestID:  traceID,
			HTTPStatus: resp.StatusCode,
		}
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		return nil, &Error{
			Message:    env.Message,
			Code:       env.Code,
			RequestID:  traceID,
			HTTPStatus: resp.StatusCode,
		}
	}
	if len(env.Data) > 0 {
		return env.Data, nil
	}
	return raw, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
