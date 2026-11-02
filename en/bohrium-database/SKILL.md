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
