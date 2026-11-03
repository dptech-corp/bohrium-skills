# PDF parse command reference

This reference describes the PDF parsing commands embedded in the same `bohr`
version as this skill.

## Submit a parse

```text
bohr pdf parse (--url <pdf-url> | --file <local-path>) [flags]
```

Exactly one input is required. `--file` uploads the selected file as multipart
data; `--url` sends the selected URL for server-side download.

| Flag | Default | Accepted values and effect |
|---|---:|---|
| `--url <value>` | empty | PDF URL; mutually exclusive with `--file` |
| `--file <path>` | empty | Existing local file; mutually exclusive with `--url` |
| `--textual <n>` | `1` | Text extraction: `0` off, `1` fast, `2` accurate, `3` direct text extraction |
| `--table <n>` | `1` | Table extraction: `0` off, `1` fast, `2` accurate |
| `--equation <n>` | `0` | Equation extraction: `0` off, `1` fast, `2` accurate |
| `--chart <n>` | `0` | Chart extraction: `0` off, `1` on |
| `--figure <n>` | `0` | Figure extraction: `-1` source image, `0` off, `1` on |
| `--molecule <n>` | `0` | Molecule extraction: `0` off, `1` on |
| `--sync` | false | Wait during submission rather than returning immediately |
| `--dry-run` | false | Print the planned input route without sending it |
| `--yes` | false | Confirm the billable submission after authorization |

The CLI forwards the integer extraction modes to Uni-Parser. Keep values inside
the documented ranges; invalid values are rejected upstream rather than
normalized locally.

Examples:

```bash
bohr pdf parse --url "https://example.org/paper.pdf" --yes -o json
bohr pdf parse --file ./paper.pdf --textual 2 --table 2 --equation 2 --yes -o json
bohr pdf parse --file ./paper.pdf --chart 1 --figure 1 --yes -o json
```

A successful machine response preserves the upstream submission data. Capture
`data.token`; common accompanying fields include `data.status` and
`data.page_count`.

## Billing and data movement

```bash
bohr billing pricing --resource pdf -o json
```

The version-matched catalog price for `pdf.parse` is **0.05 CNY per page**,
triggered by a successful parse. Non-interactive callers must obtain spending
authorization before adding `--yes`. The result command is free.

Before `--file`, verify that the exact local file is intended for upload and is
appropriate to send to Bohrium. Before `--url`, verify that the exact URL is a
directly downloadable PDF and that sending it to the service is acceptable.

## Retrieve a result

```text
bohr pdf result <token> [flags]
```

| Flag | Default | Effect |
|---|---:|---|
| `--content` | true | Return formatted full-document content |
| `--objects` | true | Return flat semantic blocks and coordinates |
| `--pages-dict` | false | Return page-level flat layouts |
| `--pages-tree` | false | Return page-level hierarchical layouts |
| `--return-half` | false | Return completed portions while parsing continues |
| `--molecule-source` | false | Preserve original molecule source data |
| `--formatted <format>` | empty | Format semantic content as `plain`, `markup`, `markdown`, `latex`, or `html` |

At least one primary output flag must be enabled. These are valid focused calls:

```bash
bohr pdf result <token> --content --objects -o json
bohr pdf result <token> --content=false --objects=false --pages-dict -o json
bohr pdf result <token> --content=false --objects=false --pages-tree -o json
bohr pdf result <token> --formatted markdown -o json
```

`--formatted` selects the formatted result route and applies one format to text,
tables, charts, figures, expressions, equations, and molecules. It changes
semantic text representation; it does not validate extracted values.

## Status and polling

Use structured output and inspect `data.status`:

| Status or error | Meaning | Action |
|---|---|---|
| `undefined` | Queued or not yet producing output | Continue bounded polling |
| `processing` | Parsing is in progress | Continue bounded polling |
| `success` | Terminal success | Consume the requested result fields |
| `PARSE_FAILED` | Terminal parser failure or error status | Stop and report the returned description |

Poll every 3–5 seconds rather than issuing a tight request loop. Set a waiting
budget proportional to the document size. Reaching the local budget leaves the
remote task unchanged; retain the token so polling can resume later.

The `--sync` submission mode waits inside the submit call. Still retain the
token and call `bohr pdf result` to request content, flat objects, page layouts,
or formatted output.

## Result fields

Machine output preserves upstream data rather than rewriting its schema. Common
fields include:

| Field | Meaning |
|---|---|
| `data.token` | Parse token used for later reads |
| `data.status` | Current task state |
| `data.content` | Extracted document text/markup when requested |
| `data.objects` | Flat semantic blocks such as text, table, chart, legend, or equation |
| `data.pages_dict` | Page-level flat layout array when requested |
| `data.pages_tree` | Page-level hierarchical layout array when requested |
| `data.proc_page` / `data.total_page` | Processed and total page counts |
| `data.description` / `data.error` | Failure detail when supplied upstream |

`pages_dict` and `pages_tree` are arrays. Their lengths do not define progress;
use `proc_page`, `total_page`, and `status`. Preserve unknown fields when
forwarding structured results because new parser metadata may appear.

Human output summarizes the token, status, progress, content, and structured
item counts. Use `-o json` when the actual blocks, hierarchy, or vendor fields
are needed.

## Workflow boundary

`bohr pdf parse` is Uni-Parser document extraction. It returns layout, text,
tables, equations, charts, figures, or molecules. `bohr lkm extract` is a separate
knowledge-extraction workflow with different task IDs, states, and result
models. Tokens and IDs are not interchangeable.

## Version contract

The files are embedded at build time and versioned with the `bohr` binary:

```bash
bohr guide read bohr-pdf-parse
bohr guide list bohr-pdf-parse/references
bohr guide read bohr-pdf-parse/references/commands.md
```

`bohr guide` does not fetch the standalone Bohrium Skills repository. Upgrade
`bohr` to receive newer command guidance.

## Payment-source option

The optional global `--payment-source` flag accepts `photon` or `balance`.
Omitting it sends no preference and preserves photon-first/balance fallback.
Follow [Choose the payment source](../SKILL.md#choose-the-payment-source) for
funding authorization and insufficient-funds recovery. After balance is authorized:

```bash
bohr pdf parse --url "https://example.org/paper.pdf" --payment-source balance --yes -o json
```

The same option is available on `bohr api`. Free commands remain free; selecting
a source does not create a charge on an otherwise free operation.
