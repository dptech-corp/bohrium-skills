# Image commit & Dockerfile build

`bohr image commit` / `bohr image build` produce container images.
The former `bohr sandbox image …` alias has been removed; use `bohr image`.

Both are billable, so both need `--yes` when stdin is not a terminal (an agent's
always is not): since 2.5.18 the CLI refuses billable commands without
confirmation, returning `CONFIRMATION_REQUIRED` with exit code 10 and building
nothing. See the "Billable commands" section in the top-level SKILL.md.

Two ways to mint a new container image; both are async. Poll with
`--wait` on submit (`buildType` tells commit and Dockerfile build apart):

| Subcommand | When to use | Source |
|---|---|---|
| `image commit` | snapshot a *running* sandbox's filesystem | needs `--sandbox-id` |
| `image build`  | reproducible build from source              | needs `--dockerfile <file>` |

## Mandatory preflight before a sandbox snapshot commit

This workflow applies only to canonical `bohr image commit`, which snapshots a
running sandbox. It does not apply to `bohr image build --dockerfile`; an
optional sandbox ID on a Dockerfile build is provenance only.

Do the capacity check before submitting commit, not after it starts. Observed
commit apply work can create a 12-15 minute burst of temporary rootfs use, so
an 80% background threshold is not a commit safety gate: ENOSPC may happen
before a reactive resize becomes Ready.

Before submitting `image commit`, run:

```bash
bohr sandbox storage inspect <sandbox_id> -o json
```

If this returns `sandbox_storage_management_disabled`, enter the same sandbox
and run `df -Pk /` plus the actual temporary/output paths. Use these live
1024-byte-block measurements for the capacity checks below; do not count free
space on the same filesystem twice. This fallback does not enable resize. If
the read fails or more space is required while resize is unavailable, stop
before commit and choose another storage plan.

For the rootfs decision, require valid `total_bytes`, `used_bytes`, and
`available_bytes` and resolve any warning that applies to rootfs or its identity.
A `partial` report may retain known rootfs data if only the selected temporary
directory was unavailable; that directory must also be checked if commit uses it.
Any required `null` measurement is unknown, not zero. `status=complete` only
describes inspection evidence; it is not proof that commit has enough space.

Let `U = rootfs.measurement.used_bytes` and let `D` be a defensible estimate of
the uncompressed rootfs change/new-file delta that commit must apply. `D` is
not a compressed registry layer size. Use the initial failure-data heuristic:

`required_bytes = max(ceil(U * 2.2), U + 3 * D)`

Then set `N = ceil(required_bytes / 1GiB)`. The coefficients are a conservative
starting estimate to calibrate with later failure samples, not a platform SLA
or a guarantee. The current storage/image responses do not expose a
last-commit delta. If `D` cannot be obtained from user input or reliable workload data,
do not silently use `D = 0` or claim that `U * 2.2` is sufficient. Stop
automatic submission and ask for an estimate or explicit approval of a
conservative target.

If current live capacity already satisfies `required_bytes`, do not resize.
Only when growth is required, first require
`capabilities.dynamic_adjustment.available=true`. False means the current API
cannot resize this sandbox; absent or null means unknown. Read its reason and
next action and choose a different storage plan before commit; do not repeatedly
try resize or silently replace the sandbox. Then check `bohr sandbox quota -o json`. If the target
is rejected or exceeds the effective extra-storage allowance, the Agent
must not silently clamp it and continue. Report the required target and
shortfall, then stop before commit. In quota output,
`max_extra_ephemeral_storage_gb.limit` is extra above the 30Gi base, not total
rootfs capacity. Quota gates a requested grow only: do not block an
already-large sandbox whose measured live capacity is sufficient merely
because its current quota is lower.

Otherwise request the smallest whole-Gi total that satisfies the estimate:

```bash
bohr sandbox storage resize <sandbox_id> --target <N>Gi --timeout 180 -o json
```

`--target` is the total rootfs capacity after resize, not an increment. Resize
is grow-only and cannot shrink the rootfs; never use a smaller target to
reclaim space. Require
`status=ready`, `ready=true`, and `pod_uid_unchanged=true`; a PATCH response or
annotation alone is not Ready. Use the resize result's final live measurement,
or immediately inspect again, and recompute if writes occurred while waiting.
Only after that gate passes may the Agent submit the separate canonical
`bohr image commit`. Partial inspection, quota rejection, timeout, Pod
replacement, or any not-ready result must stop before commit.

The CLI does not yet estimate `D`, preflight commit automatically, or make
resize and commit atomic. Storage resize does not automatically submit image
commit. See [`sandbox/storage.md`](sandbox/storage.md) for the full evidence
model.

Without `--wait`, both jobs return the new image record's `id` and reserved
`imageUrl` immediately so you can pre-wire
`bohr sandbox template create --image <imageUrl>`. That is **not** success:
only `status=2` means the image is ready. Use `--wait` before consuming
the image.

`bohr image list` lists the **catalog**; `bohr image get <id>` is the catalog
private image. There is currently no CLI listing of build/commit *tasks* —
`--sandbox-id` / `--status` / `--page` style filters do not exist on either
command.

To recover observation of a known task after interruption, use
`bohr image status <image_id> -o json`. It reads the existing build/commit task
once without resubmission; `image get` remains a catalog read. Require status=2:
a successful status query can also report a pending task. Failed/canceled
states exit with 3/4. This command does not recover an ID lost during submission.

Human build/commit output ends with `Template-ready image: <imageUrl>` so the
reserved immutable reference is easy to copy. Structured output is unchanged;
read `data.response.imageUrl` when using `--output json`.

Use `--wait` on commit/build if you want to block until `success` / `failed`
before consuming the image.

```bash
# A) commit — snapshot a running sandbox
bohr image commit \
    --sandbox-id <sandbox_id> \
    --name <image-name> \
    --desc "training-day-3 snapshot" \
    --project-id <project_id> \
    --yes \
    -o json

# B) build — from a local Dockerfile (no sandbox needed)
bohr image build \
    --dockerfile ./Dockerfile \
    --name <image-name> \
    --desc "pinned cuda 12.1" \
    --project-id <project_id> \
    --yes \
    -o json

# block until success/failed/canceled; -o json is one final get-shaped record
bohr image commit \
    --sandbox-id <sandbox_id> --name <image-name> --project-id <project_id> \
    --yes --wait -o json
# exit after --wait: 0 success · 3 failed · 4 canceled · 1 timeout

# catalog private image (not a commit/build poll)
bohr image get <id> -o json

# recover the exact Dockerfile stored for a Dockerfile build
bohr image dockerfile <id>                  # decoded source, exact bytes
bohr image dockerfile <id> -o json          # envelope with raw base64 dockerFile

# fetch the kaniko log of a dockerfile build (commit records have no per-line log)
bohr image build-log <id>            # raw text snapshot to stdout, pipe-friendly
bohr image build-log <id> --follow   # stream chunks as kaniko writes them,
                                     # auto-EOFs when the build reaches a
                                     # terminal status (success/failed)
bohr image build-log <id> --ansi never  # strip color codes (auto|never|always)

# browse catalog images
bohr image list
bohr image ls

```

## Naming — avoid a `latest` image name

`--name` is the user-visible portion of the image name; the backend may prepend
a user/project prefix and controls the final registry tag. If the name ends in
`latest`, the CLI prints a **stderr hint** while keeping structured stdout
clean. Prefer an immutable/unique name instead: a `latest` image is a poor
template base — `bohr sandbox template create/update` **rejects** a `:latest`
template image, and the platform's prewarmed image cache cannot be reused for a
mutable tag. Prefer a version (`:v1`) or a date/build tag, then use the returned
`Template-ready image` reference with `template create --image`.

## Status & failure reason

`status` / `statusText` is the only lifecycle field
(`0=creating 1=pending 2=success 3=failed 4=canceled`).
Do not treat `imageUrl` as success. Human labels may still show historical
`6=build-timeout` on old rows.

On `status=3`, read `errorMsg` / `errorDetail` from `--wait -o json`.
That is the original diagnostic; decide the next action from it.
`idempotentReplay=true` is a replay of an already-failed submit.

`--wait` exit codes: `0` success, `3` failed, `4` canceled, `1` wait timeout.
After `--wait`, read `data.errorDetail` / `data.errorMsg`.

Use `--follow` only with **dockerfile** builds; commit records only expose a few
coarse phase messages, not per-line kaniko stdout. Inspect commit failures via
`--wait -o json` (`errorMsg` / `errorDetail`) instead.

Both `commit` and `build` require a project id. Pass `--project-id <id>`, or set
`BOHR_PROJECT_ID` in the environment (the flag wins). If neither is set the CLI
refuses the call with `missing project id` instead of letting the backend return
a generic 400.

`image build` reads the Dockerfile locally and sends it inline; the hard cap is
**64 KiB**. The CLI rejects oversized files client-side. The `-o json` echo of
the request elides the Dockerfile body (only `dockerfileBytes` is shown) so logs
stay clean; the response (`id`, `imageUrl`) is unmodified.

`image dockerfile <id>` reads the persisted source through the private-image
detail endpoint and base64-decodes it to stdout without adding or removing a
trailing newline. Use `-o json` when an agent needs the original envelope and
base64 `data.dockerFile` value. Commit/import records do not have stored
Dockerfile source; the command returns `RESOURCE_NOT_FOUND` instead of an empty
success for those records.
