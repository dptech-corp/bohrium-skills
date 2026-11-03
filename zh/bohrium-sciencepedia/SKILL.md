---
name: bohrium-sciencepedia
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr wiki 检索和解释科学词条、关键词、学科课程与知识图谱。用于 SciencePedia、科学概念解释、课程知识树和百科图谱；论文与个人知识库分别使用 bohr paper 和 bohr kb。"
---

<a id="bohrium-sciencepedia"></a>

# Bohrium 科学百科

## 工作流：解释概念并说明前置知识

| 用户需要的产物 | 执行顺序 |
|---|---|
| 概念解释 | 搜索 → 分流 article 或 keyword ID → 阅读 → 附阅读链接解释 |
| 学习大纲 | 搜索 → 返回的 `field_id` → 课程树 → 选读 entry 文章 |
| 相关概念 | 确认 entry 或 keyword ID → 图谱 → 保留边的方向 |

1. 确定语言和风格，按下方价格授权有界的搜索、阅读及图谱调用。
2. 搜索后按 ID 路由章节选择记录。下面展示 article 分支；`type=keyword` 的结果应使用 `keyword get`：

```bash
bohr wiki search "quantum entanglement" --lang en-US --style Feynman --yes -o json
bohr wiki article <article_id> --lang en-US --style Feynman --yes -o json
bohr wiki graph <article_id> --yes -o json
```

3. 仅从选中且 `type=article` 的 `data.articles[].id` 填入 `<article_id>`。图谱为可选步骤，只在需要关系且费用获授权时调用。课程使用 `data.fields[].field_id`，不能使用 major 或 level ID。
4. 完成时给出有内容依据的解释、阅读链接及相关前置知识或关系。保留图谱方向，关系标签不等于因果证据。课程应保持章节和 entry 层级，而非压平成搜索列表。
5. 内容未生成或为空时，按下方语言或风格规则回退并说明。如果没有课程 `field_id`，停止该分支，不能用其他 ID 顶替。重复付费搜索或扩图需要剩余授权预算。

通过与版本匹配的 `bohr wiki` 命令操作 SciencePedia，先按用户意图选择路径：

- `wiki search` 搜索词条，返回摘要和 ID。
- `wiki article` 读取完整词条，`wiki keyword get` 读取较小的概念卡片。
- `wiki levels` 免费浏览学科/层级结构。
- `wiki course` 读取课程章节和知识条目树。
- `wiki graph` 将词条或关键词展开为相关概念图。
- `bohr wiki tools` 在工具库发现科学软件；路径和价格可通过 `bohr guide read bohr-tools` 查看。

论文使用 `bohr paper search`，个人知识库使用 `bohr kb`。SciencePedia 是百科系统，应与这两类任务区分。

<a id="authentication-output-and-billing"></a>

## 身份、输出和计费

使用现有 Bohrium 登录态或 AccessKey，优先结构化输出：

```bash
bohr auth whoami -o json
bohr billing pricing --resource wiki -o json
```

当前 CLI 的固定价格：

- **0.01 CNY**：`wiki search` 和 `wiki course`。
- **0.02 CNY**：`wiki article`、`wiki keyword get` 和 `wiki graph`。

这些调用使用 Wiki/Tool 共享的每月 1,000 次额度。即使额度可能使调用免费，确认门禁仍适用。非交互调用只有费用已获授权后才能添加 `--yes`，否则报告 `CONFIRMATION_REQUIRED`。

`wiki levels`、其兼容入口 `wiki disciplines` 和 `wiki keyword list` 是免费的发现命令。

<a id="search-and-route-ids"></a>

## 搜索并按 ID 分流

```bash
bohr wiki search "quantum entanglement" \
  --lang en-US \
  --style Feynman \
  --yes \
  -o json
```

搜索会混合概念和课程，按结构化结果分流，不能仅看显示名称：

- `data.articles[]` 中 `type=article`：其 `id` 是交给 `bohr wiki article <id>` 的词条 ID。
- `data.articles[]` 中 `type=keyword`：其 `id` 是交给 `bohr wiki keyword get <id>` 的关键词 ID。
- `data.fields[]`：`field_id` 是交给 `bohr wiki course <field_id>` 的课程领域 ID。
- 条目或关键词 ID 均可作为 `bohr wiki graph <id>` 的中心。

面向用户呈现题名、清理后的一句话摘要及阅读链接。移除结构化摘要中的 `<em>` 高亮标签；结果类型用于命令和链接分流，无需让用户额外理解内部类型。

<a id="read-concepts-and-courses"></a>

## 读取概念与课程

```bash
bohr wiki article <article_id> --lang en-US --style Feynman --yes -o json
bohr wiki keyword get <keyword_id> --lang en-US --style Feynman --yes -o json
bohr wiki levels --lang en-US -o json
bohr wiki course <field_id> --lang en-US --style Feynman --yes -o json
```

`levels` 返回学科及层级 ID；`disciplines` 是调用相同层级的兼容入口，新流程优先使用 `levels`。当前 CLI 不能列出某层级下的全部课程，应从 `wiki search` 获取课程 field ID。学科和层级 ID 仅用于导航，不是课程 field ID。搜索未返回 `field_id` 时停止课程流程并报告限制。

`course` 返回递归的领域 → 分类 → 章节 → 条目树，以及基础、核心和进阶数量。呈现提纲时保留层级；条目叶节点的 `entry_id` 可用于打开百科文章。

所选语言/风格尚未生成词条或关键词正文时，可回退搜索摘要或尝试其他受支持的语言/风格，并说明回退。不能把空文档说成完整解释。

<a id="explore-a-concept-graph"></a>

## 探索概念图谱

```bash
bohr wiki graph <entry_or_keyword_id> --yes -o json
```

图谱返回领域、概念节点和关系。节点优先使用 `display_name`，再依次回退 `node_name`、`name`、`node_id`。先解释中心，再解释最相关的关系；保留来源和目标方向，并区分关系标签与较长的关系描述。

当前 graph 没有语言、风格、深度、节点详情或跨领域过滤参数，只使用 `bohr wiki graph --help` 展示的参数。

<a id="reading-links"></a>

## 阅读链接

API 调用保留在 `bohr` 内，面向用户的阅读链接使用 `https://www.bohrium.com`。所有动态 ID 都需要 URL 编码。

- 词条：`{site}/sciencepedia/{style-lowercase}/{entry_id}`。
- 关键词：`{site}/sciencepedia/{style-lowercase}/keyword/{keyword_id}`。
- 课程：`{site}/sciencepedia/field/{style-lowercase}/{field_id}`。

`en-US` 的 `{site}` 是 `https://www.bohrium.com/en`；`zh-CN` 则是 `https://www.bohrium.com`。例如英文 Feynman 词条以 `/en/sciencepedia/feynman/` 开头，中文 Feynman 关键词路径包含 `/sciencepedia/feynman/keyword/`，所有课程链接都包含 `/sciencepedia/field/`。

<a id="capability-boundary"></a>

## 能力边界

当前 CLI 有意不提供上游服务的部分操作，包括按层级列出课程、独立图节点/边详情，以及图深度/领域过滤。以 `bohr wiki <command> --help` 为准，不要将服务请求字段推断为 CLI 参数。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


<a id="detailed-reference"></a>

## 详细参考

全部命令、参数、输出、计费、回退、图谱、阅读链接和版本信息见 [命令参考](references/commands.md)。
