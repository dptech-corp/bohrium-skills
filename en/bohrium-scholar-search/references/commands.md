# Scholar command reference

## Search

```text
bohr scholar search <query> [flags]
```

| Flag | Default | Meaning |
|---|---:|---|
| `--page` | `1` | Page number |
| `--size` | `10` | Page size |
| `--school` | empty | School filter |
| `--tags` | empty | Comma-separated research tags |
| `--affiliation` | empty | Affiliation filter |

Use the smallest useful page and preserve filters when presenting candidates.
Disambiguate common names with affiliation and research direction before
selecting a scholar ID.

## Profile

```text
bohr scholar info <scholar_id>
```

The profile response may include identity, affiliations, research interests,
publication summaries, citations, h-index, education, and work experience.
Report only fields actually returned; missing fields are unknown, not zero.

Both commands are free, authenticated, read-only operations. Prefer `-o json`
for field-preserving output and keep request IDs when results appear inconsistent.

The current CLI exposes no profile-edit or separate publications command.

```bash
bohr guide read bohr-scholar-search
bohr guide read bohr-scholar-search/references/commands.md
```
