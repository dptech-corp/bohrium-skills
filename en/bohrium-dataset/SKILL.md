---
name: bohrium-dataset
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Create, list, inspect, version, download, resume, or delete Bohrium datasets
  with `bohr dataset`. Use for 数据集, dataset uploads, dataset IDs/version IDs,
  project-owned data, dataset sharing/permissions, resumable transfers, or archives. Use `bohr file`
  for personal/share disk objects that are not dataset versions.
---

# Bohrium Dataset

## Choose the task

| Goal | Route |
|---|---|
| Publish local data | Confirm owner/source → create/upload → inspect dataset and version |
| Obtain reproducible input | Resolve dataset ID → select a ready version → download and verify ZIP |
| Recover interrupted transfer | Same upload identity + `--resume`, or same download version/destination |

## Workflow: publish and retrieve an exact dataset version

1. Confirm project, local tree and intended dataset identity, then use the create command below. Preserve `data.dataset_id`; retain `data.version_id` when returned. Do not confuse dataset ID with version ID.
2. Inspect the dataset and its existing versions:

```bash
bohr dataset get <dataset_id> -o json
bohr dataset versions <dataset_id> -o json
bohr dataset download <dataset_id> --version-id <version_id> --dest ./dataset.zip -o json
```

3. Select `<version_id>` from the chosen `data.items[].id` in `versions`. The downloadable version must be `status=committed`, `packStatus=completed`, with a usable `downloadUri`. An accepted upload or a version still packaging is not a ready input archive; recheck that version within the waiting budget.
4. Download that exact version, and complete only when the receipt has `data.verified=true`; report its `path`/`bytes` and the dataset/version binding. Preserve the ZIP as an archive until the downstream task explicitly needs extraction.
5. On interrupted upload, retain all four identity inputs and add `--resume` as described below. On download interruption, retain the version and destination so `.part` recovery remains bound to the same data. `--force` discards recovery state and needs overwrite intent; changing version is not a safe continuation of the original partial download.

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
edit metadata, or query quota through dataset subcommands.

## Inspect and change dataset access

Resolve the dataset and intended user/project IDs, then read its current access list. Granting or changing roles requires dataset management rights and clear write intent; revoking access is destructive and requires `--yes` outside a terminal.

```bash
bohr dataset permission <dataset_id> -o json
bohr dataset share <dataset_id> --user <user_id> --role member -o json
bohr dataset share <dataset_id> --project <project_id> --role manager -o json
bohr dataset unshare <dataset_id> --user <user_id> --yes -o json
```

`--user` and `--project` are repeatable and may be combined. `--role` is `member` (default) or `manager`; manager also controls access. Keep principal kind and ID together in `data.items[]`; a user ID is not a project ID. Creator access and the owning project's access are protected.

The CLI reads, merges, writes and re-reads the whole permission list, preserving other entries; it has no concurrency lock. On `RESOURCE_CONFLICT`, inspect the returned actual list and run `permission` again before deciding on another write. Complete only after the actual list matches the intended grants/revocations; do not treat request acceptance as proof.

Read [`references/commands.md`](references/commands.md) for every command,
flag, resume invariant, local-file effect, and deletion boundary.
