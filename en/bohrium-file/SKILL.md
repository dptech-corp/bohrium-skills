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
