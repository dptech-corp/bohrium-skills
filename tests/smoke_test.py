#!/usr/bin/env python3
"""
Smoke-test every Bohrium skill's primary endpoint against open.bohrium.com.

Gateway paths use the v2 prefix (aligned with the current skills). bohrium-job
stays on v1 (its v2 upstream differs and job_group has no v2 route). bohrium-file
uses v1 for normal personal/share file operations and v2 only for upload credentials.

Usage:
    export BOHR_ACCESS_KEY="..."
    # Sandbox smoke requires Python >=3.10 with prerelease lbg installed.
    # Optional: export BOHR_SANDBOX_PYTHON=/path/to/python
    # Optional: export BOHR_SANDBOX_TEMPLATE=doc-compiler
    python3 tests/smoke_test.py

Exits with non-zero if any required-endpoint test fails.

Billing: this hits real endpoints (it is not a mock). List/read endpoints are
free. Endpoints that were already billed stay billed — paper-search (v2 paper/rag,
balance), pdf-parser (v2 parse, per-page balance), bohrium-mentor (session
creation), and bohrium-sandbox (create/exec/kill) make real billable calls.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import time
import uuid
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import ssl
import urllib.request
import urllib.parse
import urllib.error


BASE = os.environ.get("BOHR_API_BASE_URL", "https://open.bohrium.com/openapi")
AK = os.environ.get("BOHR_ACCESS_KEY", "")
TIMEOUT = 60
ROOT = Path(__file__).resolve().parents[1]

# Display/recall language is resolved server-side from the Content-Language
# header (enum.LanguageHeaderKey). Values are lowercase: en-us / zh-cn. This is
# the shared mechanism for both wiki (/wiki_v2/*) and tool (/tool/*) endpoints;
# tool/wiki bodies may also override it with a mixed-case `language` (en-US/zh-CN).
LANG_HEADER = os.environ.get("BOHR_CONTENT_LANGUAGE", "en-us")


def _make_ssl_context() -> ssl.SSLContext:
    """Build an SSL context with a working CA bundle.

    macOS Python builds frequently fail with CERTIFICATE_VERIFY_FAILED because
    they cannot locate system CA certs. Prefer certifi's bundle when available.
    """
    try:
        import certifi

        return ssl.create_default_context(cafile=certifi.where())
    except Exception:  # noqa: BLE001
        return ssl.create_default_context()


SSL_CTX = _make_ssl_context()


if not AK:
    print("ERROR: set BOHR_ACCESS_KEY in env", file=sys.stderr)
    sys.exit(2)


@dataclass
class Result:
    skill: str
    endpoint: str
    status: str     # PASS / FAIL / SKIP
    code: int | None
    note: str = ""


def http(
    method: str,
    path: str,
    *,
    params: dict | None = None,
    body: dict | None = None,
) -> tuple[int, dict]:
    url = BASE + path
    if params:
        url = url + "?" + urllib.parse.urlencode(params)
    data: bytes | None = None
    headers = {"Authorization": f"Bearer {AK}", "Content-Language": LANG_HEADER}
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=TIMEOUT, context=SSL_CTX) as resp:
            raw = resp.read().decode("utf-8", errors="replace")
            try:
                return resp.status, json.loads(raw) if raw else {}
            except json.JSONDecodeError:
                return resp.status, {"_raw": raw[:200]}
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace")
        try:
            body_json = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            body_json = {"_raw": raw[:200]}
        return e.code, body_json
    except urllib.error.URLError as e:
        return 0, {"_err": str(e.reason)}
    except Exception as e:  # noqa: BLE001
        return 0, {"_err": repr(e)}


def _msg(data: dict) -> str:
    for k in ("message", "error", "_raw"):
        v = data.get(k) if isinstance(data, dict) else None
        if v:
            return str(v)[:80]
    return ""


def classify(code: int, data: dict) -> tuple[str, str]:
    """Decide PASS/FAIL and build a note.

    Any 2xx counts as pass. 401/403 counts as FAIL (auth broken). 404 / 400
    usually means wrong endpoint / wrong params. 5xx = server side.
    """
    api_code = data.get("code") if isinstance(data, dict) else None
    if 200 <= code < 300:
        # Some endpoints return {"code": <non-zero>} on error inside a 200 body
        if api_code not in (None, 0, 200, "0"):
            return "FAIL", f"HTTP 200 but body code={api_code} msg={_msg(data)}"
        return "PASS", ""
    if code == 0:
        return "FAIL", data.get("_err", "network error") if isinstance(data, dict) else "network error"
    return "FAIL", f"msg={_msg(data)}"


results: list[Result] = []


def record(
    skill: str,
    endpoint: str,
    method: str,
    *,
    params: dict | None = None,
    body: dict | None = None,
    note_on_pass: str = "",
) -> None:
    code, data = http(method, endpoint, params=params, body=body)
    status, note = classify(code, data)
    if status == "PASS" and note_on_pass:
        note = note_on_pass
    results.append(Result(skill, endpoint, status, code, note))
    print(f"  [{status}] {method:<4} {endpoint}  HTTP={code}  {note}")


def record_file(
    endpoint: str,
    method: str,
    *,
    params: dict | None = None,
    body: dict | None = None,
    require_data_keys: tuple[str, ...] = (),
    note_on_pass: str = "",
) -> None:
    code, data = http(method, endpoint, params=params, body=body)
    status, note = classify(code, data)
    if status == "PASS" and require_data_keys:
        payload = data.get("data") if isinstance(data, dict) else None
        missing = [k for k in require_data_keys if not isinstance(payload, dict) or k not in payload]
        if missing:
            status = "FAIL"
            note = f"missing data keys: {','.join(missing)}"
    if status == "PASS" and note_on_pass:
        note = note_on_pass
    results.append(Result("file", endpoint, status, code, note))
    print(f"  [{status}] {method:<4} {endpoint}  HTTP={code}  {note}")


def skip(skill: str, endpoint: str, reason: str) -> None:
    results.append(Result(skill, endpoint, "SKIP", None, reason))
    print(f"  [SKIP] {endpoint}  {reason}")


def mentor_smoke() -> None:
    query = "Smoke test: answer in one short sentence, what is molecular dynamics?"
    message_id = str(uuid.uuid4())
    answer_id = str(uuid.uuid4())
    payload = {
        "query": query,
        "model": "reason",
        "discipline": "All",
        "scene": "adk_science_navigator",
        "journal_type": "foreign",
        "snp_version": "1.0.0",
        "resource_id_list": [],
        "SNPReq": {
            "sessionId": "",
            "channel": {
                "schema": "fe",
                "version": "v1",
                "role": "user",
                "auth": 1,
                "messageId": message_id,
                "answerId": answer_id,
                "uiInfo": {
                    "layout": "main",
                    "type": "ui",
                    "subType": "@bohrium-chat/common/markdown",
                    "content": {"text": query},
                    "actionList": [{"key": "text", "action": "append"}],
                },
                "entities": [],
                "state": {},
                "meta": {},
            },
            "system": {
                "payload": {
                    "model": "reason",
                    "agentId": "science_navigator",
                    "sessionId": "",
                    "scene": "adk_science_navigator",
                    "streaming": True,
                    "biz": {
                        "uploadList": [],
                        "sn": {"discipline": "All", "journal_type": "foreign", "model": "reason"},
                    },
                }
            },
        },
    }

    endpoint = "/v2/sigma-search/api/v4/ai_search/sessions"
    code, data = http("POST", endpoint, body=payload)
    status, note = classify(code, data)
    if status == "PASS":
        session_id = ((data.get("data") or {}).get("sessionId") or "") if isinstance(data, dict) else ""
        if not session_id:
            status, note = "FAIL", f"no sessionId; body={json.dumps(data)[:160]}"
        else:
            detail_code, detail = http("GET", f"{endpoint}/{session_id}")
            detail_status, detail_note = classify(detail_code, detail)
            if detail_status != "PASS":
                status, note = "FAIL", f"detail failed HTTP={detail_code} {detail_note}"
            else:
                note = f"sessionId={session_id[:8]}..."
    results.append(Result("mentor", endpoint, status, code, note))
    print(f"  [{status}] POST {endpoint}  HTTP={code}  {note}")


def _run_sdbx(args: list[str], *, timeout: int = 600) -> tuple[int, dict, str]:
    python = os.environ.get("BOHR_SANDBOX_PYTHON", sys.executable)
    script = ROOT / "zh" / "bohrium-sandbox" / "sdbx.py"
    env = os.environ.copy()
    env.setdefault("BOHRIUM_ACCESS_KEY", AK)
    completed = subprocess.run(
        [python, str(script), *args],
        text=True,
        capture_output=True,
        timeout=timeout,
        check=False,
        env=env,
    )
    raw = (completed.stdout or completed.stderr or "").strip()
    try:
        data = json.loads(completed.stdout) if completed.stdout.strip() else {}
    except json.JSONDecodeError:
        data = {"_raw": raw[:400]}
    return completed.returncode, data, raw


def sandbox_smoke() -> None:
    endpoint = "lbg sdbx create/exec/kill"
    template = os.environ.get("BOHR_SANDBOX_TEMPLATE", "doc-compiler")
    sandbox_id = ""
    status = "FAIL"
    note = ""

    try:
        rc, created, raw = _run_sdbx(["create", template, "--timeout", "600", "--json"])
        if rc != 0:
            note = f"create failed rc={rc}; {raw[:120]}"
            return
        sandbox_id = created.get("sandboxID", "")
        if not sandbox_id:
            note = f"create returned no sandboxID; body={json.dumps(created)[:160]}"
            return

        command = "echo bohrium-sandbox-smoke && python --version"
        rc, executed, raw = _run_sdbx(["exec", "--json", sandbox_id, command])
        if rc != 0 or executed.get("exit_code") != 0:
            note = f"exec failed rc={rc} exit={executed.get('exit_code')}; {raw[:120]}"
            return
        stdout = executed.get("stdout", "")
        if "bohrium-sandbox-smoke" not in stdout:
            note = f"missing smoke marker; stdout={stdout[:120]}"
            return
        status = "PASS"
        note = f"sandboxID={sandbox_id[:18]}...  exec=0"
    finally:
        if sandbox_id:
            rc, killed, raw = _run_sdbx(["kill", "--force", "--json", sandbox_id])
            if status == "PASS" and (rc != 0 or not killed.get("killed")):
                status = "FAIL"
                note = f"kill failed rc={rc}; {raw[:120]}"
        results.append(Result("sandbox", endpoint, status, 0 if status == "PASS" else None, note))
        print(f"  [{status}] {endpoint}  {note}")


# ---------------------------------------------------------------------------

print("=" * 72)
print(f"Bohrium skill smoke test")
print(f"BASE = {BASE}")
print(f"AK   = {AK[:6]}...{AK[-4:]} (len={len(AK)})")
print("=" * 72)


# ---------------------------------------------------------------------------
# bohrium-job  (stays on v1: v2 upstream differs, job_group has no v2 route)
# ---------------------------------------------------------------------------
print("\n[bohrium-job]")
record("job", "/v1/job/list", "GET", params={"page": 1, "pageSize": 1})

# ---------------------------------------------------------------------------
# bohrium-node
# ---------------------------------------------------------------------------
print("\n[bohrium-node]")
record("node", "/v2/node/list", "GET", params={"page": 1, "pageSize": 1})

# ---------------------------------------------------------------------------
# bohrium-dataset
# ---------------------------------------------------------------------------
print("\n[bohrium-dataset]")
record("dataset", "/v2/ds/", "GET", params={"page": 1, "pageSize": 1})

# ---------------------------------------------------------------------------
# bohrium-file  (personal/share list/download use v1; upload credential uses v2)
# ---------------------------------------------------------------------------
print("\n[bohrium-file]")
code, ak_data = http("GET", "/v1/ak/get")
status, note = classify(code, ak_data)
user_id = None
if status == "PASS" and isinstance(ak_data, dict):
    payload = ak_data.get("data") if isinstance(ak_data.get("data"), dict) else {}
    user_id = payload.get("user_id") or payload.get("userId")
    if not user_id:
        status, note = "FAIL", f"no user_id; body={json.dumps(ak_data)[:160]}"
results.append(Result("file", "/v1/ak/get", status, code, note))
print(f"  [{status}] GET  /v1/ak/get  HTTP={code}  {note}")
if user_id:
    record_file(
        "/v1/file/iterate",
        "POST",
        body={
            "prefix": "personal/",
            "projectId": 0,
            "userId": user_id,
            "dirSpace": "personal",
            "maxObjects": 1,
            "nextToken": "",
        },
        require_data_keys=("objects", "hasNext"),
    )
    record_file(
        "/v1/file/stat/personal/",
        "GET",
        params={"projectId": 0, "userId": user_id},
        require_data_keys=("exist",),
    )
    record_file(
        "/v2/file/upload/binary",
        "GET",
        params={"projectId": 0, "userId": user_id, "path": "/personal/bohrium-file-smoke.txt"},
        require_data_keys=("host", "Authorization", "X-Storage-Param"),
        note_on_pass="credential issued; upload bytes not sent",
    )

# ---------------------------------------------------------------------------
# bohrium-image  — image v2 endpoints live on open-platform
# ---------------------------------------------------------------------------
print("\n[bohrium-image]")
IMAGE_BASE = "https://open.bohrium.com/openapi"


def record_image() -> None:
    url = IMAGE_BASE + "/v2/image/public?page=1&pageSize=1"
    req = urllib.request.Request(
        url, headers={"Authorization": f"Bearer {AK}", "Content-Language": LANG_HEADER}
    )
    try:
        with urllib.request.urlopen(req, timeout=TIMEOUT, context=SSL_CTX) as resp:
            code = resp.status
            data = json.loads(resp.read().decode("utf-8", errors="replace") or "{}")
    except urllib.error.HTTPError as e:
        code = e.code
        try:
            data = json.loads(e.read().decode("utf-8", errors="replace") or "{}")
        except json.JSONDecodeError:
            data = {}
    except Exception as e:  # noqa: BLE001
        code = 0
        data = {"_err": repr(e)}
    status, note = classify(code, data)
    note = (note + " (via open.bohrium.com)").strip()
    results.append(Result("image", "/v2/image/public", status, code, note))
    print(f"  [{status}] GET  /v2/image/public  HTTP={code}  {note}")


record_image()

# ---------------------------------------------------------------------------
# bohrium-project
# ---------------------------------------------------------------------------
print("\n[bohrium-project]")
record("project", "/v2/project/lite_list", "GET")

# ---------------------------------------------------------------------------
# bohrium-knowledge-base
# ---------------------------------------------------------------------------
print("\n[bohrium-knowledge-base]")
record(
    "knowledge-base",
    "/v2/knowledge/knowledge_base/list",
    "GET",
    params={"page": 1, "pageSize": 1},
)

# ---------------------------------------------------------------------------
# bohrium-paper-search  (v2 paper/rag bills account balance — already billed)
# ---------------------------------------------------------------------------
print("\n[bohrium-paper-search]")
record(
    "paper-search",
    "/v2/paper/rag/pass/keyword",
    "POST",
    body={"words": ["graphene"], "question": "graphene synthesis", "type": 0, "pageSize": 2},
)
record(
    "paper-search",
    "/v2/paper/rag/pass/patent",
    "POST",
    body={"type": 1, "words": ["neural network"], "question": "neural network", "pageSize": 2},
)

# ---------------------------------------------------------------------------
# bohrium-pdf-parser  (v2 parse bills per-page balance on trigger; was v1 quota)
# ---------------------------------------------------------------------------
print("\n[bohrium-pdf-parser]")
# Trigger a cheap single-page parse and check we at least got a token back.
code, data = http(
    "POST",
    "/v2/parse/trigger-url-async",
    body={
        "url": "https://arxiv.org/pdf/2107.06922",
        "sync": False,
        "textual": True,
        "table": False,
        "molecule": False,
        "chart": False,
        "figure": False,
        "expression": False,
        "equation": False,
        "pages": [0],
        "timeout": 300,
    },
)
if 200 <= code < 300 and isinstance(data, dict):
    inner = data.get("data") if isinstance(data.get("data"), dict) else {}
    token = data.get("token") or inner.get("token")
    if token:
        status = "PASS"
        note = f"token={token[:8]}...  status={inner.get('status') or data.get('status')}"
    else:
        status = "FAIL"
        note = f"no token; body={json.dumps(data)[:160]}"
else:
    status = "FAIL"
    note = f"no token; body={json.dumps(data)[:160]}"
results.append(Result("pdf-parser", "/v2/parse/trigger-url-async", status, code, note))
print(f"  [{status}] POST /v2/parse/trigger-url-async  HTTP={code}  {note}")

# ---------------------------------------------------------------------------
# bohrium-web-search  (v2 is free — v1 quota middleware removed)
# ---------------------------------------------------------------------------
print("\n[bohrium-web-search]")
record(
    "web-search",
    "/v2/search/web",
    "GET",
    params={"q": "deepmd-kit", "num": 3},
)

# ---------------------------------------------------------------------------
# bohrium-scholar-search
# ---------------------------------------------------------------------------
print("\n[bohrium-scholar-search]")
record(
    "scholar-search",
    "/v2/paper-server/scholar/search",
    "POST",
    body={"name": "Yann LeCun", "page": 1, "pageSize": 3},
)

# ---------------------------------------------------------------------------
# bohrium-sciencepedia (formerly bohrium-wiki)
#
# Chains through the four documented user tasks (search / browse / course
# structure / knowledge graph) using ids resolved from previous responses, so a
# doc–gateway drift on any of those endpoints surfaces immediately. All paths
# touched here are free / read-only.
# ---------------------------------------------------------------------------
print("\n[bohrium-sciencepedia]")


def sciencepedia_smoke() -> None:
    wiki_base = "/v2/literature-sage/wiki_v2"
    defaults = {"language": "en-US", "style": "Feynman"}

    def push(skill: str, endpoint: str, method: str, code: int, status: str, note: str) -> None:
        results.append(Result(skill, endpoint, status, code, note))
        print(f"  [{status}] {method:<4} {endpoint}  HTTP={code}  {note}")

    # 0) /info — service stats (no body, no auth params beyond bearer).
    record("sciencepedia", f"{wiki_base}/info", "GET")

    # 1) /search/universal — unified search; capture an article id + field id.
    search_endpoint = f"{wiki_base}/search/universal"
    code, data = http("POST", search_endpoint, body={"text": "graphene", **defaults})
    status, note = classify(code, data)
    entry_id: str | None = None
    field_id: str | None = None
    if status == "PASS":
        payload = data.get("data") if isinstance(data, dict) else None
        articles = (payload or {}).get("articles") or []
        fields = (payload or {}).get("fields") or []
        for a in articles:
            if a.get("type") == "article" and a.get("id"):
                entry_id = a["id"]
                break
        for f in fields:
            if f.get("field_id"):
                field_id = f["field_id"]
                break
        note = f"articles={len(articles)} fields={len(fields)}"
    push("sciencepedia", search_endpoint, "POST", code, status, note)

    # 1') /article — fetch full body for the article id from search.
    article_endpoint = f"{wiki_base}/article"
    if entry_id:
        code, data = http("POST", article_endpoint, body={"entry_id": entry_id, **defaults})
        status, note = classify(code, data)
        # 250002 = "content not yet generated for this language/style" — documented
        # fallback path, not a regression of the endpoint itself.
        api_code = data.get("code") if isinstance(data, dict) else None
        if status == "FAIL" and api_code == 250002:
            status = "PASS"
            note = f"tolerated api code=250002 (content not generated for entry_id={entry_id[:24]})"
        elif status == "PASS":
            note = f"entry_id={entry_id[:24]}"
        push("sciencepedia", article_endpoint, "POST", code, status, note)
    else:
        skip("sciencepedia", article_endpoint, "no entry_id from /search/universal")

    # 2) /major_levels — discipline tree; collect a few level_ids for /level_fields.
    ml_endpoint = f"{wiki_base}/major_levels"
    code, data = http("POST", ml_endpoint, body=defaults)
    status, note = classify(code, data)
    level_ids: list[str] = []
    if status == "PASS":
        majors = ((data.get("data") if isinstance(data, dict) else None) or {}).get("majors") or []
        for m in majors:
            for lv in m.get("levels") or []:
                if lv.get("node_id"):
                    level_ids.append(lv["node_id"])
        note = f"majors={len(majors)} levels={len(level_ids)}"
    push("sciencepedia", ml_endpoint, "POST", code, status, note)

    # 2') /level_fields — paged courses under those level ids.
    lf_endpoint = f"{wiki_base}/level_fields"
    if level_ids:
        record(
            "sciencepedia",
            lf_endpoint,
            "POST",
            body={
                "node_ids": level_ids[:3],
                "page_num": 1,
                "page_size": 5,
                **defaults,
            },
        )
    else:
        skip("sciencepedia", lf_endpoint, "no level_ids from /major_levels")

    # 3) /get_wiki_index — chapter tree for the course found in search.
    gwi_endpoint = f"{wiki_base}/get_wiki_index"
    if field_id:
        record(
            "sciencepedia",
            gwi_endpoint,
            "POST",
            body={"field_id": field_id, **defaults},
            note_on_pass=f"field_id={field_id[:24]}",
        )
    else:
        skip("sciencepedia", gwi_endpoint, "no field_id from /search/universal")

    # 4) /knowledge_graph — GET, single id (the entry from task 1).
    kg_endpoint = f"{wiki_base}/knowledge_graph"
    if entry_id:
        record(
            "sciencepedia",
            kg_endpoint,
            "GET",
            params={"id": entry_id, **defaults},
            note_on_pass=f"id={entry_id[:24]}",
        )
    else:
        skip("sciencepedia", kg_endpoint, "no entry_id from /search/universal")

    # Backward-compat: original single-endpoint smoke kept for regression.
    record(
        "sciencepedia",
        f"{wiki_base}/search_index_name",
        "POST",
        body={"name": "graphene", "node_types": ["field"], "style": "Feynman"},
    )


sciencepedia_smoke()

# ---------------------------------------------------------------------------
# bohrium-tools  (LKM tool library — domain is free; hybrid is billed like old wiki_tool)
# Responses use the {code, data, trace_id} envelope; data lives under "data".
# ---------------------------------------------------------------------------
print("\n[bohrium-tools]")
record("tools", "/v2/lkm/tool/domain", "GET")
record(
    "tools",
    "/v2/lkm/tool/search/hybrid",
    "POST",
    body={
        "text": "molecular dynamics simulation",
        "keywords": {"molecular dynamics": 1.0},
        "language": "en-US",
        "k": 5,
    },
)

# ---------------------------------------------------------------------------
# bohrium-mentor (Sigma deep search — creates a billable session)
# ---------------------------------------------------------------------------
print("\n[bohrium-mentor]")
mentor_smoke()

# ---------------------------------------------------------------------------
# bohrium-sandbox (billable create/exec/kill via lbg sdbx)
# ---------------------------------------------------------------------------
print("\n[bohrium-sandbox]")
sandbox_smoke()


# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
print("\n" + "=" * 72)
print("SUMMARY")
print("=" * 72)
print(f"{'Skill':<22} {'Endpoint':<50} {'Status':<6} {'HTTP':<5} Note")
print("-" * 100)
for r in results:
    ep = r.endpoint if len(r.endpoint) <= 49 else r.endpoint[:46] + "..."
    code = "-" if r.code is None else str(r.code)
    note = r.note if len(r.note) <= 40 else r.note[:37] + "..."
    print(f"{r.skill:<22} {ep:<50} {r.status:<6} {code:<5} {note}")

passes = sum(r.status == "PASS" for r in results)
fails = sum(r.status == "FAIL" for r in results)
skips = sum(r.status == "SKIP" for r in results)
total = len(results)
print("-" * 100)
print(f"PASS={passes}  FAIL={fails}  SKIP={skips}  TOTAL={total}")

sys.exit(1 if fails else 0)
