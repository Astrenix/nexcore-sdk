# Changelog

本仓库遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 和 [Semantic Versioning](https://semver.org/lang/zh-CN/).

## [3.3.0] - 2026-07-15

### 修复(Fixed)

- **Energy**:`getPrice` / `estimateEnergy` 硬编码 query 参数名与后端不符(`energy` → `energy_amount`、`receive_addr` → `to_address`),此前调用必返 400;四语言同步修正
- **Energy**:租期枚举修正 — 后端实际支持 `1H / 1D / 3D / 7D / 30D`,废除文档/注释中不存在的 `6H` / `1W`;`createOrder` 字段说明对齐后端(`receive_address` / `energy_amount` / `out_trade_no` / `remark`)
- **Payment**:`getUserAddress` 签名串多含 `trade_type` 导致调用必 401(参数移除见下方 Breaking)
- **Payment**:`createOrder` 签名未按后端规则归一 `amount` 两位小数、`timeout` 未显式传时漏入签名串,导致验签失败;现自动归一 amount 且 timeout 恒参与签名
- **Payment**:文档字段修正 — `user_id` → `out_user_id`、`expires_at` → `expired_at`、`call_type` 仅 `rotation` 模式必填
- **Withdraw**:`quoteFee` 的 `amount` 误标可选 — 后端必填,已改为必填参数
- **SMTP**:`sendBatch` 请求结构错误(`to[]` → `recipients[]`,元素 `{to, variables?, from_name?}`);`sendTemplate` 字段错误(`template_id` → `template_code`,移除 `account_id`,新增 `from_name`)
- **响应字段说明对齐后端**:SMTP `quota`(`daily_*` / `monthly_*` / `expire_at`)、SMTP `status`、Exchange `convert`(`{from, to, amount, result, rate, updated_at}`)、Exchange `getRate` 补 `inverse` 等一批与后端不符的字段说明

### 新增(Added)

- **SMTP 第 6 个 endpoint**:`POST /api/v1/smtp/inbound` — 退信/投诉上报(`reportInbound` / `report_inbound` / `ReportInbound`),自动把邮箱加入抑制名单并标记对应 send_log
- **SMTP 幂等与定时**:`send` / `sendBatch` 支持 `Idempotency-Key` 幂等头(同 key 重试直接返回首次结果,不重复发送/扣配额);`send` 支持 `send_at` 定时发送(RFC3339)
- **SMTP send 全量可选字段**:`from_name` / `reply_to` / `text_body` / `headers` / `cc` / `bcc` / `attachments`
- **Go**:新增签名语义测试(`payment_test.go`)

### 破坏性变更(Breaking)

- **Payment**:`getUserAddress(userId)` 去掉第二个 `tradeType` 参数(四语言同步);后端签名串不含 trade_type,旧版本该调用必 401
- **Exchange**:`getRates` 默认 `base` 不再由 SDK 侧写死 `CNY`;不传时由后端取默认(USDT)
- **SMTP**:`sendBatch` / `sendTemplate` 参数结构对齐后端(`recipients` 数组 / `template_code`),旧调用方式不兼容

### 变更(Changed)

- 各 SDK 覆盖 endpoint 数 43 → **44**(SMTP +1;注:v3.2 加入 account 2 + vcard 12 后实际已是 43,历史文档曾误写 29/25,本版一并修正)

## [3.2.0] - 2026-06-28

### 新增

- 4 个语言 SDK 全部新增 **`.account` namespace** — 账户(2 endpoint):
  - `GET /api/v1/account/balance` 账户 USD 余额(顺带 TRON 固定充值地址)
  - `GET /api/v1/account/deposit-address` 获取/分配 TRON 固定充值地址
- 4 个语言 SDK 全部新增 **`.vcard` namespace** — 虚拟信用卡(12 endpoint):
  - 读(X-API-Key + X-Secret-Key):`getInfo` / `listBins` / `listCards` / `getCardTransactions` / `listOrders` / `getOrder` / `updateCardRemark`
  - 资金/敏感(**HMAC 头签名**):`getCardDetails` / `getCardCode`(完整卡号·CVV·有效期 / 3DS 验证码)、`openCard` / `rechargeCard` / `cancelCard`
- 新增配置项 `apiKey` / `apiSecret`(MPK 商户密钥,Account 与 VCard 共用)
- **新增 header 式 HMAC 签名**:`X-Signature = HMAC-SHA256(apiSecret, timestamp + nonce + method + path + rawQuery + body)`,头 `X-Key-ID / X-Timestamp / X-Nonce / X-Signature`;POST 采用 body-raw 机制(签名串与实际字节一致)
- 新增各语言 **`verifyWebhook(params, secret)`** — 校验平台主动推送的虚拟卡 webhook 事件签名(开卡/充值/注销/新交易/新验证码),复刻后端 HMAC 验签 + 提示 sign_ts/nonce 防重放

## [3.1.0] - 2026-05-23

### 新增

- 4 个语言 SDK 全部新增 **`.withdraw` namespace** — 多链收款业务的资金出库端,共 4 个 endpoint:
  - `POST /api/v1/withdraw` 发起提币(`createWithdraw` / `create_withdraw`)
  - `GET  /api/v1/withdraw/:id` 查询单笔(`getWithdraw` / `get_withdraw`)
  - `GET  /api/v1/balance/withdrawable` 查询可提余额(`getWithdrawableBalance` / `get_withdrawable_balance`)
  - `GET  /api/v1/fee/quote` 费用预估(`quoteFee` / `quote_fee`)
- 提币 API 使用 **RSA-PKCS1v15-SHA256** 签名(非对称),与收款 API 的 HMAC-SHA256(对称)并存
- 各 SDK 配置项新增 `withdraw_api_key` / `withdraw_private_key_pem` / `withdraw_platform_public_key_pem`
- 新增 `Withdraw.sign()` 公开方法(便于 curl 调试场景手算签名)
- 新增 `Withdraw.verifyCallback()` 公开方法(对接方收到平台回调时校验签名)

### 变更

- 各 SDK 覆盖 endpoint 数从 25 增至 **29**(4 个新增提币 endpoint)
- `Http.request` 在 4 个语言中统一支持 raw body bytes/string 传入(RSA 签名场景必须用,保证签名串与实际 body 字节一致)

### 依赖

- Python SDK `pyproject.toml` 新增 `cryptography>=3.4.0`(运行时 lazy import,不用提币功能不会强制装)
- PHP SDK `composer.json` 新增 `ext-openssl: *`
- Go / Node 用各自语言标准库内置 crypto,无新增第三方依赖

### 文档

- 各 SDK `doc.go` / `__init__.py` / `client.js` / `Client.php` 顶部注释同步更新 namespace 列表
- `index.d.ts`(Node)新增 `WithdrawNamespace` / `WithdrawCreateParams` 类型

## [3.0.0] - 2026-05-23

### 新增

- 公开发布 **4 个语言全栈 SDK**:PHP / Python / Node.js / Go
- 业务命名空间:`.payment` / `.exchange` / `.energy` / `.smtp`
- 覆盖 **25 个 v1 公开 endpoint**,字段与 NexCore `/docs` 在线文档 100% 对齐
- 统一异常 `NexCoreError`(code / message / requestId / httpStatus)
- Webhook 签名常量时间校验(防时序攻击)

### 文件结构

各语言 SDK 按业务命名空间**拆分多文件**,而非单文件巨石:

- PHP — PSR-4 autoload + `src/Namespaces/{Payment,Exchange,Energy,Smtp}.php`
- Python — 完整 package `nexcore/namespaces/{payment,exchange,energy,smtp}.py` + `pyproject.toml`(pip 可装)
- Node.js — `src/namespaces/{payment,exchange,energy,smtp}.js` + 完整 TypeScript 类型 + `package.json`(npm 可发)
- Go — 每个业务一个 `.go` 文件(`payment.go` / `exchange.go` / `energy.go` / `smtp.go`)+ `doc.go`

### 包管理元信息

- PHP — `composer.json`(`composer require nexcore/sdk`)
- Python — `pyproject.toml`(`pip install nexcore-sdk`)
- Node.js — `package.json`(`npm install @nexcore/sdk`,含 repository / bugs / homepage)
- Go — `module github.com/DoBestone/nexcore-sdk/go`(`go get` 跟公开仓库地址对齐)

### 安全

- 全语言 Webhook 签名校验使用常量时间比较(hash_equals / hmac.compare_digest / crypto.timingSafeEqual / hmac.Equal)
- 公开仓库强 `.gitignore` 防 `.env` / `*.key` / credentials 误提交

### 文档

- 顶层 `README.md` 总览(业务/语言/统一设计/调用示例)
- 各语言独立 README + `examples/`(create_order / webhook 双示例,可直接 copy 粘贴)
- 每个 SDK 类 / 方法都有完整注释(phpDoc / docstring / JSDoc / godoc)
- `CONTRIBUTING.md` / `CODE_OF_CONDUCT.md` / `SECURITY.md`
- GitHub Actions CI(轻量语法检查,不打包 binary)
