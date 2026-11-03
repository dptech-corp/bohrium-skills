---
name: bohrium-pdf-parser
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Parse a PDF with the `bohr pdf parse` and `bohr pdf result` CLI workflow to
  extract text, tables, equations, charts, figures, molecules, or page layout.
  Use for Bohrium Uni-Parser, PDF 解析, PDF 文本提取, or structured PDF content.
  Use `bohr lkm extract` for research questions, conclusions, and reasoning
  rather than document layout.
---

# Bohrium PDF Parse

## Choose the task

| Goal | Route |
|---|---|
| Extract a local PDF | Validate file → submit `--file` → retain token → retrieve requested blocks |
| Extract a directly reachable PDF | Validate URL → submit `--url` → use the same result workflow |
| Recover an interrupted wait | Continue `result` with the original token; submission creates another parse |
| Extract claims or research reasoning | Route to LKM extract instead of layout parsing |

## Workflow: turn one PDF into a usable artifact

1. Confirm input, requested semantic modes and spending using the sections below. Record the original source alongside the task token.
2. Submit once, then replace `<token>` with the returned `data.token`. JSON is forwarded from upstream: if the token is absent, preserve the full response and report the missing handoff; do not substitute another task ID or resubmit in human mode to obtain it:

```bash
bohr pdf parse --file ./paper.pdf --table 2 --equation 2 --yes -o json
bohr pdf result <token> --formatted markdown -o json
```

3. Inspect the returned status before using content. Continue the same token through `undefined`/`processing` at a bounded interval. `success` completes parsing; `PARSE_FAILED` ends it unsuccessfully. A timeout stops this wait and retains the token for later retrieval.
4. On success, inspect the requested text/objects and `proc_page`/`total_page`; record missing or partial pages. Deliver the requested content with page provenance and the source/token needed to trace it. Non-empty `--return-half` output remains partial until terminal success.
5. When result retrieval fails transiently, retry that free read with the same token. A failed or uncertain submission needs its returned diagnostic/token examined before resubmission; a new parse may incur another charge. Critical table/equation values require comparison with the source page before use.

Use Bohrium Uni-Parser through the version-matched CLI to extract document
layout and content from one PDF. Choose exactly one input route:

- `bohr pdf parse --file <path>` uploads a local PDF.
- `bohr pdf parse --url <url>` asks the service to download a PDF URL.

Confirm the exact local file or URL before submission. A local file leaves the
current machine, and a URL must be directly reachable by the parsing service.

This workflow extracts layout, text, tables, equations, charts, figures, or molecules.
Use `bohr lkm extract` when the requested output is scientific claims,
research questions, conclusions, reasoning steps, or an LKM variable graph.

## Authenticate and authorize parsing

Check the current identity and price contract:

```bash
bohr auth whoami -o json
bohr billing pricing --resource pdf -o json
```

PDF parsing is **0.05 CNY per page** under the embedded catalog. The final cost
depends on the successfully parsed page count, so do not invent a total before
the service reports the document size or billing result. In non-interactive
use, add `--yes` only after spending is authorized; result retrieval is free
and does not need spending confirmation.

Use `--dry-run` to check the selected input without submitting or uploading it.

## Submit the PDF

The defaults enable fast text and table extraction. Enable only the additional
semantic modes needed by the task:

```bash
bohr pdf parse --file ./paper.pdf \
  --textual 2 \
  --table 2 \
  --equation 2 \
  --chart 1 \
  --figure 0 \
  --molecule 0 \
  --yes \
  -o json
```

For a remote source:

```bash
bohr pdf parse --url "https://example.org/paper.pdf" --yes -o json
```

Capture `data.token` from the successful submission. The token, not a task ID
from another parser, is the input to `bohr pdf result`.

## Retrieve to a terminal state

Parsing is asynchronous by default:

```bash
bohr pdf result <token> -o json
```

`undefined` and `processing` are non-terminal states. Poll at a moderate,
bounded interval such as 3–5 seconds. Stop when the returned status is
`success`, when the command returns `PARSE_FAILED`, or when the user's waiting
budget expires. A local waiting deadline only stops polling; it does not cancel
the remote parse.

`--sync` makes submission wait for parsing, but the returned token is still the
stable handoff to `bohr pdf result` for requested content and structured data.

Use `--return-half` only when partial completed portions are useful. Non-empty
partial content is not evidence that parsing reached a terminal state.

## Select the result shape

The default result request enables `--content` and `--objects`. For narrower or
hierarchical output, choose explicitly:

```bash
bohr pdf result <token> --content --objects -o json
bohr pdf result <token> --content=false --objects=false --pages-dict -o json
bohr pdf result <token> --content=false --objects=false --pages-tree -o json
bohr pdf result <token> --formatted markdown -o json
bohr pdf result <token> --molecule-source -o json
```

At least one of `--content`, `--objects`, `--pages-dict`, or `--pages-tree`
must remain enabled. `pages_dict` and `pages_tree` are arrays despite their
names; use `proc_page` and `total_page` for progress instead of array length.

Structured blocks preserve upstream fields such as class/type, coordinates,
confidence, hierarchy, and future vendor fields. Treat chart and table
recognition as extracted evidence, not verified numeric truth; compare critical
values against the source PDF before calculation or data entry.

## Choose the payment source

Without `--payment-source`, the CLI sends no channel preference: the gateway
uses photons first and falls back to balance only on definite photon
insufficiency. Tell the user this policy when obtaining spending authorization.
To restrict spending to one source, use `--payment-source photon` or
`--payment-source balance`. Explicit photon stops when photons are insufficient;
explicit balance charges balance directly. Existing free quotas and exemptions
still apply.

When an explicitly selected channel returns `BILLING_INSUFFICIENT` with
`retryable: false`, report the refusal and wait
for the user to recharge the selected source or authorize the other source for
the original command. Explicit balance insufficiency also stops without charging
photons; use `--payment-source photon` only after that source is authorized.
`--yes` confirms execution; it does not authorize changing the payment source.
Use `--payment-source balance` only when that source is authorized. Keep the
selected source on subsequent billable calls in a multi-step workflow; options
from a previous CLI invocation are not saved. `bohr api` follows the same rule.

## Detailed reference

Read [`references/commands.md`](references/commands.md) for exact flag defaults,
format choices, result fields, polling states, error handling, and the embedded
version contract.
