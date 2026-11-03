# Knowledge-base command reference

## Knowledge bases

| Command | Inputs | Effect |
|---|---|---|
| `bohr kb create <name>` | optional `--cover <url>` | Create KB |
| `bohr kb list` | `--page` default `1`, `--size` default `20` | Read KB catalog |
| `bohr kb search <kb_id> <query>` | KB ID and query | Search KB content |
| `bohr kb upload <local-file>` | exactly one of `--kb-id` or `--parent-id`; optional `--name` | Upload/register document |
| `bohr kb delete <kb_id>...` | one or more KB IDs | Destructive KB deletion |

## Documents

| Command | Inputs | Effect |
|---|---|---|
| `bohr kb document list` | `--kb-id` or `--parent-id`, `--page` default `1`, `--size` default `50` (`1-100`) | List stable document IDs |
| `bohr kb document get <document_id>` | document ID | Read safe metadata |
| `bohr kb document delete <document_id>` | document ID | Delete document and attachments |

## Identifier rules

- A knowledge-base ID identifies the KB and is consumed by create/list/search/delete routes.
- A root node ID is the KB's document-tree root.
- A parent node ID identifies the root or a folder destination for list/upload.
- A document ID identifies one uploaded document and is consumed by document get/delete.

Use the CLI's `--kb-id` resolver when only the KB ID is known. Never pass a
document ID as a parent node or infer IDs from titles.

All current commands are free. Create/upload are writes; KB and document delete
are destructive and require `--yes` in non-interactive use. Upload may be
asynchronous upstream; report only parse/readiness fields actually returned.

The current CLI has no commands for KB update/detail/discovery/favorites,
folder mutation, literature metadata, tags, notes, recall endpoints, or
permission management.

```bash
bohr guide read bohr-knowledge-base
bohr guide read bohr-knowledge-base/references/commands.md
```
