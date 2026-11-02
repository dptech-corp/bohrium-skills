---
name: bohrium-paper-search
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr paper search 和 bohr paper patent 检索论文或专利。用于文献发现、参考文献整理、专利检索、日期/JCR/数据库筛选和技术调研；人物画像使用 Scholar，论断与推理溯源使用 LKM。"
---

<a id="bohrium-paper-and-patent-search"></a>

# Bohrium 论文与专利检索

先选择记录类型：

- `bohr paper search <query>` 返回学术论文记录。
- `bohr paper patent <query>` 返回专利记录。

两者都是固定单次价格的操作。付费前核对具体档位：

```bash
bohr billing pricing --resource paper -o json
bohr billing pricing --resource patent -o json
```

论文普通版（`--type 0`）每次 **0.05 CNY**，加强版（`--type 1`）每次 **0.10 CNY**。专利普通版、加强版、专业版分别为 **0.10 CNY**、**0.30 CNY**、**0.50 CNY**，对应 `--type 0/1/2`。非交互调用只有在费用已获授权后才能添加 `--yes`。

使用明确的专业检索词，从满足任务所需的最低合适档位开始：

```bash
bohr paper search "solid electrolyte interphase lithium metal" \
  --year-from 2020 --year-to 2026 --jcr "Q1,Q2" --type 0 --yes -o json

bohr paper patent "solid-state battery dendrite suppression" \
  --type 0 --yes -o json
```

每次翻页都是新的付费调用。根据用户需要的证据量限定 `--page` 和 `--size`，不要在未获得额外费用授权时自动翻页。

题名、作者、DOI/出版标识、日期、期刊、摘要和专利字段均以返回记录为准。论文检索可能混合多种出版类型，需检查每条记录的类型和元数据，不能把每个结果都当作期刊论文。检索排序不等于证据质量。

学者画像使用 `bohr scholar search`；论断链和论文图谱使用 `bohr lkm`；综合解答使用 Mentor 或文献综述工作流。

计费 notice 可能只提供渠道或额度信息，而没有实际扣费金额。只报告响应明确给出的金额。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


全部过滤条件、档位、价格、结果边界和安全翻页规则见 [命令参考](references/commands.md)。
