---
name: bohrium-dataset
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Create, list, inspect, version, download, resume, or delete Bohrium datasets
  with `bohr dataset`. Use for 数据集, dataset uploads, dataset IDs/version IDs,
  project-owned data, resumable transfers, or dataset archives. Use `bohr file`
  for personal/share disk objects that are not dataset versions.
---

# Bohrium Dataset

Use read commands to resolve stable IDs before transferring or deleting data:

```bash
bohr dataset list --number 50 --title "training" -o json
bohr dataset get <dataset_id> -o json
bohr dataset versions <dataset_id> -o json
```

Creating a dataset uploads the selected local path to Bohrium. Confirm the
exact path, dataset identity, and owning project before the write:

```bash
bohr dataset create --name training-set --path training-set \
  --pid <project_id> --lp ./data --upload-concurrency 5 -o json
```

Interrupted uploads can resume only when `--name`, `--path`, `--pid`, and
`--lp` remain unchanged. Add `--resume` to reuse compatible cached state.
`--upload-concurrency` is globally bounded to `1-16`; higher values increase
simultaneous HTTP requests, not the number of dataset versions.

Downloads automatically use a stable `.part` file for resume:

```bash
bohr dataset download <dataset_id> --version-id <version_id> --dest ./dataset.zip
```

Use `--force` only when the user wants to discard an existing destination and
partial artifact and start over.

Dataset operations are currently free, but create uploads user data and delete
is destructive. Before `bohr dataset delete <dataset_id> --yes`, verify the
exact dataset ID and destructive intent. Free does not imply mutation consent.

The current CLI can list existing versions but cannot create a new version,
edit metadata, manage permissions, or query quota through dataset subcommands.

Read [`references/commands.md`](references/commands.md) for every command,
flag, resume invariant, local-file effect, and deletion boundary.
