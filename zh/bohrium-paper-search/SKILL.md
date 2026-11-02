---
name: bohrium-paper-search
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr paper search 和 bohr paper patent 检索论文或专利。用于文献发现、参考文献整理、专利检索、日期/JCR/数据库筛选和技术调研；人物画像使用 Scholar，论断与推理溯源使用 LKM。"
---

<a id="bohrium-paper-and-patent-search"></a>

# Bohrium 论文与专利检索

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 整理限定范围的文献表 | 论文搜索 → 筛选记录类型 → 去重 → 证据表 |
| 找专利证据 | 专利搜索 → 查看实际返回的公开信息、申请人和法律元数据 |
| 读取版面或科学推理 | 选中文件或 PDF URL → PDF Parser；受支持的论文标识 → LKM |

## 工作流：检索、筛选，再交接证据

1. 明确问题、纳入条件、年份、记录数量，以及获授权的付费调用次数和档位。执行前遵循下方价格和支付规则。
2. 显式限制范围，先检索一次：

```bash
bohr paper search "solid electrolyte interphase lithium metal"   --year-from 2020 --year-to 2026 --page 1 --size 10 --type 0 --yes -o json
```

3. 先查看实际记录，不假定固定结果数组。保留已有的标题、作者、年份、来源、DOI 或出版标识、摘要、记录类型和 URL。优先按 DOI 或标识去重；回退到标题和年份时标明不确定性。
4. 筛选相关性和来源类型，说明排除项和缺失证据。交付带出处的表格及检索条件。这是限定范围的检索，不代表穷尽覆盖；适用下方正式版本的翻页限制。
5. 提取全文前，先确定可访问的实际 PDF 文件或直接 PDF URL，再交给 PDF Parser。需要 LKM 推理时，按其参考使用 DOI、标题或已确认的数字论文 ID。搜索记录 ID 和落地页 URL 不会自动变成解析 token 或 LKM ID。
6. 无结果时只在预算内调整查询。付费调用被拒绝或结果未知时，先查看错误和计费证据；换档位、支付来源或重复调用都属于新的支出决策。

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

已发布的 bohr 2.9.19 虽然接受 `--page`，但论文和专利请求都不下发页码。递增页码只是使用相同查询参数再次检索，不是读取下一页，每次调用仍可能收费。保持 `--page 1`，用 `--size` 限定所需证据量；只能在已授权预算内调整检索词或筛选条件，并明确报告该版本不支持翻页，不能声称扩大了页码覆盖。

题名、作者、DOI/出版标识、日期、期刊、摘要和专利字段均以返回记录为准。论文检索可能混合多种出版类型，需检查每条记录的类型和元数据，不能把每个结果都当作期刊论文。检索排序不等于证据质量。

学者画像使用 `bohr scholar search`；论断链和论文图谱使用 `bohr lkm`；综合解答使用 Mentor 或文献综述工作流。

计费 notice 可能只提供渠道或额度信息，而没有实际扣费金额。只报告响应明确给出的金额。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


全部过滤条件、档位、价格、结果量限制和不支持翻页的边界见 [命令参考](references/commands.md)。
