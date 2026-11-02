---
name: bohrium-lkm
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Search and trace scientific claims, reasoning chains, paper knowledge graphs,
  and extracted variables with the `bohr lkm` CLI, or asynchronously extract
  research questions and reasoning from a local PDF. Use for Bohrium LKM,
  Large Knowledge Model, 大知识模型, claim provenance, paper graphs, or LKM PDF
  extraction. Use `bohr paper`, `bohr kb`, or `bohr pdf` for adjacent paper
  search, personal knowledge-base, or document-layout tasks.
---

# Bohrium LKM

LKM (Large Knowledge Model, 大知识模型) turns papers into source-grounded
reasoning graphs: questions, claims, reasoning chains, and evidence become
addressable units, aligned across papers.

Operate LKM through the version-matched `bohr lkm` commands. Choose the route
before acting:

- Search the ingested corpus with `search` for claims, questions, abstracts,
  and scientific concepts.
- Search whole argument or experimental processes with `reasoning`.
- Retrieve a known paper's `nodes` / `edges` graph with `graph`.
- List a known paper's references or cited-by set with `reference`.
- Trace a known conclusion claim with `claim reasoning`.
- Hydrate known global node IDs with `variables batch`.
- Upload a local PDF that may not be in the corpus with asynchronous `extract`.

This is not generic paper discovery (`bohr paper search`), personal knowledge
base management (`bohr kb`), or Uni-Parser layout/text/table/formula extraction
(`bohr pdf parse`).

## Authentication, output, and billing

Use an existing Bohrium login or AccessKey and prefer structured output:

```bash
bohr auth whoami -o json
bohr billing pricing --resource lkm -o json
```

`search`, `reasoning`, `graph`, `reference`, `variables batch`, and
`claim reasoning` are
fixed-price calls at **0.05 CNY** each, covered first by LKM's personal
monthly 1,000-call quota. The CLI confirmation gate applies even when remaining
quota may make the call free. In non-interactive use, add `--yes` only after
spending is authorized; otherwise report `CONFIRMATION_REQUIRED`.

Extract is separate from the monthly query quota: **1.00 CNY** per successful
paper, or **0.10 CNY** on a cache hit. The same task is not charged twice.
In non-interactive use, add `--yes` only after spending is authorized.

Submitting still uploads the selected PDF to Bohrium, so confirm the exact file
and the user's intent before the real submission.

## Search and trace the ingested corpus

Start from natural language when no identifier is known:

```bash
bohr lkm search "perovskite thermal stability" \
  --mode hybrid \
  --sort comprehensive \
  --scopes conclusion,abstract \
  --top-k 10 \
  --yes \
  -o json

bohr lkm reasoning \
  --query "infer phase stability from XRD evidence" \
  --top-k 10 \
  --yes \
  -o json
```

`search` returns representative `variables`, same-paper `related` context, and
authoritative paper metadata. `reasoning` returns whole reasoning-chain hits.
Treat `score` and `rerank_score` as ranking signals, never as confidence,
truth probability, or evidence strength.

Capture identifiers from structured results rather than inventing IDs:

- A numeric paper ID is consumed by `bohr lkm graph --paper-id` or
  `bohr lkm reference --paper-id`.
- A DOI is consumed by `bohr lkm graph --doi` or `bohr lkm reference --doi`.
- A `paper:<digits>` package key is consumed by `bohr lkm graph --package-id`.
- A title is consumed by `bohr lkm graph --title` (may return multiple graphs).
- A global `gcn_...` ID is consumed by `variables batch`; only a conclusion
  claim with reasoning should be sent to `claim reasoning`.
- A graph-local node ID describes a node only inside one returned graph and is
  not a global variable, claim, or paper ID.

```bash
bohr lkm graph --paper-id <numeric_paper_id> --yes -o json
bohr lkm graph --doi <doi> --yes -o json
bohr lkm graph --title "<paper_title>" --yes -o json
bohr lkm reference --paper-id <numeric_paper_id> --yes -o json
bohr lkm variables batch <global_gcn_id> [global_gcn_id...] --yes -o json
bohr lkm claim reasoning <global_conclusion_gcn_id> --format graph --yes -o json
```

Use the graph's direction and node kinds when explaining provenance. A
`highlight` is positive support, a `weak_point` is a limitation, and a
`subproblem` is a research question or motivation; they are not interchangeable
with the final `conclusion`.

## Extract knowledge from a local PDF

Use `extract` when the user has a local PDF rather than a known corpus result:

```bash
bohr lkm extract submit paper.pdf -o json
bohr lkm extract wait <task_id> --interval 5s --timeout 30m -o json
bohr lkm extract result <task_id> --yes -o json
bohr lkm extract result <task_id> --format graph --yes -o json
```

Extract costs **1.00 CNY** per successful paper, or **0.10 CNY** on a cache hit.
Confirm spend before a real submit.

The local file must be a non-empty regular PDF of at most 64 MiB; the service
also enforces a 50-page limit. Submission returns an extract task ID and may report
`cache_hit`; a cache hit reuses an existing extraction. Do not resubmit a
running PDF to make it faster. The service returns `290020` with the existing
task ID, which the caller should continue through `status` or `wait`.

`queued` and `running` are non-terminal. `succeeded`, `partial`, and `failed`
are terminal. `partial` is a non-retryable business terminal state for a paper
that cannot produce a complete graph; the same PDF should not be resubmitted.
`failed` is a technical terminal state and the PDF may be resubmitted. A local
wait timeout or cancellation only stops waiting; it does not cancel the remote
task. Keep the extract task ID and resume with `status`, `wait`, or `result`.

The default `local` success format returns a flat `variables` / `factors` / `motivations` graph.
`--format graph` returns `addressed_problems`, `open_questions`, and
`graph.nodes` / `graph.edges` — the same field kinds as `bohr lkm graph`, not
that command's envelope. In either format, local node identifiers exist only
in that extraction. Only a non-null `global_id` can be reused with
`variables batch` or, when it identifies a supported conclusion claim,
`claim reasoning`.

For `partial` or `failed`, inspect the stage, failure reason, and any
intermediate files. Their URLs are expiring credentials: use them only for the
requested recovery and keep them out of chat transcripts and logs.

`290017` means the task is not ready: continue waiting. For a concurrent first
Result charge, retry with the same task ID and payment source only when the
returned structured error has `retryable: true`. Photon insufficiency may retain
upstream subtype `7002` but is `BILLING_INSUFFICIENT` with `retryable: false`: stop
and follow the payment-source rule below. Keep the existing task; do not resubmit
the PDF to recover from a billing failure.

## Capability boundary

The current CLI intentionally exposes a focused subset of the broader LKM
service. Feedback submission and arbitrary response controls remain outside
this command surface. Inspect `bohr lkm <command> --help` instead of
translating every service field into a flag.

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

Read [`references/commands.md`](references/commands.md) for every current
command, flag, output field, ID rule, graph semantic, extract state, and version
contract.
