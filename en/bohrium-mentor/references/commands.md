# Mentor command reference

This reference describes the Mentor surface embedded in the same `bohr`
version as this skill.

## Canonical command

```text
bohr agents mentor <question> [flags]
```

All positional words are joined into one question. Quote the question so shell
punctuation and whitespace remain intact.

| Flag | Default | Meaning |
|---|---|---|
| `--discipline <value>` | `All` | Discipline filter forwarded to Mentor |
| `--journal-type <value>` | `foreign` | Journal scope; use `foreign` or `chinese` |
| `--model <value>` | `reason` | Mentor model |
| `-o, --output <format>` | configured default | Output format; use `json` for automation |
| `-q, --jq <expression>` | empty | Filter structured output |
| `--dry-run` | false | Print the planned session request without sending it |
| `--yes` | false | Confirm this invocation after spending is authorized |
| `--no-interactive` | false | Disable prompts; an unconfirmed paid call returns an error |

At least one question word is required. Check the installed command if this
version's help differs:

```bash
bohr agents mentor --help
```

The former top-level `bohr mentor` path has been removed; use
`bohr agents mentor`.

## Authentication

Use the CLI's configured identity rather than passing credentials in the
question or command line:

```bash
bohr auth whoami -o json
```

If authentication is missing, follow the recovery guidance returned by the
current `bohr auth` command. Keep credentials out of prompts, logs, and skill
files.

## Billing and confirmation

Inspect the catalog before spending:

```bash
bohr billing pricing --resource mentor -o json
```

The embedded catalog defines `mentor.session` as **2.00 CNY per session**. The
gateway precharges before session creation. A successful session creation does
not mean the answer stream has completed. The existing photon refund rule applies
to a definite session-creation business failure; balance precharges are not refunded
by the gateway. An answer-stream failure does not reverse the session charge.
The command uses the fixed per-call confirmation gate.

For an automated caller:

1. Present the price and intended question to the user.
2. Wait until spending is authorized.
3. Run the same command with `--yes`.

Without confirmation, a non-interactive invocation exits before execution with
`CONFIRMATION_REQUIRED`, so no session is created. A standing local setting may
disable the prompt, but it does not grant the agent spending authority.

`--dry-run` does not create a session and does not require `--yes`:

```bash
bohr agents mentor "Explain VASP's projector augmented-wave method" \
  --discipline Materials \
  --dry-run \
  -o json
```

## Output contract

### Human output

Human output writes the main Markdown text as stream fragments arrive. On a
clean end, it prints a trailing newline and any billing result notice emitted
by the gateway.

The command ignores non-main stream content. If the answer contains inline
reference markers, the CLI does not pair them with structured citation-card
records.

### Machine output

Prefer `-o json` for reliable consumption. On success, the standard envelope's
data is:

```json
{
  "answer": "<complete accumulated main-layout Markdown>",
  "session_id": "<Mentor session ID>"
}
```

The full paths are `data.answer` and `data.session_id`. The envelope may also
contain `notice.billing` and request metadata. It does not contain the original
SNP event stream or additional UI-layout data.

## Filters and question design

- Keep the default discipline `All` for cross-disciplinary or uncertain topics.
- Use a specific discipline only when it materially narrows the literature
  context.
- Use `--journal-type chinese` only when Chinese-journal coverage is intended;
  otherwise keep `foreign`.
- Ask for one coherent deliverable. Include scope, comparison axis, time window,
  and desired depth in the question when they matter.
- Treat the answer as synthesis. For exact paper metadata, run
  `bohr paper search`; for scholars, run `bohr scholar search`; for claim and
  reasoning provenance, choose the relevant `bohr lkm` workflow.

## Failure and recovery

The command allows up to five minutes for session creation and streaming.

| Failure phase | Observable result | Safe action |
|---|---|---|
| Before confirmation | `CONFIRMATION_REQUIRED` | Obtain spending authorization, then add `--yes` |
| Session request | Standard authentication, network, or upstream error | Correct the reported cause; no session ID was returned |
| Invalid session response | `UPSTREAM_INVALID_RESPONSE` | Preserve the request ID and report the upstream contract failure |
| Answer stream | `UPSTREAM_UNAVAILABLE`; human stdout may be partial | Preserve output and billing notice; ask before another session |
| Local timeout or cancellation | Stream ends with an error or cancellation | Treat the answer as incomplete and ask before another session |

**Do not automatically retry** a stream failure, timeout, or cancellation. A
new invocation creates a new session, and successful session creation is the
billing trigger even if its answer later fails to finish.

## Version contract

The files are embedded at build time and versioned with the `bohr` binary:

```bash
bohr guide read bohr-mentor
bohr guide list bohr-mentor/references
bohr guide read bohr-mentor/references/commands.md
```

`bohr guide` does not fetch the standalone Bohrium Skills repository. Upgrade
`bohr` to receive newer command guidance.

## Payment-source option

The optional global `--payment-source` flag accepts `photon` or `balance`.
Omitting it sends no preference and preserves photon-first/balance fallback.
Follow [Choose the payment source](../SKILL.md#choose-the-payment-source) for
funding authorization and insufficient-funds recovery. After balance is authorized:

```bash
bohr agents mentor "Explain graphene structure" --payment-source balance --yes -o json
```

The same option is available on `bohr api`. Free commands remain free; selecting
a source does not create a charge on an otherwise free operation.
