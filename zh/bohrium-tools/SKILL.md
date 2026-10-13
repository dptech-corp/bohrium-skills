---
name: bohrium-tools
description: "浏览与检索 Bohrium 科学工具库（Tools），通过 open.bohrium.com 访问。Use when: 用户需要按领域/子领域浏览科学计算工具、按关键词混合检索工具、查看某个工具的详情（GitHub 指标、教程、关联词条等）。NOT for: 论文检索（用 bohrium-paper-search）、百科词条（用 bohrium-sciencepedia）、知识库管理（用 bohrium-knowledge-base）。"
---

# SKILL: Bohrium 科学工具库（Tools）

## 概述

通过 `open.bohrium.com` 的 `/v2/lkm/tool/*` 端点访问 **Bohrium 科学工具库**——一个按 `domain`(领域) → `subdomain`(子领域) → `tool`(工具) 层级组织的科学计算软件/工具目录，并提供混合检索（BM25 + 向量召回）能力。工具用 `tool_unique_key` 标识。

> **两个 Host，各司其职**：调用 **API** 用 `open.bohrium.com`；给人看的**可点击阅读链接**用 `https://www.bohrium.com`（见[构造页面链接](#构造页面链接)）。二者是不同的 Host——不要把 API URL 直接发给用户。

**典型场景**：

- 浏览某个领域下的工具（如分子动力学领域下的所有工具）
- 用自然语言 + 关键词混合检索最匹配的工具
- 查看某个工具的详细信息（简介、GitHub star/fork、教程、关联词条、Docker 镜像、MCP 地址等）

**不适用**：

- 搜论文 → `bohrium-paper-search`
- 查百科词条 → `bohrium-sciencepedia`
- 管理自己的知识库 → `bohrium-knowledge-base`

**无 CLI 支持** — 通过 HTTP API 操作。

## 认证配置

```json
"bohrium-tools": {
  "enabled": true,
  "apiKey": "YOUR_BOHR_ACCESS_KEY",
  "env": {
    "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY"
  }
}
```

## 通用代码模板

```python
import os, requests

AK = os.environ["BOHR_ACCESS_KEY"]
BASE = "https://open.bohrium.com/openapi/v2/lkm/tool"

# 语言通过 Content-Language 头控制：en-us / zh-cn
H = {
    "Authorization": f"Bearer {AK}",
    "Content-Type": "application/json",
    "Content-Language": "en-us",   # 或 "zh-cn"
}

# 所有响应都包了一层信封：{"code": 0, "data": {...}, "trace_id": "..."}。
# 真正的数据（items / tools / total_num / ...）都在 "data" 里，务必先解包。
def data(r):
    r.raise_for_status()
    body = r.json()
    if isinstance(body, dict) and body.get("code") not in (0, None) and "data" not in body:
        raise RuntimeError(f"API error code={body.get('code')} msg={body.get('message')}")
    return body.get("data", body) if isinstance(body, dict) else body
```

> **语言说明**：展示语言由请求体 `language` / `lang`、HTTP 头 `Content-Language`（`en-us` / `zh-cn`）决定；都空则回落 `en-US`，无法识别也不报错。规范值 `en-US` / `zh-CN`。

---

## Action 概览

| Action | 方法 | 路径 | 用途 |
|--------|------|------|------|
| `domain` | GET | `/domain` | 列出所有工具领域 |
| `domain/summary` | GET | `/domain/summary` | 工具总数统计 |
| `subdomain` | POST | `/subdomain` | 列出领域下的子领域（分页） |
| `subdomain/detail` | POST | `/subdomain/detail` | 获取子领域详情 |
| `list` | POST | `/list` | 列出子领域下的工具（标签筛选/排序/分页） |
| `detail` | GET | `/detail` | 按 `tool_unique_key` 获取工具详情 |
| `tags` | POST | `/tags` | 获取一组子领域下的标签 |
| `search/hybrid` | POST | `/search/hybrid` | 工具混合检索（文本 + 关键词权重） |

---

## 如何选择 Action（意图 → Action）

| 用户想要…… | 用 | 拿到 |
|-----------|----|------|
| “有哪些工具领域？” / “工具总量？” | `domain` / `domain/summary` | 领域及其 `node_id`；总数 |
| “这个领域下有哪些子领域？” | `subdomain` | 子领域及 `node_id`、`tool_num`、`tags` |
| “列出该子领域下的工具”（排序/筛选） | `list` | 工具的 `name`、`star_count`、`tool_unique_key` |
| “再按标签筛这个列表” | `tags` → 再 `list` 传 `tag_ids` | 标签 id，再得筛选后的工具 |
| “找一个能做 **X** 的工具”（自然语言） | `search/hybrid` | 带 `tool_unique_key` 的工具排名 |
| “某工具的详情 / GitHub / MCP / Docker / 教程” | `detail`（`tool_unique_key`） | 完整工具档案 |

**经验法则**：开放式“帮我找个做……的工具”→ `search/hybrid`；结构化“浏览这个领域”→ `domain` → `subdomain` → `list`。推荐工具前务必先 `detail`。

## 工作流

### A. 按需求找工具（最常见）

1. `search/hybrid`，传自然语言 `text` + 带权重的 `keywords`。
2. 取 Top 的 `tool_unique_key` → `detail` 拿仓库 / 指标 / 运行方式。
3. 给出推荐，每个都附 `tool_url`（见[构造页面链接](#构造页面链接)）。

### B. 自顶向下浏览领域

1. `domain` → 选定领域 `node_id`。
2. `subdomain`（传 `domain_node_ids`）→ 选定子领域。
3. `list`（`sort_by: "popular"`）→ 热门工具 → 对优胜者 `detail`。

---

## 1. 列出领域 — `domain`

```python
r = requests.get(f"{BASE}/domain", headers=H)
for it in data(r).get("items", []):
    print(f"- {it['node_name']}  [{it['node_id']}]  tools={it['tool_num']}")
```

返回字段：`data.items[].node_id` / `node_name` / `tool_num`。

---

## 2. 工具总数 — `domain/summary`

```python
r = requests.get(f"{BASE}/domain/summary", headers=H)
print(data(r).get("total_num"))   # data: {"total_num": 1234}
```

---

## 3. 列出子领域 — `subdomain`

```python
r = requests.post(f"{BASE}/subdomain", headers=H, json={
    "domain_node_ids": ["DOMAIN_NODE_ID"],   # 领域 node_id 列表，可为空表示全部
    "page": 1,
    "page_size": 20,
})
body = data(r)
print("total:", body.get("total"))
for it in body.get("items", []):
    print(f"- {it['node_name']}  [{it['node_id']}]  tools={it['tool_num']}  tags={it.get('tags')}")
```

返回字段：`data.items[].node_id` / `node_name` / `tool_num` / `tags` / `tools[]`(轻量工具：`tool_unique_key`/`tool_name`/`avatar_url`)；分页 `data.total` / `page` / `pageSize`。

---

## 4. 子领域详情 — `subdomain/detail`

```python
r = requests.post(f"{BASE}/subdomain/detail", headers=H, json={
    "subdomain_node_id": "SUBDOMAIN_NODE_ID",
})
print(data(r))   # data: {"node_id": "...", "node_name": "..."}
```

---

## 5. 子领域下的工具列表 — `list`

```python
r = requests.post(f"{BASE}/list", headers=H, json={
    "subdomain_node_id": "SUBDOMAIN_NODE_ID",   # 必填
    "tag_ids": [],            # 可选：按标签 id 过滤
    "sort_by": "popular",     # 可选：popular(按星数) / latest(按最近提交) / similarity；默认 popular
    "sort_type": "desc",      # asc / desc
    "page": 1,
    "page_size": 10,
})
body = data(r)
print("total:", body.get("total"))
for it in body.get("items", []):
    print(f"- {it['name']}  ★{it['star_count']}  key={it['tool_unique_key']}")
    print(f"    {it.get('profile','')[:120]}")
```

返回 `data.items[]` 关键字段：`name` / `profile` / `tags` / `star_count` / `related_entry_count` / `last_commit_time` / `avatar_url` / `tool_unique_key`。

---

## 6. 工具详情 — `detail`

```python
r = requests.get(f"{BASE}/detail", headers=H, params={
    "tool_unique_key": "TOOL_UNIQUE_KEY",   # 必填
})
d = data(r)
print(f"# {d['name']}  ★{d['star_count']} / fork {d['fork_count']}")
print("repo:", d.get("repo_url"))
print(d.get("overview", "")[:2000])
```

返回 `GetToolDetailResponse` 关键字段：`name` / `profile` / `version` / `star_count` / `fork_count` / `watch_count` / `key_points` / `overview` / `tutorial` / `related_entry_list` / `primary_language` / `license` / `mcp_url` / `docker_image_uri` / `repo_url` / `help_urls` / `sub_domains` / `tags`。

---

## 7. 子领域标签 — `tags`

```python
r = requests.post(f"{BASE}/tags", headers=H, json={
    "subdomain_node_ids": ["SUBDOMAIN_NODE_ID1", "SUBDOMAIN_NODE_ID2"],   # 必填
})
for it in data(r).get("items", []):
    print(f"- {it['tag_name']}  [{it['tag_id']}]")
```

---

## 8. 工具混合检索 — `search/hybrid`

用自然语言 `text` + 关键词权重 `keywords`(关键词→权重) 做 BM25 + 向量的混合召回。

```python
r = requests.post(f"{BASE}/search/hybrid", headers=H, json={
    "text": "molecular dynamics simulation engine with GPU support",
    "keywords": {"molecular dynamics": 1.0, "GPU": 0.6},  # 必填：关键词→权重
    "language": "en-US",      # 可选；空则看 Content-Language，再空则 en-US
    "k": 50,                  # 可选；默认 100，最大 500
    "return_level": "basic",  # 可选：basic / standard / full / complete；非 full/complete 不回 score
})
body = data(r)
print("total:", body.get("total"))
for t in body.get("tools", []):
    score = t.get("score")
    score_str = f"{score:.3f}" if isinstance(score, (int, float)) else "-"   # score 可能为 null
    print(f"- {t['tool_name']}  score={score_str}  key={t.get('tool_unique_key')}")
```

返回 `data.tools[]` 关键字段：`tool_name` / `tool_unique_key` / `domains` / `subdomains` / `tags` / `profile` / `repo_url` / `score` / `matched_chunks`。`score` 仅 `return_level=full/complete` 时出现。

---

## curl 示例

```bash
AK="$BOHR_ACCESS_KEY"
BASE="https://open.bohrium.com/openapi/v2/lkm/tool"

# 工具混合检索（数据在 .data 下）
curl -s -X POST "$BASE/search/hybrid" \
  -H "Authorization: Bearer $AK" \
  -H "Content-Type: application/json" \
  -H "Content-Language: en-us" \
  -d '{"text":"molecular dynamics simulation","keywords":{"molecular dynamics":1.0},"k":20}' \
  | jq '.data.tools[] | {tool_name, score, tool_unique_key}'

# 列出领域
curl -s "$BASE/domain" \
  -H "Authorization: Bearer $AK" -H "Content-Language: en-us" \
  | jq '.data.items[] | {node_name, node_id, tool_num}'
```

---

## 构造页面链接

把工具或子领域展示给人看时，附上 `https://www.bohrium.com` 上的可点击阅读链接（**不是** API host）。

- **页面 Host**：`https://www.bohrium.com`
- **语言前缀**：`zh-CN` → 无前缀；`en-US` → 加前缀 `/en`

| 链接 | 模板 | id 来源 |
|------|------|---------|
| 工具详情 | `{lang}/sciencepedia/agent-tools/{tool_unique_key}` | `list` / `search/hybrid` / `detail` 的 `tool_unique_key` |
| 子领域 | `{lang}/sciencepedia/agent-tools/c/{subdomain_node_id}` | `subdomain` 的 `node_id` |
| 工具首页 | `{lang}/sciencepedia/agent-tools` | —（领域无独立页面，指向首页） |

```text
工具（en）：   https://www.bohrium.com/en/sciencepedia/agent-tools/openmanus
子领域（zh）： https://www.bohrium.com/sciencepedia/agent-tools/c/subdomain-llm-frameworks
```

## 回答规范

- 列出的每个工具/子领域，都用 markdown 链接呈现：`[名称](tool_url)`。
- 推荐工具时，概括：是什么（`profile`）、热度（`star_count`/`fork_count`）、怎么跑（`mcp_url` / `docker_image_uri` / `tutorial`）。
- 用 `Content-Language` 头控制展示语言（中文输出设 `zh-cn`）；少数检索接口也接受请求体里的 `language`。
- 字段缺失或接口无法满足所需粒度时，要如实说明。

---

## 常见问题

| 问题 | 原因 | 解决 |
|------|------|------|
| 检索结果全英文（或全中文） | `Content-Language` / `language` 不对 | 设置头 `Content-Language: zh-cn` 或在检索体内传 `"language":"zh-CN"` |
| `list` 的 `sort_by` 不生效 | 用了原始列名 | 合法值为 `popular` / `latest` / `similarity`（不是 `star_count`/`last_commit_time`）；未知值回退为 `popular` |
| 阅读链接拼到了 `open.bohrium.com` 上 | 混淆了 API host 与页面 host | 页面链接用 `https://www.bohrium.com`（见“构造页面链接”） |
| `search/hybrid` 报 `text is required` | 缺少 `text` | `text` 与 `keywords` 均为必填 |
| `detail` 报 `330003` | `tool_unique_key` 不存在 | 先用 `search/hybrid` 拿到正确的 `tool_unique_key` |
| `k` 过大被截断 | `search/hybrid` 的 `k` 上限 500 | 控制在 500 以内 |
| `list` 报 `subdomain_node_id` 错 | 缺少必填的子领域 id | 先用 `subdomain` 拿到 `node_id` |

## 搭配使用

- **tools** 检索/浏览工具（`domain` → `subdomain` → `list` → `detail`）
- 先用 **wiki** / **paper-search** 了解某个方向，再用 **tools** 找对应的实现工具
- 用 `search/hybrid` 做自然语言找工具
