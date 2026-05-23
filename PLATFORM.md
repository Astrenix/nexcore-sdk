# NexCore Platform — 业务全景

> 本文档介绍 **NexCore** 综合数字基础服务平台的完整业务版图.
> SDK 当前覆盖核心 4 个,其余 5 个业务通过 [线上文档](https://nexcores.net/docs) + Web 控制台使用.

## 平台定位

**NexCore** 是面向开发者与中小团队的**综合数字基础服务平台**,把跨境业务全链路所需的高频但难自研能力封装为 API/控制台,让一个开发者也能搭起完整的跨境业务系统.

跟其他垂直工具的差异:

|  | 垂直工具(各做一摊) | **NexCore** |
|---|---|---|
| 接入成本 | 每个业务一套凭据 / 文档 / SDK | 一套凭据 / 一套 SDK / 一处账单 |
| 计费 | 多家分散月底对账 | 平台统一余额,业务间共用 |
| 数据 | 跨平台手动拉数据 | 一处控制台看全业务 |
| 工单 | N 个客服群 | 一个工单系统 |

**目标用户**:跨境电商 / 海外广告投流团队 / AI 应用开发者 / 出海 SaaS / DApp/Web3 应用 / 内容创作者.

## 9 大业务模块

### 1. 多链收款 Payment 💰

USDT / USDC / TRX / BTC / ETH 等 **6 主链加密货币收款网关**.

- **轮播模式** — 平台动态分配地址,适合一次性付款 / 充值
- **一对一模式** — 用户绑定固定地址,所有收入主动推送 webhook,适合长期收款
- **秒级确认** — 链上 1 个 block 即推送
- **商户自托管** — 平台不保管资金,链上直接到商户钱包

✅ SDK 已覆盖 7 个 endpoint(`client.payment.*`).

### 2. 汇率服务 Exchange 💱

实时加密 ↔ 法币 / 法币 ↔ 法币 汇率服务,Payment 配套但独立可用.

- **GET /api/v1/rate** — 单对币种汇率
- **POST /api/v1/convert** — 金额换算
- **GET /api/v1/rates** — 批量币种 → 基准币
- **GET /api/v1/rates/fiat** — 法币间汇率
- **GET /api/v1/rates/all** — 全币种快照

✅ SDK 已覆盖 5 个 endpoint(`client.exchange.*`).

### 3. TRON 能量租赁 Energy ⚡

TRC20 转账省 **60%+ gas**,即租即用,30 秒到账.

- **常规订单** — 1H / 6H / 1D / 3D / 1W 5 档周期
- **一次性订单** — 用完即丢,适合单笔转账
- **主动回收** — 不用时手动归还能量
- **完整行情** — 平台公开报价 / 阶梯定价 / 估算所需能量

✅ SDK 已覆盖 8 个 endpoint(`client.energy.*`).

### 4. SMTP 聚合 API SMTP 📧

邮件聚合发送服务,多账户智能轮发 + 打开/点击全跟踪.

- **单封发送** — `send`,普通验证码 / 通知邮件
- **批量发送** — `sendBatch`,同主题多收件人
- **模板渲染** — `sendTemplate`,平台后台维护模板,运行时变量替换
- **配额查询** — `getQuota`,实时看本期用量
- **状态查询** — `getStatus`,投递 / 打开 / 点击 / 退订全跟踪

✅ SDK 已覆盖 5 个 endpoint(`client.smtp.*`).

### 5. 多链闪兑 Swap 🔄

任意币 ↔ 任意币,链上充提,30 分钟自动到账.

- **同链闪兑** — USDT-TRC20 → TRX-TRON 等
- **跨链兑换** — USDT-TRC20 → ETH-Ethereum 等
- **最优路径** — 平台自动找最优兑换链路
- **免 KYC 免开户** — 链上充提模式

📋 当前通过 [Web 控制台](https://nexcores.net/m/swap) 使用,SDK 后续按需求扩展.

### 6. 虚拟信用卡 Vcard 💳

USDT 充值开卡,海外广告 / AI 订阅秒结算,**Visa / Mastercard 全球可用**.

- **多 BIN 可选** — 不同卡头不同接受率
- **3DS 实时可查** — 验证码 API 实时获取
- **跨境支付** — 200+ 国家可用
- **典型场景** — Google/Meta 广告投流、ChatGPT/Claude 订阅、Netflix/Spotify、海外 SaaS

📋 当前通过 [Web 控制台](https://nexcores.net/m/vcard) 使用,SDK 后续按需求扩展.

### 7. 云服务 Cloud ☁️

域名 / 服务器 / DNS / SSL **一站式**,跟主流云厂商整合.

- **域名注册** — 30+ 顶级后缀
- **DNS 管理** — A / CNAME / TXT / MX 等
- **SSL 证书** — Let's Encrypt 自动签发
- **CDN** — Cloudflare 集成,一键启用
- **轻量 / 云服务器** — 全场景按需
- **域名邮件路由** — 把自定义域名邮件转发到现有邮箱

📋 当前通过 [Web 控制台](https://nexcores.net/m/cloudservice) 使用.

### 8. SMS 接码 + 专用邮箱 SMS 📱

海外平台注册首选 — **60+ 国家真实运营商号源** + 专用收件邮箱.

- **接码 SMS** — 一次性 / 长期号码
- **专用邮箱** — 永久邮箱地址 + 零隐私收件
- **典型场景** — ChatGPT / Claude / Gemini / Telegram / WhatsApp / Google / Twitter 等海外平台注册

📋 当前通过 [Web 控制台](https://nexcores.net/m/sms) 使用.

### 9. Astrenix AI 🤖

LLM 全代理 — **Claude / OpenAI / Gemini / 通义** 等主流模型,**完全兼容 OpenAI SDK**.

- **统一 endpoint** — `/v1/chat/completions`,跟 OpenAI 协议 100% 一致
- **统一计费** — 一个余额覆盖所有上游模型
- **官方 SDK 兼容** — 直接 `pip install openai` / `npm install openai`,修改 `base_url` 即可
- **多 Key 加权轮询** — 平台后台配置多组上游 Key,自动负载均衡
- **分组故障转移** — 单组上游故障自动切换

📋 不在本 SDK 内,**直接用 OpenAI 官方 SDK 改 base_url** 即可.详见 [`/docs?module=aiapi`](https://nexcores.net/docs?module=aiapi).

```python
# 直接用 OpenAI SDK 调 Astrenix AI
from openai import OpenAI
client = OpenAI(
    base_url="https://your-domain.com/v1",
    api_key="sk-nc-xxx",
)
reply = client.chat.completions.create(
    model="claude-opus-4-7",
    messages=[{"role": "user", "content": "Hello"}]
)
```

## 域名与品牌

| | URL |
|---|---|
| 主站 | https://nexcores.net |
| API 文档 | https://nexcores.net/docs |
| 用户后台 | https://nexcores.net/m |
| API 网关 | https://api.nexcore.io |
| 邮箱 | `support@nexcores.net` / `security@nexcores.net` |
| 易记品牌 | https://9188.pro |

## 历史与定位

NexCore 自 **2021 年 7 月** 启动,从最初的多链收款单业务出发,5 年间扩展为 **9 个业务模块** 的综合开发者平台.

服务过的典型客户:

- 跨境电商收单 + 加密结算
- 海外广告投流团队批量开卡
- 出海 AI 应用接 LLM + 海外通讯
- DApp / Web3 项目方支付 + 能量
- 内容创作者 ChatGPT / Claude 订阅 + 域名 + 主机

## 反馈

- [GitHub Issues](https://github.com/DoBestone/nexcore-sdk/issues) — 公开 Bug / 需求讨论
- NexCore 用户后台「工单」 — 私下技术支持
- `business@nexcores.net` — 商务合作 / 大客户对接
