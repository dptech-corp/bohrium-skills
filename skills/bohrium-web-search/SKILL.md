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

## Choose the task

| Goal | Route |
|---|---|
| Find a website, official manual, or recent public information | Narrow web query → select primary sources → verify their pages |
| Find papers, people, or private documents | Route to Paper, Scholar, or Knowledge Base before searching |
| Answer a factual question | Use returned links as leads; verify the evidence before writing the answer |

## Workflow: find documentation and deliver a sourced answer

1. Identify the product/topic, version if relevant, and what fact the user needs. Include those terms in the query; `--limit` bounds returned leads, not verification depth.
2. Run a small discovery query:

```bash
bohr search "official VASP documentation PAW method" --limit 5 -o json
```

3. Inspect the actual returned container/URL fields (for example `data.organic_results[].link` or `data.results[].url`); carry each selected title, URL and snippet together. Open the selected official page using an available page reader/browser. This Skill's command searches; it does not fetch full pages.
4. Verify that the page supports the requested claim and matches the version/date. Deliver the supported answer with its source URL. If only snippets are available, label the finding as a search lead and state the missing verification.
5. An empty result calls for one more precise query when useful; an authentication/network error calls for fixing that error. Neither supports a claim that the subject does not exist. Stop when the requested fact is supported, rather than accumulating unrelated hits.

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
queries, and other extras. JSON is forwarded without normalizing result containers:
inspect the actual response first (`organic_results[].link` and `results[].url`
are supported upstream shapes). The shown `-q '.data.organic_results'` projection
applies only when that field exists; an empty projection is not proof of no hits.

Read [`references/commands.md`](references/commands.md) for the exact command,
limit validation, output guidance, and routing boundary.
