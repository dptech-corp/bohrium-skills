# LKM command reference

## Command surface

| Command | Purpose |
| --- | --- |
| `bohr lkm search <query>` | Search ingested claims, questions, abstracts, and concepts |
| `bohr lkm reasoning --query <text>` | Search whole reasoning or experimental processes |
| `bohr lkm graph` | Get an ingested paper's knowledge graph (`--paper-id` / `--package-id` / `--doi` / `--title`) |
| `bohr lkm reference --paper-id <id>` | Get a paper's references and/or cited-by list |
| `bohr lkm variables batch <id>...` | Hydrate global variables by ID |
| `bohr lkm claim reasoning <id>` | Trace reasoning for one global conclusion claim |
| `bohr lkm extract submit <file.pdf>` | Upload a local PDF and start knowledge extraction |
| `bohr lkm extract status <task-id>` | Read extraction state and completed-step timings |
| `bohr lkm extract wait <task-id>` | Poll extraction to a terminal state |
| `bohr lkm extract result <task-id>` | Read the extracted graph or partial intermediates |

## Billing and automation

Inspect the version-matched catalog before fixed-price calls:

```bash
bohr billing pricing --resource lkm -o json
```

The current fixed price is **0.05 CNY** per `search`, `reasoning`, `graph`,
`reference`, `variables batch`, or `claim reasoning` call. LKM has a personal monthly
1,000-call quota before account charging. The command is still classified as
billable while quota remains, so a non-interactive caller needs `--yes` after
the user authorizes spending. `--dry-run` bypasses the billing confirmation and
makes no request.

Extract does not use the monthly query quota: **1.00 CNY** per successful
paper, or **0.10 CNY** on a cache hit. The same task is not charged twice.
Non-interactive use still needs `--yes` after the user authorizes spending.
`extract submit` uploads the chosen file and therefore still requires user
intent.

Prefer `-o json` for ID capture and full result interpretation. Human output is
optimized for inspection and may omit nested fields needed by a later command.

## ID flow

| ID kind | Shape and producer | Valid consumer |
| --- | --- | --- |
| Numeric paper ID | Numeric string printed by `search` or returned by `reasoning` / paper metadata | `lkm graph --paper-id`; `lkm reference --paper-id` |
| Package ID | `paper:<digits>` from paper metadata | `lkm graph --package-id` |
| DOI | DOI string from paper metadata or the user | `lkm graph --doi`; `lkm reference --doi` |
| Title | Paper title when no ID is available | `lkm graph --title` only |
| Global variable or claim ID | `gcn_...` from `search`, `reasoning`, or graph `global_id` | `variables batch`; qualifying conclusion claims also go to `claim reasoning` |
| Graph-local node ID | Often scoped to one paper graph, such as `paper:...::conclusion_3` | Relationships inside that same graph only |
| Extract task ID | Opaque value returned by `extract submit` | `extract status`, `extract wait`, `extract result` |
| Extract `local_id` | File-local node key in one extract result | `factors`, `motivations`, and other records in that same result only |
| Extract `global_id` | Optional non-null global mapping on an extracted variable | Global LKM commands when its type is compatible |

Do not infer IDs from titles. Remove a `paper:` package prefix only when the
CLI has printed the normalized numeric `Paper ID`; never strip or rewrite a
graph-local ID into a fake global ID.

## Corpus search

```bash
bohr lkm search <query> \
  [--top-k 10] \
  [--offset 0] \
  [--mode hybrid] \
  [--sort comprehensive] \
  [--keywords term,method] \
  [--scopes conclusion,abstract] \
  [--paper-id <id>]... \
  [--doi <doi>]... \
  [--title <fuzzy_title>] \
  [--publication-date-start YYYY-MM-DD] \
  [--publication-date-end YYYY-MM-DD] \
  [--limit-publication-date=true] \
  [--visibility public] \
  [--include-paper-enrich] \
  --yes -o json
```

- `<query>` accepts one or more arguments and is joined with spaces.
- `--mode` defaults to `hybrid`; also accepts `semantic` and `lexical`.
- `--sort` defaults to `comprehensive`; use `relevance`, `recent`, or
  `journal` only when that ranking goal matches the request.
- `--keywords` and `--scopes` are comma-separated. At most 10 keywords, each
  at most 100 bytes.
- `--top-k` defaults to `10` (1-100); `--offset` defaults to `0` (max 2000).
- `--paper-id` is a numeric string without a `paper:` prefix (repeatable, max 50).
  `--doi` is repeatable (max 50). `--title` is a fuzzy filter. These filters AND.
- `--limit-publication-date` defaults to true and is sent only when false.
- `--include-paper-enrich` is sent only when true. `--visibility` defaults to
  `public`.

Valid `--scopes` values are `abstract`, `claim`, `premise`, `conclusion`,
`question`, `problem`, `open_question`, `subproblem`, and `reasoning_chain`.
`claim` includes premise and conclusion;
`question` includes problem, open question, and subproblem. Use `conclusion`
when the next step needs a claim that may have traceable reasoning.

Structured results include:

- `data.variables[]`: one representative hit per grouped paper, with `id`,
  `type`, `role`, `content`, `score`, `rerank_score`, `has_reasoning`, and
  provenance when available.
- `data.related`: additional same-paper context, not cross-paper recommendations
  and not the complete paper graph.
- `data.papers`: authoritative paper metadata keyed by source package.
- `data.has_more`: whether another offset page is available.

Use an `abstract` as background context, not as a scientific claim. Question
roles describe problems or open questions, not premises. Scores rank retrieval;
they do not establish truth, confidence, or evidence strength.

## Reasoning-chain search

```bash
bohr lkm reasoning --query <process_description> \
  [--top-k 10] \
  [--offset 0] \
  [--mode hybrid] \
  [--sort comprehensive] \
  [--keywords term,method] \
  [--paper-id <id>]... \
  [--doi <doi>]... \
  [--title <fuzzy_title>] \
  [--publication-date-start YYYY-MM-DD] \
  [--publication-date-end YYYY-MM-DD] \
  [--limit-publication-date=true] \
  --yes -o json
```

This searches similarity of the whole argument, derivation, calculation, or
experimental process. Structured results include `reasoning_chains[]` with
`chain_id`, `conclusion_id`, conclusion text/title, numeric `paper_id`, and
ranking `score`. `--mode` / `--sort` accept the same values as `search`.
Paper and date filters match `search`, except `visibility` and
`--include-paper-enrich` are not on this command.

Use `conclusion_id` with `claim reasoning` and `paper_id` with `graph` or
`reference`. A chain score ranks similarity to the query and is not a
credibility score.

## Paper graph

```bash
bohr lkm graph --paper-id <numeric_paper_id> --yes -o json
bohr lkm graph --package-id paper:<numeric_paper_id> --yes -o json
bohr lkm graph --doi <doi> --yes -o json
bohr lkm graph --title "<paper_title>" --title-resolve-limit 5 --yes -o json
```

Give exactly one identifier. `--paper-id` is the numeric paper ID printed by
search or reasoning. `--package-id` is `paper:<digits>`. `--doi` and `--title`
resolve a paper already in LKM. `--title-resolve-limit` is only valid with
`--title` (default 5, max 20). Title lookup may return several `data.items`;
human output prints every item. Do not pass a global `gcn_...` ID or
graph-local node ID.

The response uses `data.items[].graph.nodes[]` and `edges[]`. A node may have a
local `id`, reusable `global_id`, `content`, `kind`, and `type`. Edge `source`
and `target` refer to local IDs within the graph.

Interpret common kinds and directions:

| Node/edge | Meaning |
| --- | --- |
| `conclusion` | A result or claim reached by the paper |
| `reasoning_steps` | Derivation, calculation, experiment, or other reasoning |
| `highlight` | Positive support, useful evidence, or strength |
| `weak_point` | Limitation, risk, or assumption that weakens a step |
| `subproblem` | Research motivation or problem being addressed |
| `concludes` | reasoning steps point to a conclusion |
| `highlight_of` | a highlight positively supports reasoning steps |
| `weakpoint_of` | a weak point limits or challenges reasoning steps |
| `subproblem_of` | a subproblem motivates a conclusion |

Preserve the difference between support, limitation, and motivation when
summarizing the graph.

## Paper references and cited-by

```bash
bohr lkm reference --paper-id <numeric_paper_id> --yes -o json
bohr lkm reference --doi <doi> --with-reference --yes -o json
```

This returns paper cards (`PaperMeta`), not a knowledge graph. Repeat
`--paper-id` and `--doi`. Combined count after trim and before dedupe must be
at most 20. The CLI always sends the three service switches, using the same
defaults as the API: `--with-abstract=true`, `--with-reference=false`,
`--with-cited-by=true`. Override with `--with-abstract=false`,
`--with-reference`, or `--with-cited-by=false`.

`references` / `cited_by` are `null` when that list was not queried, and `[]`
when it was queried and empty. Do not treat those as the same empty. An empty
`id` means LKM-uncovered data; a missing seed does not appear in `data.items`.
The v4 gateway renames upstream `papers` to `items` and adds `pagination`;
`references` / `cited_by` stay on each row.

Do not send a title or `package_id`. Strip a printed `paper:` prefix only from
`--paper-id`.

## Claim reasoning

```bash
bohr lkm claim reasoning <global_conclusion_gcn_id> \
  --format graph \
  --yes -o json
```

Use a global `gcn_...` ID for a conclusion with `has_reasoning=true`. A premise,
weak point, graph-local ID, extract `local_id`, or claim without reasoning can be
rejected even if the identifier exists.

The graph response includes the queried `claim`, `total_chains`, and
`reasoning_chains[]`. Each chain can contain paper metadata,
`addressed_problems`, `open_questions`, graph nodes, graph edges, and detailed
steps inside `reasoning_steps` nodes.

## Variable hydration

```bash
bohr lkm variables batch <global_gcn_id> [global_gcn_id...] --yes -o json
```

Use 1–100 non-empty global IDs. Remove nulls and duplicates before calling.
The response contains `variables[]`, source `papers`, and `not_found`. Partial
misses do not turn the whole response into failure; report `not_found` rather
than silently dropping requested IDs. Metadata or parameters can be absent or
encoded in heterogeneous shapes, so parse them defensively.

## Local PDF extraction

### Submit

```bash
bohr lkm extract submit <file.pdf> [-o json]
bohr lkm extract submit <file.pdf> --wait --interval 5s --timeout 30m -o json
```

The file must be regular, non-empty, have a `.pdf` extension and `%PDF` header,
and be no larger than 64 MiB. The service also enforces a 50-page limit.
Submission may take up to five minutes and returns `task_id`, `status`,
`pdf_md5`, `cache_hit`, and `cache_source`.

`cache_source=lkm` means the paper already existed in the corpus;
`cache_source=local` means the same file was extracted by an earlier submission.
A cache hit can be terminal immediately. A new submission for an already
running file returns `290020` with the existing task ID. Continue that task; a
duplicate submission does not accelerate the underlying extraction.

### Status and wait

```bash
bohr lkm extract status <task_id> -o json
bohr lkm extract wait <task_id> --interval 5s --timeout 30m -o json
```

| Status | Terminal | Action |
| --- | --- | --- |
| `queued` | no | Keep the task ID and wait |
| `running` | no | Inspect `stage` and completed `step_durations`; keep waiting |
| `succeeded` | yes | Read the result |
| `partial` | yes | Non-retryable business terminal state; inspect the reason/files and do not resubmit the same PDF |
| `failed` | yes | Technical terminal state; inspect the reason/files, then the PDF may be resubmitted |

Waiting commonly takes minutes. `LKM_PARSE_WAIT_TIMEOUT` and
`LKM_PARSE_WAIT_CANCELLED` stop only the local polling loop; either outcome
does not cancel the remote task. Resume using the same task ID.

### Result

```bash
bohr lkm extract result <task_id> --yes -o json
bohr lkm extract result <task_id> --format graph --yes -o json
```

The default `local` successful result contains matching `task_id`,
`status=succeeded`, boolean `cache_hit`, `step_durations`, and a flat
`variables` / `factors` / `motivations` / `stats` graph. Factors and
motivations refer to extract `local_id` values. Only a non-null variable
`global_id` can cross into corpus commands.

`--format graph` keeps the same billing/status fields and returns
`addressed_problems`, `open_questions`, and `graph.nodes` / `graph.edges` —
the same field kinds as `bohr lkm graph`, not that command's envelope. The
flag affects only a succeeded result; `partial` and `failed` continue to
return `stage`, `failed_reason`, `files`, and `step_durations`.

For a still-running task, `LKM_PARSE_RESULT_NOT_READY` means wait and retry the
same task rather than resubmit. `partial` and `failed` results can carry
expiring intermediate-file URLs. Consume them only for authorized recovery and
avoid reproducing their credentials in logs or messages.

Open Platform charges only the first `code=0` / `status=succeeded` Result per
task ID. `cache_hit=false` selects 1.00 CNY and `cache_hit=true` selects 0.10
CNY. A concurrent first retrieval can receive retryable upstream code `7002`
while another caller completes the charge; retry the same Result after a delay
only if `error.retryable` is true. A photon-insufficient refusal has
`BILLING_INSUFFICIENT` and `retryable: false`, even if its subtype is `7002`.
Keep the same task ID when the user later authorizes a different payment source.
No captured result is released on that failure.

## CLI capability boundary

The current commands do not expose service feedback submission or arbitrary
response controls. Use the flags shown by `bohr lkm <command> --help`; a
service field is not a CLI flag merely because a separate API-oriented skill
documents it.

## Version contract

This Skill is embedded in the `bohr` binary and changes with the CLI version:

```bash
bohr guide read bohr-lkm
bohr guide list bohr-lkm/references
bohr guide read bohr-lkm/references/commands.md
```

`bohr guide` does not fetch the standalone Bohrium Skills repository. Upgrade
`bohr` to get guidance for a newer LKM command surface.

## Payment-source option

The optional global `--payment-source` flag accepts `photon` or `balance`.
Omitting it sends no preference and preserves photon-first/balance fallback.
Follow [Choose the payment source](../SKILL.md#choose-the-payment-source) for
funding authorization and insufficient-funds recovery. After balance is authorized:

```bash
bohr lkm extract result <task_id> --payment-source balance --yes -o json
```

The same option is available on `bohr api`. Free commands remain free; selecting
a source does not create a charge on an otherwise free operation.
