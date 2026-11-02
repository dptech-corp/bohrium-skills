---
name: bohrium-scholar-search
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr scholar search 和 bohr scholar info 检索学者并读取画像。用于姓名消歧、机构和研究方向筛选，以及科研成果和指标的身份核对。"
---

<a id="bohrium-scholar-search"></a>

# Bohrium 学者检索

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 从姓名确定学者 | 搜索 → 消歧 → 读取选中画像 |
| 找某领域或机构的学者 | 使用用户提供的学校或标签筛选 → 比较候选 |
| 查看已知学者 | 将准确 scholar ID 直接传给 `info` |

## 工作流：先确定身份，再归属成果

1. 使用姓名搜索并限制结果数量。同名候选分别保留；仅在需求有依据时补充机构或研究筛选条件。

```bash
bohr scholar search "Geoffrey Hinton" --page 1 --size 10 -o json
bohr scholar info <scholar_id> -o json
```

2. 将 `<scholar_id>` 替换为选中记录实际返回的学者标识。搜索数据由上游透传，先查看记录再选 JSON 路径；姓名和 DOI 都不能代替该 ID。
3. 用详情中的机构和研究方向核对目标人物。如果信息冲突或仍有歧义，展示候选并取得消歧信息后再归属论文。
4. 完成时给出已确认的 scholar ID 和有依据的画像信息。缺失字段保留为未知；引用量是返回时的快照，不能重建为引用历史。`info` 失败不能视为身份已核实；只有错误允许重试时才重试该读取。

不知道学者 ID 时，先搜索：

```bash
bohr scholar search "Geoffrey Hinton" --page 1 --size 10 -o json
bohr scholar search "catalysis" --school "Tsinghua University" --tags "DFT,catalysis" -o json
```

搜索免费且只读。只有用户给出了条件，或未筛选的同名结果存在歧义时，才使用过滤参数。使用返回的学者 ID，不能仅靠显示名称匹配画像。

读取选定画像：

```bash
bohr scholar info <scholar_id> -o json
```

归属论文或指标前，结合姓名、机构、研究方向及其他返回证据核对身份。学者 ID 不是论文 ID、作者姓名字符串或知识库文档 ID。

论文记录使用 `bohr paper search`，专利使用 `bohr paper patent`。当前 scholar CLI 不修改画像，也没有独立的论文、教育经历或引用历史子命令；这些信息可能出现在可用的画像响应内。

完整参数、从搜索到画像的 ID 交接和展示边界见 [命令参考](references/commands.md)。
