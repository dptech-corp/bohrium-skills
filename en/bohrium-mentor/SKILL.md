---
name: bohrium-mentor
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Ask Bohrium's AI science mentor with `bohr agents mentor` for scientific
  explanations, research-progress summaries, and method comparisons. Use for
  Bohrium Mentor, Science Navigator, AI 科研导师, 科学小导师, or literature-grounded
  scientific Q&A. Use `bohr paper`, `bohr scholar`, or `bohr lkm` for
  record-oriented retrieval and claim provenance.
---

# Bohrium Mentor

Use the version-matched `bohr agents mentor` command to ask one focused
scientific question. Mentor is a synthesis workflow: it retrieves scientific
literature internally and streams a Markdown explanation.

Choose a more structured command when the deliverable is records rather than
an explanation:

- Use `bohr paper search` for specific papers and bibliographic fields.
- Use `bohr scholar search` for scholar profiles.
- Use `bohr lkm` for claims, reasoning chains, paper graphs, and provenance.
- Use a dedicated literature-review workflow when the user needs a reproducible
  search protocol, screened corpus, or systematic evidence table.

## Authenticate and authorize the session

Check the current identity and version-matched price before creating a session:

```bash
bohr auth whoami -o json
bohr billing pricing --resource mentor -o json
```

The bundled catalog price is **2.00 CNY per session**. The gateway precharges
before session creation; answer-stream completion is not the charge trigger.
A request to answer a question is not itself permission to spend. In non-interactive use,
add `--yes` only after spending is authorized; otherwise return the
`CONFIRMATION_REQUIRED` result to the user.

Use `--dry-run` to inspect the normalized question and options without creating
a session:

```bash
bohr agents mentor \
  "Compare density functional theory and Hartree-Fock for periodic solids" \
  --discipline Physics \
  --journal-type foreign \
  --model reason \
  --dry-run \
  -o json
```

## Ask the question

After authorization, run one complete question:

```bash
bohr agents mentor \
  "Explain recent strategies for suppressing lithium dendrites in solid-state batteries" \
  --discipline Chemistry \
  --journal-type foreign \
  --model reason \
  --yes
```

`All`, `Physics`, `Chemistry`, `Biology`, `Materials`, and
`Computer Science` are common discipline values. The CLI forwards filter values
to Mentor; keep the default `All` when the requested field is uncertain.
`foreign` and `chinese` select the journal scope. `reason` is the default model.

For automation, prefer JSON:

```bash
bohr agents mentor "Explain the evidence for high-temperature superconductivity" \
  --yes \
  -o json
```

Human output streams only the main Markdown answer. Structured success exposes
`data.answer` and `data.session_id`. Treat inline citation markers as part of
the answer, not as complete bibliographic records; use `bohr paper search` when
the user needs independently inspectable references.

## Handle delay and interruption

Deep reasoning may have a noticeable first-token delay, and the command timeout
is five minutes. Keep the process attached while useful output is still
arriving.

Session creation and answer streaming are separate phases. If streaming fails,
the session may already be billable and human output may already contain a
partial answer. Preserve the actual stdout, structured error, billing notice,
and request ID that were returned. A successful machine result is the only
place the CLI returns `data.session_id`.

**Do not automatically retry** after session creation or a stream failure. A
retry creates a new session and may charge again. Report the interruption and
ask whether the user authorizes another session.

## Capability boundary

The CLI intentionally renders only the main answer. It does not expose Mentor's
non-main UI layouts, citation-card metadata, recommended follow-ups, or a resume
command. Stay inside `bohr agents mentor`; do not reconstruct the service's
transport protocol or credential handling in shell or Python.

The old top-level `bohr mentor` path has been removed. Use
`bohr agents mentor`.

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

Read [`references/commands.md`](references/commands.md) for exact flags, output,
billing, dry-run behavior, failure handling, and the embedded-version contract.
