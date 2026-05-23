# Changelog

本仓库遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 和 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [3.0.0] - 2026-05-23

### 新增

- 初始公开发布,**4 个语言全栈覆盖**:PHP / Python / Node.js / Go
- 统一全能 `Client` API,业务命名空间:`.payment` / `.energy` / `.smtp` / `.ai`
- 统一异常 `NexCoreError`(code / message / requestId / httpStatus)
- Payment 模块:create/query/close order、bind/get/unbind address、appConfig、webhook 签名校验
- Energy 模块:info、price、estimateEnergy、createOrder、queryOrder、listOrders
- SMTP 模块:sendMail、listAccounts、listTemplates
- AI 模块:chat、models(OpenAI 兼容协议)
- 移动端友好:Node 零依赖、Go 零依赖、Python 仅 `requests`、PHP 仅 ext-curl

### 安全

- Webhook 签名校验全语言使用常量时间比较(HMAC.equal / timing_safe_eq / timingSafeEqual)
- 公开仓库强 `.gitignore` 防 `.env` / `.key` / credentials 误提交

### 文档

- 顶层 `README.md` 总览 + 各语言独立 README
- `CONTRIBUTING.md` / `CODE_OF_CONDUCT.md` / `SECURITY.md`
- GitHub Actions CI(4 个语言矩阵)
