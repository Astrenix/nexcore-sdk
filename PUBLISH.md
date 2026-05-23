# 发布到公开仓库 — 操作指南

本指南给 NexCore 维护者一步步发布 SDK 到独立公开 GitHub 仓库。

## 一、独立仓库 vs 主仓库 subtree

SDK 目录(`sdk/`)目前是 NexCore 主仓库的一部分,但**应该单独建公开仓库**便于:

- 用户 `composer require` / `pip install` / `npm install` / `go get`
- CI 独立,跑得更快
- Issue 隔离,不污染主仓库
- LICENSE 范围清晰

## 二、新建公开仓库(一次性)

**仓库建议名:** `nexcore-sdk`(简短 + 明确)
**Owner**: `nexcore-platform`(组织;若个人则 `your-handle`)
**Visibility**: Public
**License**: MIT(已在 `sdk/LICENSE` 准备好)
**Description**: `Official multi-language SDK for NexCore platform — Payment / Energy / SMTP / AI APIs.`

### 步骤 A:GitHub 网页建仓

1. 登录 [github.com](https://github.com) → New repository
2. 填名称 / 描述 / Public / 不勾任何 init 选项(README/LICENSE/.gitignore 我们自己带)
3. 创建

### 步骤 B:本地推送(SDK 仅目录,不带主仓库历史)

```bash
# 进 SDK 目录,用 git filter-repo 抽出来,或直接初始化新历史
cd /path/to/NexCorePay/sdk
git init
git add .
git commit -m "feat: initial release v3.0 — PHP/Python/Node/Go all-in-one SDK"
git branch -M main
git remote add origin git@github.com:nexcore-platform/nexcore-sdk.git
git push -u origin main
```

### 步骤 C:打第一个 Release tag

```bash
git tag -a v3.0.0 -m "Initial public release"
git push origin v3.0.0
```

GitHub 上 Releases → Draft new release → 选 `v3.0.0` tag → 标题 `v3.0.0 — Initial public release` → body 引用 `CHANGELOG.md` 内容。

## 三、各语言包发布(可选,后续做)

| 语言 | 注册中心 | 命令 |
|---|---|---|
| PHP | [Packagist](https://packagist.org/) | 提交 git url 到 packagist + tag 触发自动更新 |
| Python | [PyPI](https://pypi.org/) | 加 `pyproject.toml`,`pip install build twine && python -m build && twine upload dist/*` |
| Node | [npm](https://www.npmjs.com/) | `cd node && npm publish --access public` |
| Go | 无中心化注册 | tag 推到 GitHub 即可,`go get github.com/nexcore-platform/nexcore-sdk-go@v3.0.0` |

## 四、上线后

1. 主仓库 `README.md` 加链接:`https://github.com/nexcore-platform/nexcore-sdk`
2. `/docs` API 文档页 footer 加 SDK 仓库链接
3. NexCore 用户后台 / 工单可关联 SDK GitHub Issues
4. 用户主控制台「API 文档」入口可加 SDK 下载入口

## 五、注意

- **公开仓库严防 secrets!** `.gitignore` 已防 `.env` / `*.key`,但每次 commit 前 `git diff` 自查
- 仓库根目录的 `LICENSE` / `README.md` / `CHANGELOG.md` / `CONTRIBUTING.md` / `CODE_OF_CONDUCT.md` / `SECURITY.md` 已经准备好,直接 push 即可
- `.github/workflows/ci.yml` 已经准备好,4 个语言矩阵测试,push 后自动跑
