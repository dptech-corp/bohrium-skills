---
name: bohrium-knowledge-base
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Create, list, search, upload to, or delete Bohrium knowledge bases and manage
  their documents with `bohr kb`. Use for 知识库, KB content search, document
  IDs, folder/root node IDs, local document upload, or safe document deletion.
  Use paper search for the public literature catalog and LKM for claim graphs.
---

# Bohrium Knowledge Base

## Choose the task

| Goal | Route |
|---|---|
| Search existing private content | Resolve KB ID → targeted `search` → inspect returned evidence |
| Add a local document | Confirm file/KB → upload → locate stable document ID → inspect indexing → test retrieval |
| Remove content | List/get exact document or KB → authorized delete; folder IDs are separate |

## Workflow: upload a paper and verify it can be retrieved

1. Resolve the intended KB from `kb list`, confirm the local file and destination, then upload. Use either `--kb-id` or `--parent-id`, following the selector rules below.

```bash
bohr kb upload ./paper.pdf --kb-id <kb_id> -o json
bohr kb document list --kb-id <kb_id> --page 1 --size 50 -o json
bohr kb document get <document_id> -o json
bohr kb search <kb_id> "distinctive phrase from the uploaded paper" -o json
```

2. Preserve upload's `data.knowledge_base_id`, `parent_id` and `md5`. Its `data.document.index_state=submitted` means registration succeeded, not indexing completed. `data.document.document_id` may be absent if post-upload lookup has not found the document yet; use document listing to resolve it instead of uploading again.
3. In `data.documents[]`, match the selected document and preserve its `document_id` for get/delete. Inspect that row's `index_state`: `processing` permits bounded rechecks; `ready` permits the retrieval check; `failed`/`unsupported` needs a reported failure/format remedy. `unavailable`/`unknown` is missing evidence, not a reason for endless polling. Page-level `index_status_available=false` can be caused by another row.
4. `document get` verifies safe metadata, not full text. Search for distinctive content and confirm an actual hit attributable to the intended document before claiming it is searchable. Preserve an explicit `recallable=false` limitation; a no-hit query alone does not prove indexing failed.
5. Complete with KB ID, stable document ID, observed indexing state and the retrieval evidence. If registration succeeded but indexing/search is unverified, report that exact partial outcome. Continue the original document's checks; a repeated upload is a new write, not a readiness probe.

Start by resolving the knowledge-base ID:

```bash
bohr kb list --page 1 --size 20 -o json
bohr kb create "Project literature" -o json
```

A knowledge-base ID identifies the KB itself. Upload and document listing use a
root or parent node ID internally; the CLI can resolve the root from `--kb-id`:

```bash
bohr kb upload ./paper.pdf --kb-id <kb_id> -o json
bohr kb document list --kb-id <kb_id> --page 1 --size 50 -o json
```

Use `--parent-id` when the user selected a specific root/folder node. A parent
node ID is not a knowledge-base ID or document ID.

Search content inside one KB with:

```bash
bohr kb search <kb_id> "catalyst deactivation mechanism" -o json
```

Capture stable document IDs from `document list` before reading or deleting:

```bash
bohr kb document get <document_id> -o json
```

Upload sends the exact local file to Bohrium. Confirm the file, destination KB
or parent node, and displayed name before the write. All current KB commands
are free, but free does not imply upload or deletion consent.

Both KB deletion and document deletion are destructive. Resolve the exact
target immediately before adding `--yes`:

```bash
bohr kb document delete <document_id> --yes -o json
bohr kb delete <kb_id> --yes -o json
```

The current CLI does not manage folders, tags, notes, favorites, sharing, or
permissions. Read [`references/commands.md`](references/commands.md) for the
complete command set, identifier boundaries, upload routing, and deletion gate.
