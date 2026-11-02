---
name: bohrium-scholar-search
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Search Bohrium scholars and retrieve a known scholar profile with the
  `bohr scholar` CLI. Use for 学者搜索, researcher lookup, affiliation or
  research-tag filtering, publication/citation summaries, and scholar IDs.
  Use `bohr paper search` when the primary object is a paper rather than a person.
---

# Bohrium Scholar Search

Start with search when no scholar ID is known:

```bash
bohr scholar search "Geoffrey Hinton" --page 1 --size 10 -o json
bohr scholar search "catalysis" --school "Tsinghua University" --tags "DFT,catalysis" -o json
```

Search is free and read-only. Use filters only when the user supplied them or
the unfiltered name is ambiguous. Capture the returned scholar ID rather than
matching a profile by display name alone.

Retrieve the selected profile with:

```bash
bohr scholar info <scholar_id> -o json
```

Verify identity using name, affiliation, research directions, and other
returned evidence before attributing publications or metrics. A scholar ID is
not a paper ID, author-name string, or knowledge-base document ID.

Use `bohr paper search` for paper records and `bohr paper patent` for patents.
The current scholar CLI does not mutate profiles or expose separate publication,
education, or citation-history subcommands; those fields may appear inside the
profile response when available.

Read [`references/commands.md`](references/commands.md) for every current flag,
the search-to-profile handoff, and presentation boundaries.
