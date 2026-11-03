# File command reference

| Command | Arguments and flags | Effect |
|---|---|---|
| `bohr file list [path]` | `--project-id`, `--limit` default `50`, `--next-token` from `data.pagination.next_token` | Read one directory/prefix page; follow `has_more` |
| `bohr file stat <path>` | `--project-id` | Read metadata |
| `bohr file download <path>` | `--project-id`, `--dest`, `--recursive`, `--concurrency` default `5` (1–16) | Write a local file, or a directory tree with `--recursive`. Recursive download lists in 1000-object pages and starts transferring after the first page; retries retryable transfers up to 3 attempts per file; skips size-matched local files; continues after a per-file failure (`PARTIAL_FAILURE`). |
| `bohr file upload <local-file> [remote-path]` | `--project-id`, `--upload-concurrency` (1–16, default `5`), `--resume` | Upload local data |
| `bohr file mkdir <path>` | `--project-id` | Create remote directory marker |
| `bohr file copy <source> <destination>` | `--project-id`, `--recursive` | Remote copy |
| `bohr file move <source> <destination>` | `--project-id`, `--recursive` | Remote move/rename |
| `bohr file delete <path>` | `--project-id`, `--recursive` | Destructive remote delete |

Use canonical commands rather than the hidden `+list`, `+stat`, `+download`,
and `+upload` compatibility aliases.

## Path and project rules

- `personal/` paths use project ID `0`.
- `share/` paths require a positive project ID.
- Keep source and destination in a route the current command accepts; do not
  silently reinterpret a personal path as project-share storage.
- Recursive copy/move/delete/download operates on every object under a prefix.
  Directory download without `--recursive` fails with `FILE_IS_DIRECTORY`.

## Local effects

Before download, resolve the local `--dest` and check overwrite intent. A failed transfer preserves the old complete destination. File downloads use full requests, never Range; only recursive downloads retry internally (3 attempts). Before
upload, verify the local file and remote destination. Dry run validates routing
but is not proof of remote existence or permission.

Delete requires `--yes` in non-interactive use. Copy, move, mkdir, and upload
are non-destructive writes and do not use the destructive gate, but still need
task authorization.

## Upload recovery

- Files smaller than 50 MiB use one binary request and cannot resume partial
  bytes. Files at or above 50 MiB use 25 MiB multipart chunks.
- An interrupted multipart upload keeps compatible state. Re-run with the same
  environment, local file, and remote path plus `--resume` to send only missing
  chunks. A run without `--resume` starts a new session.
- Upload progress is written to stderr. JSON/YAML stdout remains one envelope.
- File storage requires ordered appends for one object. `--upload-concurrency`
  remains the shared uploader's upper bound, while one File upload has an
  observed peak of 1.
- Machine success data adds `uploaded_parts`, `resumed_parts`,
  `upload_concurrency`, and `peak_upload_concurrency`.
- Resume state lives under the OS user cache at `bohr/file-upload/`;
  `BOHR_FILE_UPLOAD_CACHE_DIR` overrides the cache root for isolated runs. The
  cache contains no access key, upload token, or raw storage parameter.
- Recovery warnings use `FILE_UPLOAD_SOURCE_CHANGED`,
  `FILE_UPLOAD_STATE_INVALID`, `FILE_UPLOAD_SESSION_REBUILT`, and
  `FILE_UPLOAD_CACHE_CLEANUP_FAILED`.

```bash
bohr guide read bohr-file
bohr guide read bohr-file/references/commands.md
```

## Download recovery

Use `error.code` and `retryable` together. `NETWORK_ERROR`, header/body silence (`UPSTREAM_TIMEOUT`) and incomplete bytes (`UPSTREAM_INVALID_RESPONSE`) are transport failures; their `http` is omitted when no non-2xx response occurred. `UPSTREAM_ERROR` retains the real status; 408/429/5xx can retry. `DOWNLOAD_CANCELLED` is caller cancellation, exit 1, not automatically retryable. A recursive rerun skips only regular local files matching an explicitly supplied remote size. Progress stays on stderr and the final envelope stays on stdout.
