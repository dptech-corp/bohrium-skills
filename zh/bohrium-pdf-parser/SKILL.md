---
name: bohrium-pdf-parser
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr pdf parse 和 bohr pdf result 解析 PDF 的文本、表格、公式、图表、图片、分子或页面布局。用于 Uni-Parser、PDF 解析和结构化内容；科研问题、结论或推理抽取使用 bohr lkm extract。"
---

<a id="bohrium-pdf-parse"></a>

# Bohrium PDF 解析

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 提取本地 PDF | 确认文件 → `--file` 提交 → 保存 token → 读取所需内容块 |
| 提取可直连的 PDF | 确认 URL → `--url` 提交 → 使用同一结果读取流程 |
| 继续被中断的等待 | 用原 token 调用 `result`；重新提交会创建另一份解析 |
| 提取结论或科研推理 | 转到 LKM extract，使用推理提取流程 |

## 工作流：将一份 PDF 变成可用产物

1. 按下方章节确认输入、语义模式和支出，将原始来源与任务 token 一起保留。
2. 只提交一次，再用返回的 `data.token` 替换 `<token>`。JSON 由上游透传；缺少 token 时保留完整响应并报告交接缺失，不能拿其他 task ID 顶替，也不能为了获取 token 改用 human 模式重新提交：

```bash
bohr pdf parse --file ./paper.pdf --table 2 --equation 2 --yes -o json
bohr pdf result <token> --formatted markdown -o json
```

3. 使用内容前先看返回状态。`undefined` 或 `processing` 时在有界间隔内继续查询同一 token；`success` 表示解析完成，`PARSE_FAILED` 表示失败结束。等待超时仅结束本次等待，保留 token 供稍后读取。
4. 成功后检查所需文本或对象及 `proc_page`、`total_page`，记录缺失或部分完成的页面。交付内容时保留页码出处及可追溯的来源和 token。`--return-half` 的非空输出在终态成功前仍是部分结果。
5. 结果读取发生临时错误时，用同一 token 重试免费读取。提交失败或结果未知时，先检查返回诊断和 token，再决定是否重新提交；新解析可能再次收费。关键表格或公式数值在使用前要对照原始页面。

通过与版本匹配的 CLI 使用 Bohrium Uni-Parser，从单份 PDF 提取布局和内容。只选择一种输入方式：

- `bohr pdf parse --file <path>` 上传本地 PDF。
- `bohr pdf parse --url <url>` 让服务下载 PDF URL。

提交前确认具体本地文件或 URL。本地文件会离开当前机器；URL 必须能被解析服务直接访问。

本流程提取布局、文本、表格、公式、图表、图片或分子。若用户要的是科学论断、研究问题、结论、推理步骤或 LKM 变量图谱，使用 `bohr lkm extract`。

<a id="authenticate-and-authorize-parsing"></a>

## 身份与解析费用授权

查看当前身份和价格契约：

```bash
bohr auth whoami -o json
bohr billing pricing --resource pdf -o json
```

内嵌价目中 PDF 解析为 **0.05 CNY per page（每页）**。最终费用取决于成功解析的页数，服务尚未返回文档大小或计费结果时不能编造总价。非交互调用只有费用已获授权后才能添加 `--yes`；获取已有结果免费，不需要费用确认。

可用 `--dry-run` 检查输入，不会提交或上传文件。

<a id="submit-the-pdf"></a>

## 提交 PDF

默认启用快速文本和表格提取。只启用任务需要的额外语义模式：

```bash
bohr pdf parse --file ./paper.pdf \
  --textual 2 \
  --table 2 \
  --equation 2 \
  --chart 1 \
  --figure 0 \
  --molecule 0 \
  --yes \
  -o json
```

远程文件示例：

```bash
bohr pdf parse --url "https://example.org/paper.pdf" --yes -o json
```

从成功提交的响应中保留 `data.token`。`bohr pdf result` 使用这个 token，而不是其他解析器的任务 ID。

<a id="retrieve-to-a-terminal-state"></a>

## 查询直到明确终态

解析默认异步执行：

```bash
bohr pdf result <token> -o json
```

`undefined` 和 `processing` 是非终态。使用适中的有界轮询间隔，例如 3–5 秒。返回 `success`、命令返回 `PARSE_FAILED` 或用户等待预算耗尽时停止。本地等待截止只停止轮询，不会取消远程解析。

`--sync` 使提交等待解析完成，但返回 token 仍是交给 `bohr pdf result` 获取所需内容和结构化数据的稳定标识。

只有任务确实需要已完成的部分结果时才使用 `--return-half`。非空的部分内容不能证明解析已经进入终态。

<a id="select-the-result-shape"></a>

## 选择结果形状

默认请求启用 `--content` 和 `--objects`。需要更窄或分层输出时显式选择：

```bash
bohr pdf result <token> --content --objects -o json
bohr pdf result <token> --content=false --objects=false --pages-dict -o json
bohr pdf result <token> --content=false --objects=false --pages-tree -o json
bohr pdf result <token> --formatted markdown -o json
bohr pdf result <token> --molecule-source -o json
```

`--content`、`--objects`、`--pages-dict`、`--pages-tree` 至少保留一项启用。尽管名称像对象，`pages_dict` 和 `pages_tree` 实际是数组；进度使用 `proc_page` 和 `total_page`，不能用数组长度代替。

结构化块保留上游的 class/type、坐标、置信度、层级及未来供应商字段。图表和表格识别结果是提取证据，不是已验证的数值真相；计算或录入关键数值前与原始 PDF 核对。

<a id="choose-the-payment-source"></a>

## 选择支付来源

不指定 `--payment-source` 时，CLI 不发送支付渠道偏好：网关优先使用光子，仅在明确光子不足时回退余额。获取费用授权时向用户说明此策略。限定单一来源时，使用 `--payment-source photon` 或 `--payment-source balance`。显式选择 photon 时，光子不足即停止；显式选择 balance 时直接使用余额。现有免费额度和豁免仍然适用。

显式指定的渠道返回 `BILLING_INSUFFICIENT` 且 `retryable: false` 时，报告拒绝结果，等待用户为该渠道充值，或授权原命令使用另一渠道。显式选择余额但余额不足时，也会停止而不扣光子；只有获准后才能改用 `--payment-source photon`。`--yes` 确认执行，不代表授权切换支付来源。只有余额渠道已获授权时才使用 `--payment-source balance`。多步流程的后续计费调用应保持已选渠道；上次 CLI 调用的参数不会保存。`bohr api` 遵循同一规则。


<a id="detailed-reference"></a>

## 详细参考

准确的参数默认值、格式选项、结果字段、轮询状态、错误处理和内嵌版本契约见 [命令参考](references/commands.md)。
