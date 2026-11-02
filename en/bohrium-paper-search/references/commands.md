# Paper and patent command reference

## Academic papers

```text
bohr paper search <query> [flags]
```

| Flag | Default | Meaning |
|---|---:|---|
| `--page` | `1` | Page number; each invocation may be billed |
| `--size` | `10` | Page size |
| `--year-from` | unset | Inclusive start year |
| `--year-to` | unset | Inclusive end year |
| `--type` | `0` | `0` normal, `1` enhanced |
| `--jcr` | empty | Comma-separated JCR zones |
| `--db` | empty | Comma-separated source databases |

Prices: normal **0.05 CNY**, enhanced **0.10 CNY** per request.

## Patents

```text
bohr paper patent <query> [flags]
```

| Flag | Default | Meaning |
|---|---:|---|
| `--page` | `1` | Page number; each invocation may be billed |
| `--size` | `10` | Page size |
| `--type` | `0` | `0` normal, `1` enhanced, `2` professional |

Prices: normal **0.10 CNY**, enhanced **0.30 CNY**, professional **0.50 CNY**
per request.

## Confirmation and result handling

```bash
bohr billing pricing --resource paper -o json
bohr billing pricing --resource patent -o json
```

The CLI confirmation gate applies to every page and tier. A request to prepare
a search is not spending authorization. After authorization, add `--yes` to
that exact invocation and preserve its request ID.

Use structured output for record fields. Do not infer missing authors, DOI,
abstracts, dates, JCR zones, patent applicants, or legal status. Deduplicate by
stable identifiers when present, not normalized titles alone.

Paper search can contain mixed publication types. Validate the returned type,
venue, DOI/patent fields, and source before classifying an item. Runtime billing
may omit the actual charged amount even when a channel is present.

The current CLI has no author, language, custom sort, citation-export, or bulk
review flag. Express unsupported intent in the query or post-process returned
records without inventing request options.

```bash
bohr guide read bohr-paper-search
bohr guide read bohr-paper-search/references/commands.md
```

## Payment-source option

The optional global `--payment-source` flag accepts `photon` or `balance`.
Omitting it sends no preference and preserves photon-first/balance fallback.
Follow [Choose the payment source](../SKILL.md#choose-the-payment-source) for
funding authorization and insufficient-funds recovery. After balance is authorized:

```bash
bohr paper search "graphene" --payment-source balance --yes -o json
```

The same option is available on `bohr api`. Free commands remain free; selecting
a source does not create a charge on an otherwise free operation.
