# Web search command reference

```text
bohr search <query> [--limit <1-10>]
```

- `query` is required and all positional words are joined into one query.
- `--limit` defaults to `5`, accepts `1-10`, and bounds returned results.
- The command is authenticated, free, and read-only.
- Prefer `-o json` when downstream work needs stable URLs and snippets.
- Use `-q '.data.organic_results'` to select organic results and omit provider extras.

Examples:

```bash
bohr search "LAMMPS official fix npt documentation" --limit 5 -o json
bohr search "recent solid-state battery industry news" --limit 10 -o json
```

The CLI has no advanced web-search filters. Express source preference, date,
language, or scope in the query and verify them from the returned source rather
than assuming the proxy enforced them.

Use the structured commands when their object model matters:

```bash
bohr paper search "solid electrolyte interface" -o json
bohr scholar search "John Goodenough" -o json
bohr kb search <kb_id> "experimental protocol" -o json
```

```bash
bohr guide read bohr-web-search
bohr guide read bohr-web-search/references/commands.md
```
