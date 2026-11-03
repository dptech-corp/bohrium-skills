# Dataset command reference

| Command | Current inputs | Effect |
|---|---|---|
| `bohr dataset list` | `--number/-n` default `50`, `--project-id`, `--title/-t` | Read catalog |
| `bohr dataset get <id>` | dataset ID | Read dataset detail |
| `bohr dataset versions <id>` | dataset ID | Read existing versions |
| `bohr dataset create` | see below | Create and upload |
| `bohr dataset download <dataset-id>` | `--version-id`, `--dest`, `--force` | Write local ZIP/partial |
| `bohr dataset delete <datasetId>...` | positional IDs or repeatable `--dataset-id` | Destructive remote delete |

## Create and resume

```text
bohr dataset create --name <name> --path <dataset-path> --pid <project-id> --lp <local-path> [flags]
```

- `--comment/-m` sets the description.
- `--upload-concurrency` defaults to `5` and accepts `1-16`.
- `--resume` reuses only a compatible cached upload session.

Resume identity is the tuple of name, dataset path, project ID, and canonical
local path. Changing any member starts a different upload; do not use resume to
merge unrelated local trees.

## Download and local artifacts

`--version-id` selects an exact existing version. `--dest` selects the ZIP
path. Interruption keeps `<dest>.part` and `<dest>.part.meta`. Resume requires a matching fixed dataset/version binding; an old unbound partial needs explicit `--force` to restart. Keep the same destination and version when retrying. `--force` resets partial/metadata/invalid artifacts and replaces the old complete ZIP only after the new archive passes CRC verification; obtain overwrite intent first.

Ordinary recovery has 4 total attempts with cancellable 1/2/4-second delays. Signed URL refresh has an independent limit of 3 and stays on the selected version. Headers have a 30-second limit; stalled network reads have 60 seconds. Continuous body progress has no total limit. Classify failures with the same NETWORK_ERROR / UPSTREAM_TIMEOUT / UPSTREAM_INVALID_RESPONSE / UPSTREAM_ERROR vocabulary as file downloads; `http` only carries real non-2xx responses. Caller cancellation is DOWNLOAD_CANCELLED, exit 1 and retryable=false. ZIP corruption remains DATASET_ARCHIVE_INVALID; an existing destination remains DATASET_DOWNLOAD_EXISTS. Raw unsigned/partial bytes are never a completed ZIP.

## Delete

```bash
bohr dataset delete <dataset_id> --yes -o json
```

Delete is destructive. Resolve IDs through list/get, show the exact targets,
and obtain confirmation immediately before the command. Do not infer deletion
completion from an accepted request unless the returned envelope says it
completed.

All current dataset commands have free billing annotations. Upload, overwrite,
local download, and delete effects still require user intent.

```bash
bohr guide read bohr-dataset
bohr guide read bohr-dataset/references/commands.md
```
