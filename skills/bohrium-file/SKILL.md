---
name: bohrium-file
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Manage files and directories on Bohrium personal or project-share disks with
  `bohr file`: list, stat, download, upload, mkdir, copy, move, or delete. Use
  for personal/, share/, 云盘文件, project shared storage, recursive prefixes,
  and local/remote file transfer. Use Dataset commands for versioned datasets.
---

# Bohrium File

## Choose the task

| Goal | Route |
|---|---|
| Send/receive one disk object | Resolve space/project/path → transfer → stat/receipt verification |
| Copy or download a directory tree | List the narrow prefix → explicit recursive operation → inspect per-file outcomes |
| Recover an upload | Preserve local file and remote target → inspect uncertain outcome → resume compatible session |

## Workflow: transfer a project input and verify the round trip

1. Confirm the local source and destination, project ownership and intended remote key. A share disk path requires the same positive project ID on every call.

```bash
bohr file upload ./input.dat share/experiment/input.dat --project-id <project_id> -o json
bohr file stat share/experiment/input.dat --project-id <project_id> -o json
bohr file download share/experiment/input.dat --project-id <project_id> --dest ./retrieved-input.dat -o json
```

2. Upload success returns `data.remote_path`, `space`, `project_id` and `size`. Carry the returned full `remote_path` and project scope into stat/download; a basename alone loses the storage route. Check remote metadata against the intended object and expected size.
3. Download success reports `data.dest_path`/`size`. Complete with the verified remote target and actual local destination; for critical content, compare bytes/checksum locally when that verification is available. Size equality alone is not content integrity.
4. For recursive downloads, inspect `downloaded`, `skipped` and `failed`; `files` is not the count newly transferred. Report partial delivery and failed paths rather than treating a populated directory as complete.
5. On `UPLOAD_NOT_CONFIRMED`, inspect the same remote object and the returned recovery hint before another write. Apply the multipart resume rules below for compatible interrupted uploads. A failed download preserves the previous complete destination; retrying is a read plus an authorized local-file write.

Use canonical path prefixes:

- `personal/<path>` addresses the current user's personal disk and requires
  `--project-id 0` (the default).
- `share/<path>` addresses a project share disk and requires a positive
  `--project-id` for the owning project.

Resolve a target before changing it:

```bash
bohr file list personal/ --limit 50 -o json
bohr file stat personal/results/run.log -o json
bohr file list share/project-data --project-id <project_id> -o json
```

Canonical write and transfer commands are:

```bash
bohr file upload ./input.dat personal/jobs/input.dat -o json
bohr file upload ./large.tar personal/jobs/large.tar --upload-concurrency 8 -o json
bohr file upload ./large.tar personal/jobs/large.tar --upload-concurrency 8 --resume -o json
bohr file download personal/jobs/output.dat --dest ./output.dat
bohr file download personal/jobs/results/ --recursive --dest ./results
bohr file mkdir share/new-run --project-id <project_id> -o json
bohr file copy personal/a personal/b -o json
bohr file move share/old share/new --project-id <project_id> -o json
```

Confirm the exact local source/destination and remote path before uploads or
downloads. Copy and move affect a single object unless `--recursive` is
explicitly selected for a prefix.

Uploads at or above 50 MiB use 25 MiB multipart chunks. Set a bounded global
request limit with `--upload-concurrency` (1–16, default 5). After an
interruption, keep the same local file and remote path and add `--resume`; the
CLI reuses completed chunks. Progress goes to stderr, so machine stdout remains
one result envelope. The File append contract keeps one object's chunks ordered,
so its observed peak is 1 even when the configured upper bound is higher.
Smaller binary uploads restart from byte zero.

Delete is destructive. List/stat the exact target and use recursive deletion
only for an intentionally selected prefix:

```bash
bohr file delete share/obsolete --project-id <project_id> --recursive --yes -o json
```

All current File commands are free, but writes, local overwrites, moves, and
deletion still require user intent. The disk space comes only from the
`personal/` or `share/` path prefix.

The current CLI does not expose decompression or server-side filename/content
search. Read [`references/commands.md`](references/commands.md) for every
canonical command, flag, upload result field, recovery warning, path invariant,
and confirmation boundary.
