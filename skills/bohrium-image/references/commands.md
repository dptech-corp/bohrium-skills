# Image command reference

| Command | Inputs | Risk/billing |
|---|---|---|
| `bohr image list` | `--type/-t`; `--page` default `1`, `--size` default `100` (`1-512`, server can cap at 100); hidden `--limit` compatibility alias | Free paged read |
| `bohr image search [keyword...]` | software capability AND search; repeatable `--capability` / `--framework`; `--by-name` requires keyword and excludes those filters; `--page`, `--page-size` | Free public-catalog read |
| `bohr image packages [imageId]` | catalog ID or mutually exclusive `--image <full-reference>` | Free software-facts read |
| `bohr image get <id>` | image ID | Free read |
| `bohr image status <image_id>` | build/commit task ID; reads once, never resubmits | Free read |
| `bohr image pull <image_address>` | `--image_address/-i`, `--timeout` | Free local Docker write |
| `bohr image push <local_image> <image_name>` | `--project-id` or `BOHR_PROJECT_ID`; `--desc`, `--timeout` | Image write; local Docker required |
| `bohr image delete <imageId>...` | positional IDs or `--image-id` | Destructive |
| `bohr image build` | Dockerfile/name/project flags; `--mode remote|local` | Remote: balance-only; local: prepare record and Docker build |
| `bohr image commit` | sandbox/name/project flags | Balance-only resource |
| `bohr image build-log <image_id>` | `--follow/-f`, `--ansi auto|never|always` | Free read |
| `bohr image dockerfile <image_id>` | image ID | Free read |

## Build

Required: `--dockerfile`, `--name`, and a project ID supplied by
`--project-id` or configured project context. Default `--mode remote` accepts
`--desc`, `--sandbox-id`, `--wait`, and `--wait-timeout` (default `1800`
seconds). The Dockerfile must be readable UTF-8 and at most 64 KiB. Only its
text is sent for a remote build; there is no local build context, so `COPY`/`ADD`
of local files cannot find them. Fetch inputs with `RUN` or `ADD <url>`, or use a base image that
already contains them. `--sandbox-id` does not provide
sandbox files or create a snapshot; use `bohr image commit` for a snapshot.

`--mode local` uses the Docker CLI on PATH with `--context` (default `.`) and
`--timeout` (default `30m`). It prepares a remote image record and consumes quota,
then builds with the service-assigned full reference. The result includes `id`,
`localImage`, `imageUrl`, `uploaded: false`, and `pushCommand`; run that command
separately to upload. Keep Docker context/daemon, Bohrium profile and identity,
project and short name identical for build and push; push prepares again and
reuses the pending record. Use unique names and avoid concurrent same-name
operations. Known build failure or cancellation attempts abort. If prepare
does not return a reliable result, inspect the state first: even an error can
leave a record, so do not assume it is absent or was cleaned up. After local
success, run `pushCommand`; do not poll `image status` or `build-log` with the
prepare record ID. Dry-run does not reserve, build or upload.
Local build does not submit remote compute or require `--yes`. Remote rejects
`--context` and `--timeout`; local rejects `--wait`, `--wait-timeout`, and
`--sandbox-id`.

## Commit

Required: `--sandbox-id`, `--name`, and project ID. Optional: `--desc`,
`--wait`, and `--wait-timeout` default `1800` seconds. Commit captures the
selected sandbox container root filesystem on the server; it does not use a local
Docker image and is not a source-control commit. Do not assume mounted volume
contents are included; place needed files in the container root filesystem first.

Before submit, actively preflight the running rootfs. This applies to
`bohr image commit`, not `bohr image build --dockerfile`, and must finish
before commit starts. An observed 12-15 minute temporary apply burst can reach
ENOSPC before an 80% background threshold reacts.

```bash
bohr sandbox storage inspect <sandbox_id> -o json
```

Require `status=complete`, no warnings, and non-null live `total_bytes`,
`used_bytes`, and `available_bytes`. Let `U = used_bytes` and let `D` be a
defensible estimate of the uncompressed rootfs change/new-file delta, not a
compressed registry layer size:

`required_bytes = max(ceil(U * 2.2), U + 3 * D)`

Set `N = ceil(required_bytes / 1GiB)`. This is an initial failure-data
heuristic, not a guarantee. The current CLI does not expose last-commit delta.
If `D` is unknown, never substitute `D = 0`; stop automatic submission and ask
for an estimate or explicit approval of a conservative target.

If live capacity already satisfies `required_bytes`, do not resize. Otherwise
check `bohr sandbox quota -o json`; never silently clamp an over-quota target.
`max_extra_ephemeral_storage_gb.limit` is extra above the 30Gi base, not total
rootfs capacity. Quota gates only a requested grow; an already-large sandbox
with sufficient measured capacity must not be blocked by a lower current quota.
Then request the smallest sufficient whole-Gi total, not an increment:

```bash
bohr sandbox storage resize <sandbox_id> --target <N>Gi --timeout 180 -o json
```

Resize is grow-only: it cannot shrink the rootfs. Never pass a smaller target
to reclaim space or assume that it will reduce an already-expanded sandbox.

Proceed only after `status=ready`, `ready=true`, and
`pod_uid_unchanged=true`. Recompute from the final measurement if writes
occurred while waiting. Partial inspection, quota rejection, timeout, Pod
replacement, or not-ready evidence stops the Agent before the separate
`bohr image commit` submission. The CLI does not perform this preflight or
submit commit automatically.

## Status and logs

After an interrupted remote build or sandbox commit `--wait`, use
`bohr image status <image_id> -o json` with the returned task ID. This reads `/sandbox_work/image/commit/:id`;
`image get` still reads the catalog. Exit zero includes pending states, so
require `data.status=2` before consuming the image. Failed/canceled states
exit with 3/4. A local wait timeout does not authorize a duplicate submission.

| Status | Meaning | Terminal/exit behavior |
|---:|---|---|
| `0` | creating | non-terminal |
| `1` | pending | non-terminal |
| `2` | success | terminal, exit `0` |
| `3` | failed | terminal, exit `3` |
| `4` | canceled | terminal, exit `4` |
| `6` | build-timeout leaked by legacy backend | inspect diagnostics; not a launchable success |

Use `errorDetail` / `errorMsg`, digest, and image URL from the final record.
`build-log --follow` streams until the build finishes but is not itself a
replacement for the final status record.

## Billing and confirmation

```bash
bohr billing pricing --resource sandbox -o json
```

For remote build or commit, inspect the `sandbox.image_build` or `sandbox.image_commit` item. Price depends
on the selected operation/resource contract, so do not invent a fixed amount.
Add `--yes` only after spending is authorized. Delete needs separate destructive
authorization even though deletion is free.

Pull requires local Docker and accepts only Bohrium image addresses. It performs
local registry login and pull; progress belongs on stderr while structured
stdout stays machine-readable.

Push invokes the Docker CLI on `PATH` and requires an image already present in
the Docker daemon. Prepare it with `docker build`, `docker load`, or `docker pull`;
push does not run a build or upload a Dockerfile or build context. It requires a
short build/commit-style image name, not a remote registry argument. Its order
is prepare → `docker tag` → `docker login` → `docker push` → complete. Only
prepare's canonical `imageUrl` is rewritten
onto the AccessKey-enabled registry face; never build a project/user path in
the CLI. Login uses `LoginViaBohriumAksk` with the AccessKey on stdin only.

A Docker-step failure triggers best-effort abort while preserving the original
error. After complete starts, a failure does not abort: Docker push already
succeeded and completion is unconfirmed. Dry-run performs no HTTP or Docker
operations and shows the full lifecycle with prepare placeholders. Successful
JSON data is the complete response, including ID, final image URL, status,
digest, and size in bytes. Progress belongs on stderr.

```bash
bohr guide read bohr-image
bohr guide read bohr-image/references/commands.md
```
