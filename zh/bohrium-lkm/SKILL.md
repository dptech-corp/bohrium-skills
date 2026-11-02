---
name: bohrium-lkm
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr lkm 检索科学论断、问题、推理链、证据和论文图谱，或从本地 PDF 异步抽取知识。用于论文标识、全局变量、论断溯源和抽取任务生命周期；版面解析使用 bohr pdf parse。"
---

<a id="bohrium-lkm"></a>

# Bohrium 大知识模型

LKM（Large Knowledge Model，大知识模型）将论文转化为可溯源的推理图谱：问题、论断、推理链和证据成为可单独寻址、可跨论文对齐的单元。

通过与版本匹配的 `bohr lkm` 命令操作，先选择路径：

- `search` 在已有语料中检索论断、问题、摘要和科学概念。
- `reasoning` 检索完整论证或实验过程。
- `graph` 获取已知论文的 `nodes` / `edges` 图谱。
- `reference` 列出已知论文的参考文献或被引集合。
- `claim reasoning` 追踪已知结论论断的推理链。
- `variables batch` 补全已知全局节点 ID 的内容。
- `extract` 异步上传并抽取可能尚未收录的本地 PDF。

一般论文发现使用 `bohr paper search`；个人知识库管理使用 `bohr kb`；Uni-Parser 的布局、文本、表格和公式提取使用 `bohr pdf parse`。

<a id="authentication-output-and-billing"></a>

## 身份、输出和计费

使用现有 Bohrium 登录态或 AccessKey，优先结构化输出：

```bash
bohr auth whoami -o json
bohr billing pricing --resource lkm -o json
```

`search`、`reasoning`、`graph`、`reference`、`variables batch`、`claim reasoning` 的固定单次价格均为 **0.05 CNY**，优先使用 LKM 个人每月 1,000 次额度。即使剩余额度可能使调用免费，确认门禁仍然适用。非交互调用只有费用已获授权后才能添加 `--yes`，否则报告 `CONFIRMATION_REQUIRED`。

Extract 与每月查询额度分开计费：成功解析每篇 **1.00 CNY**，命中缓存每篇 **0.10 CNY**，同一任务不重复收费。非交互调用同样在费用已获授权后才添加 `--yes`。

真实提交会把所选 PDF 上传到 Bohrium，提交前确认具体文件和用户意图。

<a id="search-and-trace-the-ingested-corpus"></a>

## 检索并追踪已有语料

不知道标识符时从自然语言开始：

```bash
bohr lkm search "perovskite thermal stability" \
  --mode hybrid \
  --sort comprehensive \
  --scopes conclusion,abstract \
  --top-k 10 \
  --yes \
  -o json

bohr lkm reasoning \
  --query "infer phase stability from XRD evidence" \
  --top-k 10 \
  --yes \
  -o json
```

`search` 返回代表性 `variables`、同一论文的 `related` 上下文及论文权威元数据；`reasoning` 返回完整推理链命中。`score` 和 `rerank_score` 只表示排序信号，不代表置信度、真实概率或证据强度。

从结构化结果获取标识符，不要编造 ID：

- 数字论文 ID 用于 `bohr lkm graph --paper-id` 或 `bohr lkm reference --paper-id`。
- DOI 用于 `bohr lkm graph --doi` 或 `bohr lkm reference --doi`。
- `paper:<digits>` 包键用于 `bohr lkm graph --package-id`。
- 题名用于 `bohr lkm graph --title`，可能返回多个图谱。
- 全局 `gcn_...` ID 用于 `variables batch`；只有带推理的结论论断才交给 `claim reasoning`。
- 图内局部节点 ID 只在单个返回图谱内有效，不是全局变量、论断或论文 ID。

```bash
bohr lkm graph --paper-id <numeric_paper_id> --yes -o json
bohr lkm graph --doi <doi> --yes -o json
bohr lkm graph --title "<paper_title>" --yes -o json
bohr lkm reference --paper-id <numeric_paper_id> --yes -o json
bohr lkm variables batch <global_gcn_id> [global_gcn_id...] --yes -o json
bohr lkm claim reasoning <global_conclusion_gcn_id> --format graph --yes -o json
```

解释来源关系时使用图谱方向和节点类型。`highlight` 是正向支持，`weak_point` 是局限，`subproblem` 是研究问题或动机，不能与最终 `conclusion` 混用。

<a id="extract-knowledge-from-a-local-pdf"></a>

## 从本地 PDF 抽取知识

用户持有本地 PDF 而非已有语料结果时，使用 `extract`：

```bash
bohr lkm extract submit paper.pdf -o json
bohr lkm extract wait <task_id> --interval 5s --timeout 30m -o json
bohr lkm extract result <task_id> --yes -o json
bohr lkm extract result <task_id> --format graph --yes -o json
```

每篇成功抽取 **1.00 CNY**，缓存命中 **0.10 CNY**；真实提交前确认费用。

输入必须是非空的普通 PDF 文件，最大 64 MiB，服务另有 50 页限制。提交返回 extract task ID，可能返回 `cache_hit` 表示复用已有结果。不要为了加速重复提交正在处理的 PDF；服务返回 `290020` 及现有 task ID 时，继续对该任务使用 `status` 或 `wait`。

`queued` 和 `running` 为非终态；`succeeded`、`partial`、`failed` 为终态。`partial` 是无法生成完整图谱的非重试业务终态，同一 PDF 不应重新提交。`failed` 是技术失败终态，可以重新提交 PDF。本地等待超时或取消只停止等待，不取消远程任务。保留任务 ID，通过 `status`、`wait` 或 `result` 恢复。

默认 `local` 成功结果是平铺的 `variables` / `factors` / `motivations` 图谱；`--format graph` 返回 `addressed_problems`、`open_questions` 和 `graph.nodes` / `graph.edges`，字段类型与 `bohr lkm graph` 相同，但不是该命令的信封。两种格式的局部节点 ID 均仅在本次抽取内有效。只有非空的 `global_id` 才能复用于 `variables batch`，或在它指向受支持的结论论断时用于 `claim reasoning`。

`partial` 或 `failed` 时检查阶段、失败原因和中间文件。文件 URL 是会过期的凭据，只用于当前恢复任务，不进入聊天记录或日志。

`290017` 表示任务尚未就绪，应继续等待。并发首次 Result 扣费只在返回的结构化错误带 `retryable: true` 时，使用相同任务 ID 和支付来源重试。光子不足可能保留上游子类型 `7002`，但归一化结果是 `BILLING_INSUFFICIENT` 且 `retryable: false`：停止并按下面的支付来源规则处理。保留现有任务，不要为解决计费失败重新提交 PDF。

<a id="capability-boundary"></a>

## 能力边界

当前 CLI 仅提供 LKM 上游服务的聚焦子集，不提供反馈提交或任意响应控制。以 `bohr lkm <command> --help` 为准，不要把每个服务字段转换为参数。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


<a id="detailed-reference"></a>

## 详细参考

全部命令、参数、输出字段、ID 规则、图谱语义、抽取状态和版本契约见 [命令参考](references/commands.md)。
