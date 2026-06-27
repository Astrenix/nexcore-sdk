# Changelog

本仓库遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 和 [Semantic Versioning](https://semver.org/lang/zh-CN/).

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
