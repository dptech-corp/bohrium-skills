---
name: bohrium-tools
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr wiki tools 浏览领域和科学工具、混合检索并读取详情。用于工具分类标识、唯一工具键、显式关键词权重、定价和适用条件核对。"
---

<a id="bohrium-scientific-tools"></a>

# Bohrium 科学工具库

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 了解可选科学领域 | 免费浏览分类；返回的 node ID 是分类节点 |
| 为具体任务找软件 | 付费混合检索 → 选定工具键 → 付费详情 |
| 推荐可用工具 | 查看详情和来源 → 核对许可证、平台和安装要求 |

## 工作流：为分析任务选择工具

1. 明确输入格式、科学操作、运行环境及平台限制，以及获授权的检索和详情预算。分类浏览可帮助缩小检索词，但不会返回可执行工具记录。
2. 搜索任务，再查看一个有希望的候选：

```bash
bohr wiki tools search "molecular dynamics trajectory analysis" --k 10 --lang en-US --yes -o json
bohr wiki tools info <tool_unique_key> --yes -o json
```

3. 从选中的搜索记录 `data.tools[].tool_unique_key` 填入 `<tool_unique_key>`。分类发现使用 `data.items[].node_id`，该节点 ID 不能作为工具键。搜索没有 `--domain` 过滤参数；将领域写进查询，不要编造参数。
4. 查看返回的能力、仓库、教程和使用细节，核对候选能否处理用户输入，以及是否满足目标平台和许可证条件。返回命令属于参考资料，仅在任务范围包含执行或安装时才执行。
5. 交付有理由的选择、来源链接、限制和下一步可执行动作。命中排名只完成发现，还未完成兼容性核验。详情缺失或无结果时保留限制；再查候选或详情需要已有预算余量。

已知领域时，优先免费浏览，再考虑付费查询：

```bash
bohr wiki tools domains -o json
bohr wiki tools subdomains --domain <domain_id> -o json
bohr wiki tools list --domain <domain_id> -o json
```

`domains`、`subdomains` 和 `list` 用于免费发现分类。`list --domain` 返回子分类记录，不是工具；它们的 node ID 是分类标识。先从混合检索取得唯一工具键，再读取工具详情。

按任务意图发现工具：

```bash
bohr billing pricing --resource tools -o json
bohr wiki tools search "molecular dynamics trajectory analysis" --k 20 --lang en-US --yes -o json
bohr wiki tools info <unique_tool_key> --yes -o json
```

混合检索每次 **0.01 CNY**，详情每次 **0.02 CNY**，优先使用 Wiki/Tool 共享的每月 1,000 次额度。即使额度可能使调用免费，确认门禁仍然适用。费用已获授权后才能添加 `--yes`。

运行时 notice 可能只给出剩余额度或支付渠道，而没有实际扣费金额。不能根据额度用量推算实际收费。

省略 `--keywords` 时，CLI 将每个位置参数映射为权重 `1.0`；带引号的多词查询是一个完整关键词短语，不会让服务端推导权重。需要显式权重时，用 JSON 对象提供 `--keywords`。检索排名只是返回顺序，不是质量或兼容性保证。推荐执行工具前，根据返回来源核对仓库、许可证、安装要求和平台条件。

除上述规范命令外，当前 CLI 没有独立的 summary、tags、subdomain-detail 或 tools-under-subdomain 接口。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


完整命令、标识符交接、当前参数、价格与能力边界见 [命令参考](references/commands.md)。
