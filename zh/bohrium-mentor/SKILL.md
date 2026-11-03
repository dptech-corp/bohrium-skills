---
name: bohrium-mentor
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr agents mentor 向 Bohrium 科学小导师提问，获取文献支持的科学解释、研究进展或方法比较。独立论文记录使用 bohr paper，人物画像使用 scholar，论断证据链使用 LKM。"
---

<a id="bohrium-mentor"></a>

# Bohrium 科学小导师

## 按交付物选择路径

| 目标 | 路径 |
|---|---|
| 解释机制或比较方法 | 用一个聚焦的 Mentor 问题明确范围和比较维度 |
| 获取筛选后的语料或证据表 | 先检索并筛选论文；单次 Mentor 不能证明系统覆盖 |
| 核对某个结论或引用 | 查看论文记录或 LKM 出处，再核实来源 |

## 工作流：产出并检查科研解释

1. 明确科学问题、领域、时间范围和输出要求。比较任务将假设、精度、成本和适用性等维度写进问题文本；这些属于问题内容，不是新的 CLI 参数。
2. 用 `--dry-run` 检查规范化输入，并查看下方会话价格。会话预算获授权后请求结构化输出：

```bash
bohr agents mentor   "Compare DFT and Hartree-Fock for periodic solids: assumptions, accuracy, computational cost and limitations"   --discipline Physics --journal-type foreign --model reason --yes -o json
```

3. 保持进程连接以获取答案。JSON 成功时一起保留 `data.answer` 和 `data.session_id`。会话 ID 用于追溯，不能传给不存在的 CLI 恢复或读取命令。
4. 检查实际答案：即使 CLI 返回 `ok=true`，`data.answer` 为空或明显截断仍是交付未完成；CLI 不会验证服务完成事件。再检查是否覆盖所有比较维度，区分返回的综合解释和独立核实的证据。交付物需要可核验引用时，通过论文或来源检索解析关键引用。
5. 完成时交付答案、不确定性和实际核验范围。流中断或超时时按下方规则报告部分或未知结果。创建另一会话是新的付费动作，不能将重试当作原会话的继续。

通过与版本匹配的 `bohr agents mentor` 提出一个明确的科学问题。Mentor 是综合回答流程，会在内部检索科学文献并流式输出 Markdown 解释。

需要记录而非解释时，选择结构化命令：

- `bohr paper search` 获取具体论文和文献字段。
- `bohr scholar search` 获取学者画像。
- `bohr lkm` 获取论断、推理链、论文图谱和来源关系。
- 可复现检索方案、筛选语料或系统证据表使用专门的文献综述工作流。

<a id="authenticate-and-authorize-the-session"></a>

## 身份与会话费用授权

创建会话前检查当前身份和版本匹配的价格：

```bash
bohr auth whoami -o json
bohr billing pricing --resource mentor -o json
```

内嵌价目为每个会话 **2.00 CNY**。网关在创建会话前预扣费，并非在回答流完成时才收费。请求回答问题本身不代表允许付费。非交互调用在费用已获授权后才添加 `--yes`，否则向用户报告 `CONFIRMATION_REQUIRED`。

用 `--dry-run` 查看归一化的问题和选项，不创建会话：

```bash
bohr agents mentor \
  "Compare density functional theory and Hartree-Fock for periodic solids" \
  --discipline Physics \
  --journal-type foreign \
  --model reason \
  --dry-run \
  -o json
```

<a id="ask-the-question"></a>

## 提出问题

获得授权后，提交一个完整问题：

```bash
bohr agents mentor \
  "Explain recent strategies for suppressing lithium dendrites in solid-state batteries" \
  --discipline Chemistry \
  --journal-type foreign \
  --model reason \
  --yes
```

常见学科值有 `All`、`Physics`、`Chemistry`、`Biology`、`Materials`、`Computer Science`。CLI 将筛选值传给 Mentor；领域不明确时保留默认 `All`。`foreign` 和 `chinese` 选择期刊范围，默认模型为 `reason`。

自动化调用优先 JSON：

```bash
bohr agents mentor "Explain the evidence for high-temperature superconductivity" \
  --yes \
  -o json
```

human 输出只流式呈现主 Markdown 回答；结构化成功结果包含 `data.answer` 和 `data.session_id`。回答内的引用标记属于回答内容，不是完整文献记录；需要独立核查的文献时使用 `bohr paper search`。

<a id="handle-delay-and-interruption"></a>

## 处理延迟和中断

深度推理可能有明显的首 token 延迟，命令超时为五分钟。仍有有用输出到达时，保持进程连接。

创建会话与流式回答属于两个阶段。流中断时，会话可能已经计费，human 输出也可能已有部分回答。保留真实 stdout、结构化错误、计费 notice 和 request ID。只有成功的机器结果才返回 `data.session_id`。

**创建会话后或流中断后不要自动重试。** 重试会新建会话并可能再次收费。报告中断，取得用户对新会话的费用授权后再继续。

<a id="capability-boundary"></a>

## 能力边界

CLI 有意只输出主回答，不提供 Mentor 的其他 UI 布局、引用卡片元数据、推荐追问或 resume 命令。使用 `bohr agents mentor`，不要用 shell 或 Python 重建服务传输协议或凭据处理。

旧的顶层 `bohr mentor` 路径已移除，使用 `bohr agents mentor`。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


<a id="detailed-reference"></a>

## 详细参考

准确参数、输出、计费、dry-run 行为、失败处理和内嵌版本契约见 [命令参考](references/commands.md)。
