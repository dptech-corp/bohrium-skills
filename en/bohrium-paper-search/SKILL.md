---
name: bohrium-paper-search
description: "Search academic papers and patents via open.bohrium.com RAG engine. Use when: user asks about searching/finding academic papers, literature review, patent search, or technical survey using keywords or natural language questions. NOT for: knowledge base management, file management, or dataset operations."
---

# SKILL: Bohrium Paper & Patent Search

## Overview

Search academic papers and patents using the Bohrium RAG search engine. Combines keyword matching with semantic understanding, supporting natural language queries, date range filtering, JCR zone/database filtering, and AI reranking.

**Supported Search Types:**

| Type | Endpoint | Corpus |
|------|----------|--------|
| English papers | `/openapi/v2/paper/rag/pass/keyword` | English academic papers (title, abstract, corpus, figures) |
| Patents | `/openapi/v2/paper/rag/pass/patent` | Global patents (with classification, assignee filtering) |

**Use cases:** Literature review, technical survey, method comparison, trend analysis.

**No CLI support** — all operations use the HTTP API.

## Authentication

BOHR_ACCESS_KEY is read from the OpenClaw config `~/.openclaw/openclaw.json`:

```json
"bohrium-paper-search": {
  "enabled": true,
  "apiKey": "YOUR_BOHR_ACCESS_KEY",
  "env": {
    "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY"
  }
}
```

OpenClaw automatically injects `env.BOHR_ACCESS_KEY` into the runtime.

## Common Code Template

```python
import os, requests

AK = os.environ.get("BOHR_ACCESS_KEY", "")
BASE = "https://open.bohrium.com"
PAPER_BASE = f"{BASE}/openapi/v2/paper"
HEADERS_JSON = {"Authorization": f"Bearer {AK}", "Content-Type": "application/json"}
```

---

## English Paper Search

### Basic Search

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

### Advanced Search (Date Range + JCR Zone + Database + Type)

```python
r = requests.post(f"{PAPER_BASE}/rag/pass/keyword", headers=HEADERS_JSON, json={
    "words": ["deep learning", "protein structure"],
    "question": "deep learning protein structure prediction",
    "type": 1,                          # Search version: 0=basic, 1=enhanced
    "startTime": "2024-01-01",          # Start date YYYY-MM-DD
    "endTime": "2025-01-01",            # End date
    "jcrZones": ["Q1", "Q2"],           # JCR zone filter
    "includeDbs": ["SCI"],              # Database filter
    "areaIds": [],                      # Area IDs (optional)
    "publicationIds": [],               # Publication IDs (optional)
    "pageSize": 20
})
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `words` | string[] | Yes | Keyword list; recommend 3-8 English terms |
| `question` | string | Yes | Natural language research question |
| `type` | integer | No | Keyword endpoint accepts only 0=basic and 1=enhanced. Do not send 3 |
| `search_type` | integer | Required for original text | Send `3` to receive original-text `pieces`. The JSON field is `search_type`, not `type` |
| `startTime` | string | No | Start date `YYYY-MM-DD`, empty string for no limit |
| `endTime` | string | No | End date `YYYY-MM-DD` |
| `jcrZones` | string[] | No | JCR zone filter, e.g. `["Q1","Q2"]` |
| `includeDbs` | string[] | No | Database filter, e.g. `["SCI","SSCI"]` |
| `areaIds` | string[] | No | Area IDs |
| `publicationIds` | number[] | No | Publication IDs |
| `subjectIds` | number[] | No | Subject IDs |
| `pageSize` | integer | Yes | Result count, 1-100, default 50 |

### Response Fields

| Field | Description |
|-------|-------------|
| `code` | 0=success |
| `message` | Status message |
| `data[]` | Paper list |
| `data[].doi` | DOI |
| `data[].paperId` | Paper ID |
| `data[].enName` | English title |
| `data[].zhName` | Chinese title |
| `data[].enAbstract` | English abstract |
| `data[].zhAbstract` | Chinese abstract |
| `data[].authors` | Author list |
| `data[].coverDateStart` | Publication date |
| `data[].publicationEnName` | Journal name |
| `data[].publicationCover` | Journal cover URL |
| `data[].impactFactor` | Impact factor |
| `data[].citationNums` | Citation count |
| `data[].popularity` | Popularity score |
| `data[].pieces` | Original-text string. With `search_type=3`, keep this string. An empty `alltext` does not matter |
| `data[].figures[]` | Related figures (`figureId`, `imageUrl`, `enExplain`) |
| `data[].languageType` | 0=English |

---

## Patent Search

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

**Note**: keyword `type` supports only 0 (basic) and 1 (enhanced). For original-text snippets send `search_type: 3`. `type: 3` returns HTTP 400 with message `requested type is not available`. Patent `type` still supports 0, 1, and 2. `alltext` may be empty.

### Patent Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | integer | No | Patent search version; only 0, 1, 2 are supported |
| `words` | string[] | Yes | Keyword list |
| `question` | string | Yes | Search question or keyword description |
| `pageSize` | integer | Yes | Results per page |

### Patent Response Fields

Returns a dict/object. Patent results are in `data[]`; top-level fields usually include `code`, `data`, `message`, and `trace_id`.

---

## Search Tips

### Keyword Selection

```python
# GOOD: 3-8 professional terms
words = ["molecular dynamics", "force field", "deep potential", "neural network"]

# BAD: Too generic
words = ["science", "research"]
```

### Combine question for Better Relevance

`words` is for exact keyword matching, `question` is for semantic understanding. Best results come from combining both; this applies to English paper keyword search and patent search:

```python
{
    "words": ["GNN", "molecular property", "prediction"],
    "question": "How do graph neural networks predict molecular properties?",
    "pageSize": 20
}
```

### Filter for High-Quality Journals

```python
{
    "words": ["..."],
    "question": "...",
    "jcrZones": ["Q1"],          # Q1 journals only
    "includeDbs": ["SCI"],       # SCI-indexed only
    "startTime": "2023-01-01",   # Recent 2 years
    "endTime": "2025-12-31"
}
```

---

## curl Examples

```bash
AK="$BOHR_ACCESS_KEY"
BASE="https://open.bohrium.com"
PAPER_BASE="$BASE/openapi/v2/paper"

# English paper search (enhanced metadata)
curl -s -X POST "$PAPER_BASE/rag/pass/keyword" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"words":["deep learning","protein"],"question":"deep learning protein structure prediction","type":1,"startTime":"2024-01-01","endTime":"2025-01-01","jcrZones":["Q1"],"pageSize":5}'

# Original-text snippets. The field is search_type, not type
curl -s -X POST "$PAPER_BASE/rag/pass/keyword" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"words":["deep learning","protein"],"question":"deep learning protein structure prediction","search_type":3,"pageSize":5}'

# Patent search
curl -s -X POST "$PAPER_BASE/rag/pass/patent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"type":1,"words":["neural network"],"question":"neural network","pageSize":5}'
```

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `code` is non-zero | Request parameter error | Check `message` field for details |
| 401 Unauthorized | Invalid auth | Verify BOHR_ACCESS_KEY is correct |
| Irrelevant results | Keywords too generic or vague question | Use 3-8 professional terms + clear question |
| Empty results | Date range too narrow or filters too strict | Widen startTime/endTime or remove jcrZones |
| Response has multiple JSON lines | Normal behavior (streaming) | Parse first line only: `json.loads(response.text.split('\n')[0])` |
| Patent pieces empty | Some patents lack corpus indexing | Normal; use `abstracts` for content instead |
| Keyword `type: 3` returns 400 | v2 rejects that field value | Send `search_type: 3`. Original text is `data[].pieces`; do not wait for `alltext` |
