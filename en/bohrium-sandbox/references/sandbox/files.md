# Sandbox files

Upload and download files and directories.

## Uploading files and directories

`files write --source` accepts either a single file or a directory. When given
a directory, every regular file under it is uploaded with relative paths
preserved under the remote root. For very large trees, tar locally then untar
inside the sandbox via `exec`.

Each file is checked against the size the sandbox reports after the write. A
short copy fails with `UPLOAD_INCOMPLETE`; `data.remote_path`,
`data.sent_bytes`, `data.remote_bytes` and `data.verified_files` say which file
and how much of it arrived. An upload that is cut off or stops moving fails
with `UPLOAD_NOT_CONFIRMED` (`retryable: true`, `data.unconfirmed: true`), or
with `OPERATION_TIMEOUT` once the CLI has seen no progress for 90 seconds; the
sandbox may close a stopped upload sooner, so expect the first code. Its
`data.remote_path` and `data.sent_bytes` name the file, and `data.remote_bytes`
is the size now there when the sandbox still answers. A slow upload that keeps
moving is not cut off. After any of these failures the remote file may be
partial: re-run the same command to overwrite it (there is no resume).

```bash
bohr sandbox files write --source ./run.py <sandbox_id> /workspace/run.py -o json
bohr sandbox files write --source ./project <sandbox_id> /workspace/project -o json
bohr sandbox files write --content 'hello' <sandbox_id> /workspace/note.txt -o json
```

## Fast uploads to /bohr-workspace (tiefblue, `--ti`)

`/bohr-workspace` is not a sandbox-local disk — it is a **persistent,
session-scoped directory backed by object storage (tiefblue)**. Files written
there survive `delete` and are visible to any sandbox created with the same
session id. Pass `--ti` to upload a file/directory **straight to that object
storage** instead of streaming it through the sandbox filesystem: this is much
faster for large files (automatic multipart) and skips the sandbox hop.

```bash
# a single file
bohr sandbox files write --ti --session-id <session-id> --source ./model.pt \
  <sandbox_id> /bohr-workspace/model.pt -o json
# a whole directory
bohr sandbox files write --ti --session-id <session-id> --source ./out_dir \
  <sandbox_id> /bohr-workspace/out -o json
```

Rules and caveats:

- **`--session-id` is required** and must match the id the sandbox was created
  with (`bohr sandbox create --session-id <id>`). `/bohr-workspace` is bound to
  the session, so the token is scoped per (user, session).
- Only paths under **`/bohr-workspace`** (or its sub-directories) are supported.
  For any other path, upload without `--ti` (via the sandbox filesystem).
- `--content` is not supported with `--ti`; pass a real `--source` file or dir.
- **Directory uploads are recursive with no dry-run or confirmation** — every
  file under `--source` (including hidden files, `.git/`, `node_modules/`, any
  `.env` secrets) is pushed and overwrites existing remote objects. Double-check
  the source path and prune unwanted files first; for big/messy trees prefer
  `tar` + `exec`.
- The `-o json` result reports `objectKey`, `sandboxPath`, `host`, and `bytes`
  per file.
- Files uploaded via `--ti` appear inside the sandbox owned by `root:root` with
  mode `640`. The default root `exec` user can read them directly; if you
  override `--user` to a non-root identity, `chown`/`chmod` them as root first.

## Downloading files

`files read` prints a file to stdout, or writes it to local disk with
`--destination` / `-D`. The global `--output` / `-o` flag only selects the CLI
output format.

With `--destination`, a regular file is streamed into a temporary file beside
the target and published only after its byte count matches the remote size.
An interrupted transfer or size mismatch returns `DOWNLOAD_INCOMPLETE` and
keeps an existing target intact. A file target must be a regular file or a new
path: symlinks and special files are rejected. Keep space for both copies.
Temporary-file cleanup is best effort; a warning identifies any residue.
An interrupt after the final cancellation check can race with publication.
Only the byte count is checked: same-size changes are not detected. Directory
recursion retains the SDK behavior, and stdout reads have no publication step.
Do not let another process replace a directory download root during transfer.
Replacing a file does not preserve its owner, ACL or hard-link relationships;
on non-Unix platforms atomic replacement and ACL behavior are not guaranteed.

```bash
bohr sandbox files read <sandbox_id> /workspace/result.csv                        # print text to stdout
bohr sandbox files read <sandbox_id> /workspace/result.csv --destination ./result.csv  # save to disk
bohr sandbox files read <sandbox_id> /workspace/model.pt   --destination ./model.pt    # binary, auto-detected
```

### Text vs bytes (important for binary files)

When `--format` is omitted it defaults to **auto**, which detects the type from
the file extension: known binary types (`.pt`, `.pth`, `.bin`, `.ckpt`,
`.safetensors`, `.npy`, `.tar`, `.gz`, `.zip`, `.parquet`, `.pdf`, `.so`,
images, ...) are read as **bytes**, everything else as **text**. A
`binary file detected ...` warning is printed to stderr when this happens.

Reading a binary file as `text` runs it through a charset decode + UTF-8
re-encode roundtrip that **silently corrupts the bytes** (wrong size,
`torch.load` / unzip fails with "invalid header or archive is corrupted"). The
auto-detection prevents this for known extensions; for an unrecognized
extension on a binary file, pass `--format bytes` explicitly:

```bash
bohr sandbox files read <sandbox_id> /workspace/checkpoint.weights --format bytes --destination ./checkpoint.weights
```

When reading a single file to stdout, `--format bytes` emits base64 (the `-o json`
payload carries `content_base64` + `byte_count`); `--format text` validates
UTF-8 and carries `content`. Force `--format text` only when you knowingly want
the decoded string of a binary file (rare).


## SHA-256 transfer receipts

Non-`--ti` writes and `read --destination` include `data.manifest`: one item per
ordinary file, with `path`, `size`, lowercase `sha256` and `status`. A matching
local/remote digest is `verified`; an unavailable digest is `unverified` with
`unverified_reason`. A mismatch fails with `TRANSFER_HASH_MISMATCH`. The upload
hash covers the bytes actually streamed, including stdin/`--content`.

`--no-hash-verify` skips the remote comparison, still calculates local SHA-256,
and reports `unverified`/`skipped_by_flag`. It retains all size checks. Stdout
reads and `--ti` retain their own contracts.

Directory receipts cover the transferred files, using remote enumeration and
size/count checks rather than including unrelated existing local files.
Directory downloads use the SDK merge behavior: child files are not published
atomically as a whole tree, and neither enumeration nor hashes provide a remote
snapshot or prevent later changes. The current SDK requires child-only directory
listings; a self entry is rejected before download, with an explicit-file
export hint. Stop writers before relying on an export.
