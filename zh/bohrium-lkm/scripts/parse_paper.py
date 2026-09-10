#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Submit a PDF and/or parser markdown to LKM /parse/task, poll, then fetch result.

Uses BOHR_ACCESS_KEY from the environment.
Base URL: https://open.bohrium.com/openapi/v2/lkm

Give at least a PDF or --content. --content is posted as the API content text
field (not a second file part) and skips LAS OCR. If you already have
--content and also have the PDF, pass the PDF too: that raises the chance
of a cache hit. --md5 is the PDF digest; use it on content-only submits.
--page is optional. When a PDF is given, do not send md5 or page.

Price: 1 CNY; 0.1 CNY on cache hit.
partial is a non-retryable business failure; failed may be resubmitted.

Usage:
  python3 parse_paper.py paper.pdf
  python3 parse_paper.py paper.pdf --content paper.md
  python3 parse_paper.py --content paper.md --md5 <32-hex> --page 12
  python3 parse_paper.py paper.pdf --format graph --out result.json

Stdout is the final JSON envelope (pretty). Progress goes to stderr.
"""

from __future__ import annotations

import argparse
import json
import math
import os
import re
import sys
import tempfile
import time
from pathlib import Path

import requests

BASE = "https://open.bohrium.com/openapi/v2/lkm"
TERMINAL = {"succeeded", "partial", "failed"}
NOT_READY = 290017
MAX_UPLOAD_BYTES = 64 * 1024 * 1024
RESULT_FORMATS = {"local", "graph"}
CONTENT_EXTS = {".md", ".txt", ".markdown"}
MD5_RE = re.compile(r"^[0-9a-fA-F]{32}$")


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


def _looks_like_content_path(raw: str) -> bool:
    if "/" in raw or "\\" in raw:
        return True
    return Path(raw).suffix.lower() in CONTENT_EXTS


def _read_content_file(path: Path) -> str:
    if not path.is_file():
        raise SystemExit(f"content file not found or not a regular file: {path}")
    size = path.stat().st_size
    if size == 0:
        raise SystemExit(f"content file is empty: {path}")
    if size > MAX_UPLOAD_BYTES:
        raise SystemExit(f"content file is {size} bytes; the upload limit is 64 MiB")
    text = path.read_text(encoding="utf-8")
    if not text.strip():
        raise SystemExit(f"content file is empty: {path}")
    return text


def resolve_content(raw: str) -> str:
    text = raw.strip()
    if not text:
        raise SystemExit("--content is empty")
    if text.startswith("@"):
        path = text[1:].strip()
        if not path:
            raise SystemExit("--content @file is missing a path")
        return _read_content_file(Path(path).expanduser())
    if _looks_like_content_path(text):
        return _read_content_file(Path(text).expanduser())
    candidate = Path(text).expanduser()
    try:
        if candidate.is_file():
            return _read_content_file(candidate)
    except OSError:
        pass
    encoded = text.encode("utf-8")
    if len(encoded) > MAX_UPLOAD_BYTES:
        raise SystemExit(f"content is {len(encoded)} bytes; the upload limit is 64 MiB")
    return text


def submit(
    pdf: Path | None,
    content: str | None,
    digest: str | None,
    page: int | None,
) -> dict:
    data: dict[str, str] = {}
    if content:
        data["content"] = content
    if pdf is None:
        if digest:
            data["md5"] = digest
        if page is not None:
            data["page"] = str(page)

    headers = _auth_headers()
    if pdf is None:
        resp = requests.post(
            f"{BASE}/parse/task",
            headers=headers,
            data=data or None,
            timeout=120,
        )
    else:
        with pdf.open("rb") as fh:
            resp = requests.post(
                f"{BASE}/parse/task",
                headers=headers,
                files={"file": (pdf.name, fh, "application/pdf")},
                data=data or None,
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
    resp = requests.get(
        f"{BASE}/parse/task/{task_id}/result",
        headers=_auth_headers(),
        params=params or None,
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


def _validate_pdf(pdf: Path) -> str | None:
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
    return None


def _validate_inputs(
    pdf: Path | None,
    content_raw: str | None,
    digest: str | None,
    page: int | None,
    out: Path | None,
    interval: float,
    timeout: float,
) -> str | None:
    if pdf is None and not (content_raw or "").strip():
        return "requires a PDF argument, or --content"
    if pdf is not None:
        err = _validate_pdf(pdf)
        if err:
            return err
        if digest or page is not None:
            print("md5/page are unused when a PDF is given; not sending them", file=sys.stderr)
    if digest:
        if not MD5_RE.fullmatch(digest.strip()):
            return "md5 must be a 32-char hex digest"
    if page is not None and page < 0:
        return "--page must be >= 0"
    if not math.isfinite(interval) or not 0 < interval <= 300:
        return "--interval must be a finite value between 0 and 300 seconds"
    if not math.isfinite(timeout) or not 0 < timeout <= 86400:
        return "--timeout must be a finite value between 0 and 86400 seconds"
    if pdf is not None and out is not None and out.expanduser().resolve() == pdf.resolve():
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
    parser.add_argument(
        "pdf",
        nargs="?",
        type=Path,
        default=None,
        help="Optional local PDF (field name file; max 64 MiB; 50-page reject is PDF-only)",
    )
    parser.add_argument(
        "--content",
        help="API content field: path, @file, or literal markdown (file is read locally). Skips LAS OCR. Also pass the PDF when you have it; that raises cache-hit chance",
    )
    parser.add_argument(
        "--md5",
        dest="digest",
        help="PDF MD5 (32-char hex). Content-only: you may pass it. Unused when a PDF is given",
    )
    parser.add_argument(
        "--page",
        type=int,
        default=None,
        help="Optional. Unused when a PDF is given",
    )
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

    pdf = args.pdf.expanduser() if args.pdf is not None else None
    validation_error = _validate_inputs(
        pdf, args.content, args.digest, args.page, args.out, args.interval, args.timeout
    )
    if validation_error is not None:
        print(validation_error, file=sys.stderr)
        return 2

    content = resolve_content(args.content) if args.content else None
    digest = args.digest.strip().lower() if args.digest else None

    submitted = submit(pdf, content, digest, args.page)
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
