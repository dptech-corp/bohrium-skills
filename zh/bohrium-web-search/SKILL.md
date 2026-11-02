---
name: bohrium-web-search
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr search 免费检索公开网页；用于一般互联网资料查找、来源发现、结果数量控制及 organic_results 提取。学术记录和私有知识库使用对应领域命令。"
---

<a id="bohrium-web-search"></a>

# Bohrium 网页搜索

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

原始供应商数据可能带 base64 favicon、内部搜索 URL、相关查询等额外字段。只需要自然搜索结果时，使用受支持的 jq-like 路径 `-q '.data.organic_results'`，而不是构造对象表达式。

完整命令、数量校验、输出说明和任务路由边界见 [命令参考](references/commands.md)。
