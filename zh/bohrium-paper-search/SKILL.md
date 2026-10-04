---
name: bohrium-paper-search
description: "Search academic papers and patents via open.bohrium.com RAG engine. Use when: user asks about searching/finding academic papers, literature review, patent search, or technical survey using keywords or natural language questions. NOT for: knowledge base management, file management, or dataset operations."
---

# SKILL: Bohrium 论文与专利搜索

## 概述

通过 Bohrium RAG 搜索引擎进行学术论文和专利检索。基于关键词匹配 + 语义理解，支持自然语言问题描述、时间范围限定、JCR 分区/数据库筛选和 AI 重排序。

**支持的搜索类型：**

| 类型 | 端点 | 语料库 |
|------|------|--------|
| 英文文献 | `/openapi/v2/paper/rag/pass/keyword` | 英文学术论文（题目、摘要、语料、图片） |
| 专利 | `/openapi/v2/paper/rag/pass/patent` | 全球专利（含分类号、专利权人筛选） |

**适用场景：** 文献综述、技术调研、方法对比、趋势分析。

**无 CLI 支持** — 全部通过 HTTP API 操作。

## 认证配置

BOHR_ACCESS_KEY 从 OpenClaw 配置文件 `~/.openclaw/openclaw.json` 中读取：

```json
"bohrium-paper-search": {
  "enabled": true,
  "apiKey": "YOUR_BOHR_ACCESS_KEY",
  "env": {
    "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY"
  }
}
```

OpenClaw 会自动将 `env.BOHR_ACCESS_KEY` 注入到运行环境。

## 通用代码模板

```python
import os, requests

AK = os.environ.get("BOHR_ACCESS_KEY", "")
BASE = "https://open.bohrium.com"
PAPER_BASE = f"{BASE}/openapi/v2/paper"
HEADERS_JSON = {"Authorization": f"Bearer {AK}", "Content-Type": "application/json"}
```

---

## 英文文献搜索

### 基本搜索

```python
r = requests.post(f"{PAPER_BASE}/rag/pass/keyword", headers=HEADERS_JSON, json={
    "words": ["deep learning", "molecular dynamics"],
    "question": "How to use deep learning for molecular dynamics simulation?",
    "startTime": "",
    "endTime": "",
    "pageSize": 10
})
data = r.json()
print(f"Found {len(data['data'])} papers")
for p in data["data"]:
    print(f"  [{p['doi']}] {p['enName']}")
    print(f"    Journal: {p.get('publicationEnName', '')}, IF: {p.get('impactFactor', 0)}")
    print(f"    Date: {p['coverDateStart']}, Citations: {p['citationNums']}")
```

### 高级搜索（时间范围 + JCR 分区 + 数据库 + type）

```python
r = requests.post(f"{PAPER_BASE}/rag/pass/keyword", headers=HEADERS_JSON, json={
    "words": ["deep learning", "protein structure"],
    "question": "deep learning protein structure prediction",
    "type": 1,                          # 搜索版本：0=普通, 1=加强版
    "startTime": "2024-01-01",          # 起始日期 YYYY-MM-DD
    "endTime": "2025-01-01",            # 截止日期
    "jcrZones": ["Q1", "Q2"],           # JCR 分区筛选
    "includeDbs": ["SCI"],              # 数据库筛选
    "areaIds": [],                      # 领域 ID（可选）
    "publicationIds": [],               # 期刊 ID（可选）
    "pageSize": 20
})
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `words` | string[] | 是 | 关键词列表，建议 3-8 个英文术语 |
| `question` | string | 是 | 研究问题的自然语言描述 |
| `type` | integer | 否 | keyword 接口只支持 0=普通、1=加强版。不要传 3 |
| `search_type` | integer | 核对原文时必填 | 传 `3` 才返回原文片段 `pieces`。JSON 字段名是 `search_type`，不是 `type` |
| `startTime` | string | 否 | 起始日期 `YYYY-MM-DD`，空字符串表示不限 |
| `endTime` | string | 否 | 截止日期 `YYYY-MM-DD` |
| `jcrZones` | string[] | 否 | JCR 分区筛选，如 `["Q1","Q2"]` |
| `includeDbs` | string[] | 否 | 数据库筛选，如 `["SCI","SSCI"]` |
| `areaIds` | string[] | 否 | 领域 ID |
| `publicationIds` | number[] | 否 | 期刊 ID |
| `subjectIds` | number[] | 否 | 主题 ID |
| `pageSize` | integer | 是 | 返回数量，1-100，默认 50 |

### 返回字段

| 字段 | 说明 |
|------|------|
| `code` | 0=成功 |
| `message` | 状态消息 |
| `data[]` | 论文列表 |
| `data[].doi` | DOI |
| `data[].paperId` | 论文 ID |
| `data[].enName` | 英文标题 |
| `data[].zhName` | 中文标题 |
| `data[].enAbstract` | 英文摘要 |
| `data[].zhAbstract` | 中文摘要 |
| `data[].authors` | 作者列表 |
| `data[].coverDateStart` | 发表日期 |
| `data[].publicationEnName` | 期刊英文名 |
| `data[].publicationCover` | 期刊封面 URL |
| `data[].impactFactor` | 影响因子 |
| `data[].citationNums` | 被引次数 |
| `data[].popularity` | 热度 |
| `data[].pieces` | 原文字符串。`search_type=3` 时保留；空 `alltext` 不影响，有 `pieces` 即用 `pieces` |
| `data[].figures[]` | 相关图片（含 `figureId`, `imageUrl`, `enExplain`） |
| `data[].languageType` | 0=英文 |

---

## 专利搜索

```python
r = requests.post(f"{PAPER_BASE}/rag/pass/patent", headers=HEADERS_JSON, json={
    "type": 1,
    "words": ["neural network"],
    "question": "neural network",
    "pageSize": 5
})
data = r.json()
for p in data["data"]:
    print(f"  Patent: {p}")
```

**注意**: keyword 的 `type` 只支持 0（普通）、1（加强版）。要原文片段时改传 `search_type: 3`；`type: 3` 会返回 HTTP 400，`message` 为 `requested type is not available`。patent 的 `type` 仍支持 0、1、2。`alltext` 可以为空。

### 专利请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | integer | 否 | 专利搜索版本，仅支持 0、1、2 |
| `words` | string[] | 是 | 关键词列表 |
| `question` | string | 是 | 检索问题或关键词描述 |
| `pageSize` | integer | 是 | 每页数量 |

### 专利返回字段

返回结构是 dict；专利结果列表在 `data[]`，顶层通常包含 `code`、`data`、`message`、`trace_id` 等字段。

---

## 搜索技巧

### 关键词选择

```python
# GOOD: 3-8 个专业术语
words = ["molecular dynamics", "force field", "deep potential", "neural network"]

# BAD: 太笼统
words = ["science", "research"]
```

### 结合 question 提升相关性

`words` 用于精确关键词匹配，`question` 用于语义理解。两者结合效果最佳，英文文献 keyword 搜索和专利搜索都适用：

```python
{
    "words": ["GNN", "molecular property", "prediction"],
    "question": "How do graph neural networks predict molecular properties?",
    "pageSize": 20
}
```

### 限定高质量期刊

```python
{
    "words": ["..."],
    "question": "...",
    "jcrZones": ["Q1"],          # 只看 Q1 期刊
    "includeDbs": ["SCI"],       # 只看 SCI 收录
    "startTime": "2023-01-01",   # 近两年
    "endTime": "2025-12-31"
}
```

---

## curl 示例

```bash
AK="$BOHR_ACCESS_KEY"
BASE="https://open.bohrium.com"
PAPER_BASE="$BASE/openapi/v2/paper"

# 英文文献搜索（加强版题录）
curl -s -X POST "$PAPER_BASE/rag/pass/keyword" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"words":["deep learning","protein"],"question":"deep learning protein structure prediction","type":1,"startTime":"2024-01-01","endTime":"2025-01-01","jcrZones":["Q1"],"pageSize":5}'

# 原文片段。字段是 search_type，不是 type
curl -s -X POST "$PAPER_BASE/rag/pass/keyword" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"words":["deep learning","protein"],"question":"deep learning protein structure prediction","search_type":3,"pageSize":5}'

# 专利搜索
curl -s -X POST "$PAPER_BASE/rag/pass/patent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"type":1,"words":["neural network"],"question":"neural network","pageSize":5}'
```

---

## 常见问题

| 问题 | 原因 | 解决 |
|------|------|------|
| `code` 非 0 | 请求参数错误 | 检查 `message` 字段获取错误详情 |
| 401 Unauthorized | 鉴权无效 | 确认 BOHR_ACCESS_KEY 正确 |
| 结果不相关 | 关键词太泛或 question 不够具体 | 使用 3-8 个专业术语 + 清晰的 question |
| 返回为空 | 时间范围太窄或筛选条件过严 | 放宽 startTime/endTime 或去掉 jcrZones 限制 |
| 响应含多行 JSON | 正常行为（streaming） | 取第一行解析即可，或用 `json.loads(response.text.split('\n')[0])` |
| 专利 pieces 为空 | 部分专利无语料索引 | 正常现象，可通过 `abstracts` 获取摘要内容 |
| keyword `type: 3` 返回 400 | v2 不接受这个字段值 | 改成 `search_type: 3`。原文在 `data[].pieces`，不要等 `alltext` |
