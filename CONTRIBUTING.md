# Contributing to Tovanix SDK

谢谢你想为 Tovanix SDK(原 NexCore SDK)贡献!以下是参与流程。

## 仓库结构

```
sdk/
├── php/      # PHP 7.4+
├── python/   # Python 3.8+
├── node/     # Node.js 16+(含 TypeScript 类型)
├── go/       # Go 1.21+
├── README.md
├── LICENSE   (MIT)
└── .github/workflows/   CI
```

每个语言子目录是**独立可发布**的包,统一遵守:
- 全能 `Client` 主类
- 业务 namespace:`.payment` / `.energy` / `.smtp` / `.ai`
- 错误统一 `NexCoreError`(包含 `code` / `message` / `requestId` / `httpStatus`)

## 提 PR 前自检

1. **代码风格** — 各语言遵循官方风格(Go fmt / Python PEP 8 / JS Prettier)
2. **零依赖优先** — 每个 SDK 必须能用语言标准库或最少依赖运行
3. **新增字段不破 API** — 后端 API 返回字段可能变,SDK 不要强映射成 struct(Go 用 `json.RawMessage`、Python/Node 返回 dict / object)
4. **示例同步** — 新增方法时更新 `examples/`
5. **测试** — CI 跑通(后续会加单元测试)

## 报 Bug

[GitHub Issues](https://github.com/nexcore-platform/sdk/issues) 提交,模板:

```markdown
**SDK 语言版本**:Python 3.11 / nexcore-sdk-py 3.0.1
**Tovanix API 版本**:v3.0
**问题描述**:...
**最小复现代码**:```python
...
```
**期望行为**:...
**实际行为**(含 X-Trace-Id):...
```

## 安全漏洞

请**不要**在公开 issue 报告安全漏洞。发邮件到 `security@tovanix.com` 或通过 Tovanix 用户后台「工单」选择「安全」类型。

## 许可

提交即表示你同意作品在 [MIT License](./LICENSE) 下发布。
