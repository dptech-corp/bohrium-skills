---
name: bohrium-pdf-parser
description: "Use when extracting text, tables, charts, formulas, or molecules from PDF files through Bohrium OpenAPI, including URL submission, file upload, asynchronous polling, or per-page results."
---

# SKILL: Bohrium PDF Parser

## Overview

Parse PDF documents using the `open.bohrium.com` PDF parsing service. Extract text, tables, charts, formulas, and molecular structures from PDFs. Two submission methods:

- **URL submission** — provide a PDF download link (e.g. arXiv link)
- **File upload** — upload a local PDF file

**No CLI support** — all operations use the HTTP API.

## Authentication

BOHR_ACCESS_KEY is read from the OpenClaw config `~/.openclaw/openclaw.json`:

```json
"bohrium-pdf-parser": {
  "enabled": true,
  "apiKey": "YOUR_BOHR_ACCESS_KEY",
  "env": {
    "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY"
  }
}
```

OpenClaw automatically injects `env.BOHR_ACCESS_KEY` into the runtime.

The Python examples require the third-party `requests` package. First verify that `python -c "import requests"` succeeds; if the current environment cannot provide it, use the `curl` examples in this skill.

## Common Code Template

```python
import os, time, requests

AK = os.environ.get("BOHR_ACCESS_KEY", "")
BASE = "https://open.bohrium.com/openapi/v2/parse"
HEADERS = {"Authorization": f"Bearer {AK}"}
HEADERS_JSON = {**HEADERS, "Content-Type": "application/json"}
```

---

## Parsing Workflow

```
1. Submit PDF (URL or file upload) → get token
2. Poll result with token → complete when status == "success"
```

Synchronous mode (`sync=true`) blocks until parsing completes but does not include content in the response — you still need `get-result` to retrieve it. Asynchronous mode (`sync=false`, default) requires polling `get-result` until status is `success`.

---

## URL Submission

```python
r = requests.post(f"{BASE}/trigger-url-async", headers=HEADERS_JSON, json={
    "url": "https://arxiv.org/pdf/2107.06922",
    "sync": False,
    "textual": True,
    "table": True,
    "molecule": True,
    "chart": True,
    "figure": False,
    "expression": True,
    "equation": True,
    "pages": [0],           # 0-indexed, omit to parse all pages
    "timeout": 1800
})
data = r.json()
token = data["token"]
print(f"Token: {token}, PDF pages: {data.get('page_count')}")
```

**Response Fields:**

| Field | Description |
|-------|-------------|
| `token` | Task identifier for querying results |
| `page_count` | Total pages in the source PDF |
| `model_version` | Parser model version selected for the task |
| `model_version_source` | Source of the model-version selection |

An asynchronous submission response does not guarantee `status`, `created_time`, or `time_dict`. Read task status from the subsequent `/get-result` response.

---

## File Upload

```python
from pathlib import Path

pdf_path = Path("./paper.pdf")
with open(pdf_path, "rb") as f:
    r = requests.post(f"{BASE}/trigger-file-async",
        headers=HEADERS,       # No Content-Type; requests handles multipart automatically
        files={"file": (pdf_path.name, f, "application/pdf")},
        data={
            "sync": "false",
            "textual": "true",
            "table": "true",
            "molecule": "true",
            "chart": "true",
            "figure": "false",
            "expression": "true",
            "equation": "true",
            "pages": 0,         # multipart only accepts a single integer
            "timeout": 1800
        })
token = r.json()["token"]
```

> **Important**: `pages` in multipart/form-data only accepts a **single integer** (e.g. `0`), not a JSON array `[0]`, or you'll get an `int_parsing` error. In JSON request bodies, arrays like `[0, 1, 2]` are supported.

---

## Query Parse Result

```python
r = requests.post(f"{BASE}/get-result", headers=HEADERS_JSON, json={
    "token": token,
    "content": True,        # Return extracted text
    "objects": False,        # Return extracted objects (tables, figures, etc.)
    "pages_dict": True       # Return per-page results
})
data = r.json()
print(f"Status: {data['status']}, Content length: {len(data.get('content', ''))}")
```

**Response Fields:**

| Field | Description |
|-------|-------------|
| `status` | `success` / `undefined` (processing) / `failed` |
| `token` | Task identifier |
| `content` | Extracted text (LaTeX markup format) |
| `pages_dict` | Result list, not a dictionary; even with `pages=[0]`, its length may equal the source PDF's total pages, so do not infer processed pages from its length or guess the meaning of uninspected elements |
| `result_schema_version` | Upstream result schema version; preserve it when processing or forwarding structured results |
| `lang` | Detected language (`en` / `zh` etc.) |
| `proc_page` / `total_page` | Processed / total pages |
| `proc_textual` / `total_textual` | Processed / total text blocks |
| `proc_table` / `total_table` | Processed / total tables |
| `proc_mol` / `total_mol` | Processed / total molecules |
| `proc_equa` / `total_equa` | Processed / total equations |
| `time_dict` | Per-stage timing details |
| `cost` | Cost |

---

## Full Async Polling Example

```python
import os, time, requests

AK = os.environ.get("BOHR_ACCESS_KEY", "")
BASE = "https://open.bohrium.com/openapi/v2/parse"
HEADERS = {"Authorization": f"Bearer {AK}"}
HEADERS_JSON = {**HEADERS, "Content-Type": "application/json"}

# 1. Submit
r = requests.post(f"{BASE}/trigger-url-async", headers=HEADERS_JSON, json={
    "url": "https://arxiv.org/pdf/2107.06922",
    "sync": False,
    "textual": True, "table": True, "molecule": False,
    "chart": False, "figure": False,
    "expression": True, "equation": True,
    "pages": [0],
    "timeout": 1800
})
submit = r.json()
if submit.get("code"):
    print(f"Submit failed: {submit.get('message')}")
    exit(1)

token = submit["token"]
print(f"Submitted, token={token}")

# 2. Poll for result
for attempt in range(30):
    time.sleep(2)
    r = requests.post(f"{BASE}/get-result", headers=HEADERS_JSON, json={
        "token": token,
        "content": True,
        "objects": False,
        "pages_dict": False
    })
    result = r.json()
    status = result.get("status", "")
    print(f"  [{attempt+1}] status={status}")

    if status == "success":
        print(f"Done! Content length: {len(result.get('content', ''))}")
        print(f"Language: {result.get('lang')}, Cost: {result.get('cost')}")
        print(f"Preview: {result.get('content', '')[:200]}")
        break
    elif status == "failed":
        print(f"Failed: {result.get('description', 'unknown error')}")
        break
else:
    print("Timeout: task did not complete within 60 seconds")
```

---

## Synchronous Mode Example

Synchronous mode (`sync=true`) blocks until parsing completes, so no polling is needed. However, the **response does not include the `content` field** — you still need to call `get-result` to retrieve the parsed content:

```python
# 1. Synchronous submit — blocks until parsing completes
r = requests.post(f"{BASE}/trigger-url-async", headers=HEADERS_JSON, json={
    "url": "https://arxiv.org/pdf/2107.06922",
    "sync": True,           # Wait for completion
    "textual": True, "table": True,
    "molecule": False, "chart": False, "figure": False,
    "expression": True, "equation": True,
    "pages": [0],
    "timeout": 1800
})
submit = r.json()
token = submit["token"]
# submit["status"] == "success", but no content field

# 2. Retrieve content
r = requests.post(f"{BASE}/get-result", headers=HEADERS_JSON, json={
    "token": token,
    "content": True, "objects": False, "pages_dict": False
})
result = r.json()
print(f"Content: {result['content'][:200]}")
```

---

## Parse Options Reference

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sync` | bool | `false` | `true` blocks until complete (still need `get-result` for content), `false` requires polling |
| `textual` | bool | - | Extract text content |
| `table` | bool | - | Extract tables |
| `molecule` | bool | - | Extract molecular structures |
| `chart` | bool | - | Extract charts |
| `figure` | bool | - | Extract figures/images |
| `expression` | bool | - | Extract math expressions |
| `equation` | bool | - | Extract equations |
| `pages` | list[int] | all | Pages to parse (0-indexed) |
| `timeout` | int | - | Timeout in seconds |

---

## curl Examples

```bash
AK="$BOHR_ACCESS_KEY"
BASE="https://open.bohrium.com/openapi/v2/parse"

# URL submission
curl -s -X POST "$BASE/trigger-url-async" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"url":"https://arxiv.org/pdf/2107.06922","sync":false,"textual":true,"table":true,"molecule":false,"chart":false,"figure":false,"expression":true,"equation":true,"pages":[0],"timeout":1800}'

# File upload
curl -s -X POST "$BASE/trigger-file-async" \
  -H "Authorization: Bearer $AK" \
  -F "file=@paper.pdf" \
  -F "sync=false" -F "textual=true" -F "table=true" \
  -F "pages=0"

# Query result
curl -s -X POST "$BASE/get-result" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AK" \
  -d '{"token":"YOUR_TOKEN","content":true,"objects":false,"pages_dict":true}'
```

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `AccessKey is required` | Missing or incorrect auth | Use `Authorization: Bearer $BOHR_ACCESS_KEY` |
| `int_parsing` error | `pages` sent as JSON array in file upload | Use a single integer for `pages` in multipart form |
| `status: undefined` | Async task not yet complete | Poll `get-result` again; recommended interval: 2 seconds |
| Connection timeout | Domain/network issue | Use `open.bohrium.com`; test connectivity via `curl -I https://open.bohrium.com/openapi` |
| Content has LaTeX markup | Normal behavior | Results use `\begin{title}` etc. to mark structure; post-process to extract plain text |
| Large file parses slowly | Many pages or complex content | Use `pages` parameter to limit scope |
| `ModuleNotFoundError: requests` | The current Python environment lacks `requests` | Use an environment that provides it, or use the `curl` examples in this skill |
| `pages_dict` is longer than requested pages | List length does not necessarily equal processed pages | Use `proc_page` / `total_page` for progress and inspect the required elements themselves; do not rely on list length |
