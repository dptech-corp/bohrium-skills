---
name: bohrium-web-search
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr search 免费检索公开网页；用于一般互联网资料查找、来源发现、结果数量控制及 organic_results 提取。学术记录和私有知识库使用对应领域命令。"
---

<a id="bohrium-web-search"></a>

# Bohrium 网页搜索

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 找官网、官方手册或近期公开信息 | 缩小查询 → 选择一手来源 → 阅读并核实页面 |
| 找论文、学者或私有文档 | 先转到 Paper、Scholar 或 Knowledge Base |
| 回答事实问题 | 将搜索链接作为线索，核实证据后再作答 |

## 工作流：找到文档并交付有出处的答案

1. 明确产品或主题、相关版本，以及用户需要确认的事实。把这些词放进查询；`--limit` 限制线索数量，不代表核验深度。
2. 先做小范围发现：

```bash
bohr search "official VASP documentation PAW method" --limit 5 -o json
```

3. 查看实际返回的容器和 URL 字段，例如 `data.organic_results[].link` 或 `data.results[].url`，将选中条目的标题、URL 和摘要一起保留。用可用的网页读取工具或浏览器打开官方页面。本 Skill 的命令负责搜索，不会抓取全文。
4. 确认页面支持所需结论，且版本和日期适用，再交付答案及来源链接。如果只有摘要，将其标为搜索线索，并说明缺少的核验。
5. 无结果时可按需要再做一次更精确的查询；鉴权或网络错误则先修复对应错误。这些情况都不能证明主题不存在。所需事实已有证据时即完成，避免堆积无关结果。

使用免费、只读的网页搜索命令获取一般公开网页信息：

```bash
bohr search "official VASP documentation PAW method" --limit 5 -o json
bohr search "official VASP documentation PAW method" --limit 5 -o json -q '.data.organic_results'
```

query 是单个位置参数字符串。`--limit` 范围 `1-10`，默认 `5`。优先使用包含主题和目标来源类型的明确检索词。

结构化学术内容和私有内容使用对应命令：

- `bohr paper search`：论文记录及文献字段过滤。
- `bohr scholar search`：研究人员和画像。
- `bohr kb search <kb_id> <query>`：用户 Bohrium 知识库中已有的内容。

搜索结果是发现线索，不会自动成为权威事实。保留响应中的题名、URL、摘要和来源名称；准确性重要时打开或核对一手来源。当前 CLI 没有时间新鲜度、站点/域名、语言区域或供应商专属过滤参数。

原始供应商数据可能带 base64 favicon、内部搜索 URL、相关查询等额外字段。JSON 透传且不会统一结果容器，先检查实际响应；上游可返回 `organic_results[].link` 或 `results[].url`。示例 `-q '.data.organic_results'` 仅在该字段存在时适用，投影为空不能证明没有命中。

完整命令、数量校验、输出说明和任务路由边界见 [命令参考](references/commands.md)。
