#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Upload a paper PDF to LKM /parse/task, poll status, then fetch the result.

Uses BOHR_ACCESS_KEY from the environment.
Base URL: https://open.bohrium.com/openapi/v2/lkm

Price: 1 CNY; 0.1 CNY on cache hit.
partial is a non-retryable business failure; failed may be resubmitted.

Usage:
  python3 parse_paper.py paper.pdf
  python3 parse_paper.py paper.pdf --format graph --out result.json
  python3 parse_paper.py paper.pdf --interval 5 --timeout 1800 --out result.json

Stdout is the final JSON envelope (pretty). Progress goes to stderr.
"""

from __future__ import annotations

import argparse
import json
import math
import os
import sys
import tempfile
import time
from pathlib import Path

import requests

BASE = "https://open.bohrium.com/openapi/v2/lkm"
TERMINAL = {"succeeded", "partial", "failed"}
NOT_READY = 290017
BILLING_IN_PROGRESS = 7002
MAX_UPLOAD_BYTES = 64 * 1024 * 1024
RESULT_FORMATS = {"local", "graph"}


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


def get_result(task_id: str, result_format: str = "local") -> dict:
    params = {}
    if result_format and result_format != "local":
        params["format"] = result_format
    last = None
    for attempt in range(4):
        resp = requests.get(
            f"{BASE}/parse/task/{task_id}/result",
            headers=_auth_headers(),
            params=params or None,
            timeout=120,
        )
        last = _envelope(resp)
        if last.get("code") != BILLING_IN_PROGRESS:
            return last
        wait = 2 ** attempt
        print(
            f"result not available yet (code={BILLING_IN_PROGRESS}); retry in {wait}s",
            file=sys.stderr,
        )
        time.sleep(wait)
    return last


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


def _validate_inputs(pdf: Path, out: Path | None, interval: float, timeout: float) -> str | None:
    if not pdf.is_file():
        return f"PDF not found: {pdf}"
    if pdf.suffix.lower() != ".pdf":
        return f"Input must use the .pdf extension: {pdf}"
    size = pdf.stat().st_size
    if size == 0:
        return f"PDF is empty: {pdf}"
    if size > MAX_UPLOAD_BYTES:
        return f"PDF is {size} bytes; the upload limit is 64 MiB"
    try:
        with pdf.open("rb") as fh:
            if fh.read(5) != b"%PDF-":
                return f"Input does not contain a valid PDF header: {pdf}"
    except OSError as exc:
        return f"Cannot read PDF {pdf}: {exc}"
    if not math.isfinite(interval) or not 0 < interval <= 300:
        return "--interval must be a finite value between 0 and 300 seconds"
    if not math.isfinite(timeout) or not 0 < timeout <= 86400:
        return "--timeout must be a finite value between 0 and 86400 seconds"
    if out is not None and out.expanduser().resolve() == pdf.resolve():
        return "--out must not overwrite the input PDF"
    return None


def _write_text_atomic(path: Path, text: str) -> None:
    destination = path.expanduser()
    destination.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix=".lkm-parse-", dir=destination.parent)
    temporary_path = Path(temporary)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as fh:
            fh.write(text)
        os.replace(temporary_path, destination)
    except Exception:
        temporary_path.unlink(missing_ok=True)
        raise


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("pdf", type=Path, help="Local PDF to upload (field name must be file; default max 64 MiB / 50 pages)")
    parser.add_argument("--interval", type=float, default=5.0, help="Poll interval in seconds (default 5)")
    parser.add_argument("--timeout", type=float, default=1800.0, help="Max wait in seconds (default 1800)")
    parser.add_argument(
        "--format",
        dest="result_format",
        choices=sorted(RESULT_FORMATS),
        default="local",
        help="Result shape: local (default flat graph) or graph (same as /papers/graph)",
    )
    parser.add_argument("--out", type=Path, help="Write the result envelope to PATH instead of stdout")
    args = parser.parse_args(argv)

    pdf = args.pdf.expanduser()
    validation_error = _validate_inputs(pdf, args.out, args.interval, args.timeout)
    if validation_error is not None:
        print(validation_error, file=sys.stderr)
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

    result = get_result(task_id, args.result_format)
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
        _write_text_atomic(args.out, text + "\n")
        print(f"wrote {args.out}", file=sys.stderr)
    result_data = result.get("data")
    result_status = result_data.get("status") if isinstance(result_data, dict) else None
    reason = result_data.get("failed_reason") if isinstance(result_data, dict) else None
    if status == "partial" or result_status == "partial":
        print(
            f"extraction partial (non-retryable business failure): {reason or 'no reason reported'}",
            file=sys.stderr,
        )
        return 0
    if status == "failed" or result_status == "failed":
        print(f"extraction failed: {reason or 'no reason reported'}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
