---
name: bohrium-paper-search
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Search academic papers and patents with `bohr paper search` and
  `bohr paper patent`. Use for 文献检索, paper discovery, bibliography building,
  patent search, date/JCR/database filtering, or technical surveys. Use Scholar
  Search for people and LKM for claim/reasoning provenance.
---

# Bohrium Paper and Patent Search

Choose the record type first:

- `bohr paper search <query>` returns academic paper records.
- `bohr paper patent <query>` returns patent records.

Both are fixed-price per-call operations. Check the exact tier before spending:

```bash
bohr billing pricing --resource paper -o json
bohr billing pricing --resource patent -o json
```

Paper tiers are **0.05 CNY** for normal (`--type 0`) and **0.10 CNY** for
enhanced (`--type 1`). Patent tiers are **0.10 CNY** normal, **0.30 CNY**
enhanced, and **0.50 CNY** professional (`--type 0/1/2`). In non-interactive
use, add `--yes` only after spending is authorized.

Start with a narrow professional query and the least expensive suitable tier:

```bash
bohr paper search "solid electrolyte interphase lithium metal" \
  --year-from 2020 --year-to 2026 --jcr "Q1,Q2" --type 0 --yes -o json

bohr paper patent "solid-state battery dendrite suppression" \
  --type 0 --yes -o json
```

Pagination is a new paid call each time. Bound `--page` and `--size` to the
requested evidence volume; do not page automatically without additional
spending authorization.

Treat titles, authors, DOI/publication identifiers, dates, journals, abstracts,
and patent fields as returned records. The upstream service may return mixed
publication types from paper search, so inspect each record's type and metadata
instead of assuming every hit is a journal paper. Retrieval rank is not evidence quality.
For scholar profiles use `bohr scholar search`; for claim chains or paper
graphs use `bohr lkm`; for synthesis use Mentor or a literature-review workflow.

Billing notices may contain only channel or quota information; the actual charged amount may be absent.
Report only the amount explicitly returned.
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


Read [`references/commands.md`](references/commands.md) for every filter, tier,
price, result boundary, and safe pagination rule.
