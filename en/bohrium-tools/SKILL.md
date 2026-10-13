---
name: bohrium-tools
description: "Browse and search the Bohrium scientific Tools library via open.bohrium.com. Use when: user wants to browse scientific computing tools by domain/subdomain, hybrid-search tools by keywords, or view a tool's details (GitHub metrics, tutorial, related entries). NOT for: paper search (use bohrium-paper-search), encyclopedia entries (use bohrium-sciencepedia), knowledge base management (use bohrium-knowledge-base)."
---

# SKILL: Bohrium Scientific Tools Library

## Overview

Access the **Bohrium Tools library** through the `/v2/lkm/tool/*` endpoints on `open.bohrium.com` — a directory of scientific computing software/tools organized as `domain` → `subdomain` → `tool`, with hybrid retrieval (BM25 + vector recall). Tools are identified by `tool_unique_key`.

> **Two hosts, two jobs**: call the **API** on `open.bohrium.com`; build the **clickable reader link** you show to a human on `https://www.bohrium.com` (see [Building page links](#building-page-links)). They are different hosts — don't paste the API URL to a user.

**Use when**:

- Browsing tools under a domain (e.g., all tools under molecular dynamics)
- Hybrid-searching the best-matching tools with natural language + keyword weights
- Viewing a tool's details (profile, GitHub star/fork, tutorial, related entries, Docker image, MCP URL, etc.)

**Don't use for**:

- Paper search → `bohrium-paper-search`
- Encyclopedia entries → `bohrium-sciencepedia`
- Managing your own knowledge base → `bohrium-knowledge-base`

**No CLI support** — HTTP API only.

## Auth configuration

```json
"bohrium-tools": {
  "enabled": true,
  "apiKey": "YOUR_BOHR_ACCESS_KEY",
  "env": {
    "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY"
  }
}
```

## Common template

```python
import os, requests

AK = os.environ["BOHR_ACCESS_KEY"]
BASE = "https://open.bohrium.com/openapi/v2/lkm/tool"

# Language is controlled by the Content-Language header: en-us / zh-cn
H = {
    "Authorization": f"Bearer {AK}",
    "Content-Type": "application/json",
    "Content-Language": "en-us",   # or "zh-cn"
}

# All responses are wrapped in an envelope: {"code": 0, "data": {...}, "trace_id": "..."}.
# The real payload (items / tools / total_num / ...) lives under "data". Always unwrap it.
def data(r):
    r.raise_for_status()
    body = r.json()
    if isinstance(body, dict) and body.get("code") not in (0, None) and "data" not in body:
        raise RuntimeError(f"API error code={body.get('code')} msg={body.get('message')}")
    return body.get("data", body) if isinstance(body, dict) else body
```

> **Language note**: Display language comes from body `language` / `lang`, then the `Content-Language` header (`en-us` / `zh-cn`). If both are empty it falls back to `en-US`; unrecognized values also fall back, they do not error. Canonical values are `en-US` / `zh-CN`.

---

## Actions overview

| Action | Method | Path | Purpose |
|--------|--------|------|---------|
| `domain` | GET | `/domain` | List all tool domains |
| `domain/summary` | GET | `/domain/summary` | Total tool count |
| `subdomain` | POST | `/subdomain` | List subdomains under domains (paged) |
| `subdomain/detail` | POST | `/subdomain/detail` | Get subdomain detail |
| `list` | POST | `/list` | List tools under a subdomain (tag filter / sort / page) |
| `detail` | GET | `/detail` | Get tool detail by `tool_unique_key` |
| `tags` | POST | `/tags` | Get tags under a set of subdomains |
| `search/hybrid` | POST | `/search/hybrid` | Tool hybrid search (text + keyword weights) |

---

## Choosing an action (intent → action)

| The user wants… | Use | You get back |
|-----------------|-----|--------------|
| "What tool domains exist?" / "How many tools?" | `domain` / `domain/summary` | domains with `node_id`; total count |
| "What subdomains are under this domain?" | `subdomain` | subdomains with `node_id`, `tool_num`, `tags` |
| "List tools in this subdomain" (sort / filter) | `list` | tools with `name`, `star_count`, `tool_unique_key` |
| "Filter that list by tag" | `tags` → then `list` with `tag_ids` | tag ids, then filtered tools |
| "Find a tool that does **X**" (natural language) | `search/hybrid` | ranked tools with `tool_unique_key` |
| "Details / GitHub / MCP / Docker / tutorial of this tool" | `detail` (`tool_unique_key`) | full tool profile |

**Rule of thumb**: open-ended "find me a tool for…" → `search/hybrid`; structured "browse this domain" → `domain` → `subdomain` → `list`. Always finish with `detail` before recommending a tool.

## Workflows

### A. Find a tool by need (most common)

1. `search/hybrid` with a natural-language `text` + weighted `keywords`.
2. Take the top `tool_unique_key`(s) → `detail` for repo / metrics / how-to-run.
3. Recommend, each with its `tool_url` (see [Building page links](#building-page-links)).

### B. Browse a domain top-down

1. `domain` → pick a domain `node_id`.
2. `subdomain` (pass `domain_node_ids`) → pick a subdomain.
3. `list` (`sort_by: "popular"`) → top tools → `detail` on the winner.

---

## 1. List domains — `domain`

```python
r = requests.get(f"{BASE}/domain", headers=H)
for it in data(r).get("items", []):
    print(f"- {it['node_name']}  [{it['node_id']}]  tools={it['tool_num']}")
```

Response fields: `data.items[].node_id` / `node_name` / `tool_num`.

---

## 2. Tool count — `domain/summary`

```python
r = requests.get(f"{BASE}/domain/summary", headers=H)
print(data(r).get("total_num"))   # data: {"total_num": 1234}
```

---

## 3. List subdomains — `subdomain`

```python
r = requests.post(f"{BASE}/subdomain", headers=H, json={
    "domain_node_ids": ["DOMAIN_NODE_ID"],   # list of domain node_ids; empty = all
    "page": 1,
    "page_size": 20,
})
body = data(r)
print("total:", body.get("total"))
for it in body.get("items", []):
    print(f"- {it['node_name']}  [{it['node_id']}]  tools={it['tool_num']}  tags={it.get('tags')}")
```

Response: `data.items[].node_id` / `node_name` / `tool_num` / `tags` / `tools[]` (lite: `tool_unique_key`/`tool_name`/`avatar_url`); paging `data.total` / `page` / `pageSize`.

---

## 4. Subdomain detail — `subdomain/detail`

```python
r = requests.post(f"{BASE}/subdomain/detail", headers=H, json={
    "subdomain_node_id": "SUBDOMAIN_NODE_ID",
})
print(data(r))   # data: {"node_id": "...", "node_name": "..."}
```

---

## 5. Tools under a subdomain — `list`

```python
r = requests.post(f"{BASE}/list", headers=H, json={
    "subdomain_node_id": "SUBDOMAIN_NODE_ID",   # required
    "tag_ids": [],            # optional: filter by tag ids
    "sort_by": "popular",     # optional: popular (by star count) / latest (by last commit) / similarity; default popular
    "sort_type": "desc",      # asc / desc
    "page": 1,
    "page_size": 10,
})
body = data(r)
print("total:", body.get("total"))
for it in body.get("items", []):
    print(f"- {it['name']}  star {it['star_count']}  key={it['tool_unique_key']}")
    print(f"    {it.get('profile','')[:120]}")
```

Key `data.items[]` fields: `name` / `profile` / `tags` / `star_count` / `related_entry_count` / `last_commit_time` / `avatar_url` / `tool_unique_key`.

---

## 6. Tool detail — `detail`

```python
r = requests.get(f"{BASE}/detail", headers=H, params={
    "tool_unique_key": "TOOL_UNIQUE_KEY",   # required
})
d = data(r)
print(f"# {d['name']}  star {d['star_count']} / fork {d['fork_count']}")
print("repo:", d.get("repo_url"))
print(d.get("overview", "")[:2000])
```

Key fields of `GetToolDetailResponse`: `name` / `profile` / `version` / `star_count` / `fork_count` / `watch_count` / `key_points` / `overview` / `tutorial` / `related_entry_list` / `primary_language` / `license` / `mcp_url` / `docker_image_uri` / `repo_url` / `help_urls` / `sub_domains` / `tags`.

---

## 7. Subdomain tags — `tags`

```python
r = requests.post(f"{BASE}/tags", headers=H, json={
    "subdomain_node_ids": ["SUBDOMAIN_NODE_ID1", "SUBDOMAIN_NODE_ID2"],   # required
})
for it in data(r).get("items", []):
    print(f"- {it['tag_name']}  [{it['tag_id']}]")
```

---

## 8. Tool hybrid search — `search/hybrid`

Hybrid (BM25 + vector) recall using natural-language `text` + keyword weights `keywords` (keyword → weight).

```python
r = requests.post(f"{BASE}/search/hybrid", headers=H, json={
    "text": "molecular dynamics simulation engine with GPU support",
    "keywords": {"molecular dynamics": 1.0, "GPU": 0.6},  # required: keyword -> weight
    "language": "en-US",      # optional; empty falls back to Content-Language, then en-US
    "k": 50,                  # optional; default 100, max 500
    "return_level": "basic",  # optional: basic / standard / full / complete; score only on full/complete
})
body = data(r)
print("total:", body.get("total"))
for t in body.get("tools", []):
    score = t.get("score")
    score_str = f"{score:.3f}" if isinstance(score, (int, float)) else "-"   # score may be null
    print(f"- {t['tool_name']}  score={score_str}  key={t.get('tool_unique_key')}")
```

Key `data.tools[]` fields: `tool_name` / `tool_unique_key` / `domains` / `subdomains` / `tags` / `profile` / `repo_url` / `score` / `matched_chunks`. `score` only appears when `return_level` is `full` or `complete`.

---

## curl examples

```bash
AK="$BOHR_ACCESS_KEY"
BASE="https://open.bohrium.com/openapi/v2/lkm/tool"

# Tool hybrid search (payload lives under .data)
curl -s -X POST "$BASE/search/hybrid" \
  -H "Authorization: Bearer $AK" \
  -H "Content-Type: application/json" \
  -H "Content-Language: en-us" \
  -d '{"text":"molecular dynamics simulation","keywords":{"molecular dynamics":1.0},"k":20}' \
  | jq '.data.tools[] | {tool_name, score, tool_unique_key}'

# List domains
curl -s "$BASE/domain" \
  -H "Authorization: Bearer $AK" -H "Content-Language: en-us" \
  | jq '.data.items[] | {node_name, node_id, tool_num}'
```

---

## Building page links

When you show a tool or subdomain to a human, attach the clickable reader link on `https://www.bohrium.com` (NOT the API host).

- **Page host**: `https://www.bohrium.com`
- **Language prefix**: `zh-CN` → no prefix; `en-US` → prefix with `/en`

| Link | Pattern | id source |
|------|---------|-----------|
| Tool detail | `{lang}/sciencepedia/agent-tools/{tool_unique_key}` | `tool_unique_key` from `list` / `search/hybrid` / `detail` |
| Subdomain | `{lang}/sciencepedia/agent-tools/c/{subdomain_node_id}` | `node_id` from `subdomain` |
| Tools home | `{lang}/sciencepedia/agent-tools` | — (domain has no standalone page; link to home) |

```text
Tool (en):       https://www.bohrium.com/en/sciencepedia/agent-tools/openmanus
Subdomain (zh):  https://www.bohrium.com/sciencepedia/agent-tools/c/subdomain-llm-frameworks
```

## Response standards

- For each tool / subdomain you list, render it as a markdown link: `[name](tool_url)`.
- When recommending a tool, summarize: what it is (`profile`), traction (`star_count`/`fork_count`), and how to run it (`mcp_url` / `docker_image_uri` / `tutorial`).
- Drive display language with the `Content-Language` header (`zh-cn` for Chinese output); a few search endpoints also accept `language` in the body.
- Be explicit when a field is missing or an endpoint can't deliver the requested granularity.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| All-English (or all-Chinese) results | Wrong `Content-Language` / `language` | Set header `Content-Language: zh-cn` or pass `"language":"zh-CN"` in the search body |
| `list` ignores my `sort_by` | Used a raw column name | Valid values are `popular` / `latest` / `similarity` (not `star_count`/`last_commit_time`); unknown values fall back to `popular` |
| Built a reader link on `open.bohrium.com` | Mixed up API host and page host | Page links use `https://www.bohrium.com` (see *Building page links*) |
| `search/hybrid` returns `text is required` | Missing `text` | Both `text` and `keywords` are required |
| `detail` returns `330003` | `tool_unique_key` does not exist | First obtain a valid `tool_unique_key` via `search/hybrid` |
| `k` truncated | `search/hybrid` caps `k` at 500 | Keep `k` ≤ 500 |
| `list` reports `subdomain_node_id` error | Missing required subdomain id | First obtain `node_id` via `subdomain` |

## Pairs well with

- **tools** to search/browse tools (`domain` → `subdomain` → `list` → `detail`)
- Use **wiki** / **paper-search** to understand a direction first, then **tools** to find implementations
- Use `search/hybrid` for natural-language tool discovery
