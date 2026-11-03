---
name: bohrium-database
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Query Bohrium scientific databases with `bohr database`: browse polymer
  records, search crystal structures, or inspect and query private tabular
  databases by db_ak and table_ak. Use for Bohrium database, 私有数据库, polymer
  library, crystal structure search, table schema, or filtered row queries.
---

# Bohrium Database

## Choose the task

| Goal | Route |
|---|---|
| Obtain crystal structures | Search → inspect usable structures/warnings → optional archive download → validate local inputs |
| Browse polymers | Bounded library list → inspect returned scientific records |
| Query an existing private table | Supplied `db_ak` → tables → schema → typed filter/query, only where those routes are enabled |

## Workflow: retrieve structures for a calculation

1. Define composition/query, maximum matches, desired CIF/POSCAR format and the local destination. Archive download writes locally even though the remote operation is read-only.

```bash
bohr database structure search --query "Li Fe P O" --n-results 5   --output-format cif --download --out ./structures.tgz -o json
```

2. Inspect `data.cleaned_structures` for the usable scientific records. The command uses that same search's `data.output_dir` to download the archive; it is a retrieval URL, not a private `db_ak` or local path.
3. Check `notice.warnings` and `data.errors`. `STRUCTURE_SEARCH_PARTIAL` means some sources are missing despite success. Inspect archive contents and structure format before selecting local files as a Job input directory; neither match count nor a downloaded filename proves valid calculation inputs.
4. Complete with selected structure provenance, missing-source disclosure and the actual verified local files. If download fails, retain the search records; use the returned recovery hint and destination rules in the reference. An unusable archive link cannot be fixed by repeatedly downloading it.
5. For private tables, first confirm the routes are available: the published CLI can return `db_route_unavailable` and has no private-database discovery command. When enabled, pass the selected `table_ak` from table records to schema/query, and use schema column names/types for filter values. A route failure is not evidence of an empty database and is not fixed by inventing keys.

Choose the database family before acting:

- Domain libraries use `bohr database polymer` or `bohr database structure`.
  They return scientific records and do not use private table keys.
- Private tabular access uses `tables`, `schema`, and `query` with stable
  `db_ak` and `table_ak` identifiers.

All current database commands are read-only and free. Prefer structured output:

```bash
bohr auth whoami -o json
bohr database polymer list --page 1 --size 20 -o json
bohr database structure search --query "Si O" --n-results 5 -o json
```

For a private database, discover tables before querying rows:

```bash
bohr database tables <db_ak> -o json
bohr database schema <db_ak> <table_ak> -o json
bohr database query <db_ak> <table_ak> \
  --filter '[{"field":"status","operator":"eq","value":"active"}]' \
  --limit 20 \
  --offset 0 \
  -o json
```

Treat display names as labels, not identifiers. Preserve `db_ak` and
`table_ak` exactly between discovery, schema, and query calls. The filter is a
JSON array; use field names and types from the returned schema rather than
guessing them.

Structure search can return CIF or POSCAR output metadata and optionally
download its result archive. Confirm the local destination before `--download`;
the remote operation remains read-only, but it writes a local file.

The current CLI does not list private databases or mutate rows, tables, or
columns. If the requested workflow needs database creation, insert/update/delete,
or schema alteration, report that it is outside this command surface.

Read [`references/commands.md`](references/commands.md) for every current
command, flag, identifier, and output/download boundary.
