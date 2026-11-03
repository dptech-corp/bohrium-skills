# Tools command reference

## Free taxonomy discovery

| Command | Inputs | Output handoff |
|---|---|---|
| `bohr wiki tools domains` | none | Domain node IDs |
| `bohr wiki tools subdomains --domain <id1,id2>` | required comma-separated domain IDs | Subdomain records |
| `bohr wiki tools list [--domain <id>]` | optional domain ID | Domains or subdomains, depending on input |

These commands are free and read-only. The hidden `+list` alias is compatibility
only; use canonical commands in new automation.

## Paid hybrid search

```text
bohr wiki tools search <query> [--lang en-US] [--k 50] [--keywords <json>]
```

- `--lang` defaults to `en-US`.
- `--k` defaults to `50` and controls candidate retrieval.
- `--keywords` is a JSON object such as `{"dynamics":1.0,"trajectory":0.8}`;
  omit it to let the service derive keywords from the query.
- Price: **0.01 CNY** per request.

Results include display names, unique keys, and repository URLs. The
unique key, not the display name or taxonomy ID, is consumed by detail.

## Paid detail

```text
bohr wiki tools info <unique_tool_key>
```

Price: **0.02 CNY** per request. Detail may include overview, profile,
subdomains, tutorial, usage command, related entries, and repository metadata.
Report only returned fields.

## Billing

```bash
bohr billing pricing --resource tools -o json
```

Search and detail share the Wiki/Tool personal monthly 1,000-call quota. In
non-interactive use, add `--yes` only after spending is authorized. Free
taxonomy commands do not require billing confirmation.

The actual charged amount may be absent when quota or channel information is
returned. Preserve the notice as-is instead of converting quota use to money.

```bash
bohr guide read bohr-tools
bohr guide read bohr-tools/references/commands.md
```

## Payment-source option

The optional global `--payment-source` flag accepts `photon` or `balance`.
Omitting it sends no preference and preserves photon-first/balance fallback.
Follow [Choose the payment source](../SKILL.md#choose-the-payment-source) for
funding authorization and insufficient-funds recovery. After balance is authorized:

```bash
bohr wiki tools search "molecular dynamics" --payment-source balance --yes -o json
```

The same option is available on `bohr api`. Free commands remain free; selecting
a source does not create a charge on an otherwise free operation.
