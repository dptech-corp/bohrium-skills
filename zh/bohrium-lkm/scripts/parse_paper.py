#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Upload a paper PDF to LKM /parse/task, poll status, then fetch the result.

Uses BOHR_ACCESS_KEY from the environment.
Base URL: https://open.bohrium.com/openapi/v2/lkm

Usage:
  python3 parse_paper.py paper.pdf
  python3 parse_paper.py paper.pdf --interval 5 --timeout 1800 --out result.json

Stdout is the final JSON envelope (pretty). Progress goes to stderr.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import time
from pathlib import Path

import requests

BASE = "https://open.bohrium.com/openapi/v2/lkm"
TERMINAL = {"succeeded", "partial", "failed"}
NOT_READY = 290017


def _auth_headers() -> dict[str, str]:
    ak = (os.environ.get("BOHR_ACCESS_KEY") or "").strip()
    if not ak:
        raise SystemExit("Missing BOHR_ACCESS_KEY")
    return {"Authorization": f"Bearer {ak}"}


def _envelope(resp: requests.Response) -> dict:
    try:
        body = resp.json()
    except ValueError as exc:
        raise SystemExit(f"LKM returned non-JSON (HTTP {resp.status_code})") from exc
    if not isinstance(body, dict):
        raise SystemExit(f"LKM returned unexpected payload type: {type(body).__name__}")
    return body


def submit(pdf: Path) -> dict:
    with pdf.open("rb") as fh:
        resp = requests.post(
            f"{BASE}/parse/task",
            headers=_auth_headers(),
            files={"file": (pdf.name, fh, "application/pdf")},
            timeout=120,
        )
    body = _envelope(resp)
    if body.get("code") != 0:
        raise SystemExit(f"submit failed code={body.get('code')}: {body.get('message') or body.get('error')}")
    return body["data"]


def get_status(task_id: str) -> dict:
    resp = requests.get(
        f"{BASE}/parse/task/{task_id}",
        headers=_auth_headers(),
        timeout=60,
    )
    body = _envelope(resp)
    if body.get("code") != 0:
        raise SystemExit(f"status failed code={body.get('code')}: {body.get('message') or body.get('error')}")
    return body["data"]


def get_result(task_id: str) -> dict:
    resp = requests.get(
        f"{BASE}/parse/task/{task_id}/result",
        headers=_auth_headers(),
        timeout=120,
    )
    return _envelope(resp)


def wait_for_terminal(task_id: str, interval: float, timeout: float) -> dict:
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        last = get_status(task_id)
        status = last.get("status")
        stage = last.get("stage")
        print(f"status={status} stage={stage}", file=sys.stderr)
        step_durations = last.get("step_durations")
        if step_durations:
            print(
                f"step_durations={json.dumps(step_durations, ensure_ascii=False)}",
                file=sys.stderr,
            )
        if status in TERMINAL:
            return last
        time.sleep(interval)
    raise SystemExit(
        f"timed out after {timeout:.0f}s waiting for task {task_id}; last={last}"
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("pdf", type=Path, help="Local PDF to upload (field name must be file)")
    parser.add_argument("--interval", type=float, default=5.0, help="Poll interval in seconds (default 5)")
    parser.add_argument("--timeout", type=float, default=1800.0, help="Max wait in seconds (default 1800)")
    parser.add_argument("--out", type=Path, help="Write the result envelope to PATH instead of stdout")
    args = parser.parse_args(argv)

    pdf = args.pdf.expanduser()
    if not pdf.is_file():
        print(f"PDF not found: {pdf}", file=sys.stderr)
        return 2

    submitted = submit(pdf)
    task_id = submitted["task_id"]
    cache_hit = bool(submitted.get("cache_hit"))
    cache_source = submitted.get("cache_source")
    status = submitted.get("status")
    print(
        f"submitted task_id={task_id} cache_hit={cache_hit} "
        f"cache_source={cache_source} status={status}",
        file=sys.stderr,
    )

    if not (cache_hit and status in TERMINAL):
        last = wait_for_terminal(task_id, args.interval, args.timeout)
        status = last.get("status")

    result = get_result(task_id)
    code = result.get("code")
    if code == NOT_READY:
        print(
            f"result not ready (code={NOT_READY}); poll GET /parse/task/{{task_id}} and retry",
            file=sys.stderr,
        )
        return 1
    if code != 0:
        print(f"result failed code={code}: {result.get('message') or result.get('error')}", file=sys.stderr)
        return 1

    text = json.dumps(result, ensure_ascii=False, indent=2)
    if args.out is None:
        print(text)
    else:
        args.out.write_text(text + "\n", encoding="utf-8")
        print(f"wrote {args.out}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
