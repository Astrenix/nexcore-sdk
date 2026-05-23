# Changelog

本仓库遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 和 [Semantic Versioning](https://semver.org/lang/zh-CN/).

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
