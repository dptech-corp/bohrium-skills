---
name: bohrium-tools
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Browse and search the Bohrium scientific Tools library with `bohr wiki tools`.
  Use for scientific software discovery, 工具搜索, domain/subdomain browsing,
  hybrid keyword search, tool unique keys, tutorials, usage commands, or
  repository links. Use paper, Wiki, or LKM commands for literature knowledge.
---

# Bohrium Scientific Tools

## Choose the task

| Goal | Route |
|---|---|
| Understand available scientific domains | Free taxonomy discovery; returned node IDs are categories |
| Find software for a concrete task | Paid hybrid search → selected tool key → paid detail |
| Recommend a usable tool | Inspect detail/source → check license, platform and installation requirements |

## Workflow: select a tool for an analysis task

1. Define input format, scientific operation, runtime/platform constraints and the authorized search/detail budget. Taxonomy browsing can refine the wording but does not return executable tool records.
2. Search for the task, then inspect one promising candidate:

```bash
bohr wiki tools search "molecular dynamics trajectory analysis" --k 10 --lang en-US --yes -o json
bohr wiki tools info <tool_unique_key> --yes -o json
```

3. Replace `<tool_unique_key>` with the selected search record's `data.tools[].tool_unique_key`. Category discovery uses `data.items[].node_id`; that node ID is not accepted as a tool key. Search has no `--domain` filter: encode the relevant field in the query rather than inventing a flag.
4. Inspect returned capabilities, repository, tutorial and usage details. Verify that the candidate accepts the user's inputs and is available for the target platform/license. A returned command is reference material; execute or install only when that action is in scope.
5. Complete with a justified selection, source link, constraints and next runnable step. A ranked hit alone is discovery, not a compatibility check. If detail is missing or search is empty, retain the limitation; another paid candidate/detail requires room in the existing budget.

Choose free browsing before a paid query when the domain is known:

```bash
bohr wiki tools domains -o json
bohr wiki tools subdomains --domain <domain_id> -o json
bohr wiki tools list --domain <domain_id> -o json
```

`domains`, `subdomains`, and `list` are free taxonomy discovery. `list --domain`
returns subdomain records, not tools. Their node IDs describe categories. Obtain
a unique tool key from hybrid search before requesting tool detail.

For intent-driven discovery:

```bash
bohr billing pricing --resource tools -o json
bohr wiki tools search "molecular dynamics trajectory analysis" --k 20 --lang en-US --yes -o json
bohr wiki tools info <unique_tool_key> --yes -o json
```

Hybrid search costs **0.01 CNY** per request and detail costs **0.02 CNY**,
covered first by the shared monthly 1,000-call Wiki/Tool quota. The confirmation
gate still applies when quota may make the request free. Add `--yes` only after
spending is authorized.

Runtime notices may show quota remaining or a billing channel while the actual
charged amount may be absent. Do not infer a monetary charge from quota usage.

Search may auto-generate weighted keywords from the query. Supply `--keywords`
only as a JSON object when the user needs explicit weights. Treat ranking as
retrieval order, not a quality or compatibility guarantee. Verify repository,
license, installation, and platform requirements from returned sources before
recommending execution.

The current CLI exposes no separate summary, tags, subdomain-detail, or tools-
under-subdomain endpoint beyond the canonical commands above.
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


Read [`references/commands.md`](references/commands.md) for every command,
identifier handoff, current flag, price, and capability boundary.
