---
name: bohrium-tools
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr tools 浏览领域和科学工具、混合检索并读取详情。用于工具分类标识、唯一工具键、显式关键词权重、定价和适用条件核对。"
---

<a id="bohrium-scientific-tools"></a>

# Bohrium 科学工具库

已知领域时，优先免费浏览，再考虑付费查询：

```bash
bohr wiki tools domains -o json
bohr wiki tools subdomains --domain <domain_id> -o json
bohr wiki tools list --domain <domain_id> -o json
```

`domains`、`subdomains` 和 `list` 免费。领域 ID 是分类节点标识，不是工具标识。读取详情使用返回的唯一工具键。

按任务意图发现工具：

```bash
bohr billing pricing --resource tools -o json
bohr wiki tools search "molecular dynamics trajectory analysis" --k 20 --lang en-US --yes -o json
bohr wiki tools info <unique_tool_key> --yes -o json
```

混合检索每次 **0.01 CNY**，详情每次 **0.02 CNY**，优先使用 Wiki/Tool 共享的每月 1,000 次额度。即使额度可能使调用免费，确认门禁仍然适用。费用已获授权后才能添加 `--yes`。

运行时 notice 可能只给出剩余额度或支付渠道，而没有实际扣费金额。不能根据额度用量推算实际收费。

搜索可能根据 query 自动生成加权关键词。只有用户需要显式权重时，才将 `--keywords` 作为 JSON 对象提供。检索排名只是返回顺序，不是质量或兼容性保证。推荐执行工具前，根据返回来源核对仓库、许可证、安装要求和平台条件。

除上述规范命令外，当前 CLI 没有独立的 summary、tags、subdomain-detail 或 tools-under-subdomain 接口。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


完整命令、标识符交接、当前参数、价格与能力边界见 [命令参考](references/commands.md)。
