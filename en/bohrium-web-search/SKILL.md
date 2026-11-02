---
name: bohrium-web-search
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Search the open web with `bohr search` for documentation, tutorials, news,
  websites, and general research information. Use for Bohrium web search or
  网页搜索. Use `bohr paper search`, `bohr scholar search`, or `bohr kb search`
  when the requested records belong to those structured Bohrium collections.
---

# Bohrium Web Search

Use the free, read-only web-search command for general public web information:

```bash
bohr search "official VASP documentation PAW method" --limit 5 -o json
bohr search "official VASP documentation PAW method" --limit 5 -o json -q '.data.organic_results'
```

The query is one positional string. `--limit` accepts `1-10` and defaults to
`5`. Prefer a narrow query containing the subject and desired source type.

Route structured academic or private content elsewhere:

- `bohr paper search` for paper records and bibliographic filters.
- `bohr scholar search` for researchers and profiles.
- `bohr kb search <kb_id> <query>` for content already stored in one of the
  user's Bohrium knowledge bases.

Search results are discovery leads, not automatically authoritative facts.
Preserve titles, URLs, snippets, and source names from the response; open or
verify primary sources when accuracy matters. The current CLI exposes no
recency, site/domain, locale, or provider-specific filter.

Raw provider data may include base64 favicons, internal search URLs, related
queries, and other provider extras. When only organic results matter, use the
supported jq-like path `-q '.data.organic_results'` rather than object construction.

Read [`references/commands.md`](references/commands.md) for the exact command,
limit validation, output guidance, and routing boundary.
