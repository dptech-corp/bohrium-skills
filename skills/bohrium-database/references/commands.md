# Database command reference

## Domain libraries

```bash
bohr database polymer list --page 1 --size 20 -o json
bohr database structure search --query "Li Fe P O" --n-results 5 --output-format cif -o json
```

| Command | Current inputs |
|---|---|
| `bohr database polymer list` | `--page` default `1`; `--size` default `20`, range `1-100` |
| `bohr database structure search` | `--query`, `--n-results` default `5`, `--output-format cif|poscar`, optional `--project-id` |

To save the structure archive locally, add `--download --out <archive-path>`.
If `--out` is omitted, use the CLI-reported destination. Do not infer that a
downloaded archive contains one structure per search hit without inspecting it.

If the archive download itself fails, the search results stay in `data` and
the error says why: `UPSTREAM_TIMEOUT` (exit 11), `NETWORK_ERROR`,
`UPSTREAM_ERROR` carrying the storage HTTP status, or
`STRUCTURE_ARCHIVE_DOWNLOAD_CANCELLED` when the download is interrupted (Ctrl-C
or SIGTERM). Its `hint.command` reruns the same search, because the archive link
is issued by the search. A failed or interrupted download leaves no partial file.
If the search returns no usable archive link (`data.output_dir` empty or not an
http(s) URL), `--download` fails with non-retryable `UPSTREAM_INVALID_RESPONSE`
and no rerun command; the structures are still in `data.cleaned_structures`.

When some matched structures cannot be retrieved from their source, the search
still succeeds with the ones that could: `notice.warnings[]` carries
`STRUCTURE_SEARCH_PARTIAL` with how many are missing, and `data.errors` lists
them. If none could be retrieved it fails with `UPSTREAM_ERROR` (subtype
`structure_source_unavailable`, retryable) and `data.errors` lists what failed;
its `hint.command` reruns the same search with a larger `--n-results` so other
sources can fill in.

## Private tabular databases

```text
bohr database tables <db_ak>
bohr database schema <db_ak> <table_ak>
bohr database query <db_ak> <table_ak> [flags]
```

`tables` discovers table names and `table_ak` values. `schema` returns the
field names and types accepted by filters. `query` supports:

| Flag | Default | Meaning |
|---|---:|---|
| `--limit` | `20` | Maximum rows returned |
| `--offset` | `0` | Rows skipped before this page |
| `--filter` | empty | Filter conditions as a JSON array |

Example filter shape:

```json
[
  {"field": "temperature", "operator": "gte", "value": 300},
  {"field": "status", "operator": "eq", "value": "active"}
]
```

Use operators and value types supported by the live schema. An empty result is
not evidence that the table or field does not exist; preserve pagination and
filter details when reporting it.

## Capability boundary

The current CLI exposes no private-database catalog command and no row/table
mutation commands. It also does not export an entire table in one operation;
page with `limit`/`offset` and stop at the requested bound.

These files are versioned with the binary:

```bash
bohr guide read bohr-database
bohr guide read bohr-database/references/commands.md
```
