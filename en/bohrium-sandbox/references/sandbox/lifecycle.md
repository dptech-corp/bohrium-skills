# Sandbox lifecycle

## Renew an existing sandbox

Use the existing sandbox ID when the current task needs more time. Inspect its lifetime before starting long work and renew before expiry if the remaining lifetime will not cover the work, output retrieval and a reasonable margin within the task's resource budget.

```bash
bohr sandbox describe <sandbox_id> -o json
# Request at least one hour of remaining lifetime from this call.
bohr sandbox connect --timeout 3600 <sandbox_id> -o json
bohr sandbox describe <sandbox_id> -o json
```

Replace `<sandbox_id>` with the ID already assigned to this task; keep the same environment/profile and identity throughout. Choose N from the remaining work, not from time already spent.

`connect --timeout` requests sandbox lifetime in seconds; it does not set the HTTP waiting budget. The default request is 300 seconds.

For a running, finite-lifetime sandbox on the OpenKruise backend, connect uses **extend-only** semantics: `new_deadline = max(old_deadline, now + N seconds)`. A later existing deadline remains unchanged. N is not added to the old deadline. A never-timeout sandbox stays unlimited; do not pass zero to connect to request unlimited lifetime. A paused sandbox may be resumed by connect; a deleted sandbox cannot be revived. Backend limits and provider support still apply.

After the call, check a trustworthy runtime `endAt` or, when available through an authorized read, the live Sandbox CR deadline. A successful connect receipt alone is not proof of the final deadline. `expires_at` in a business/list record is not the lifecycle deadline; it may be fixed far in the future. If the readback is missing or stale, report renewal as unverified rather than guessing.

The CLI does not renew continuously. Neither `exec --background` nor an active foreground command guarantees the sandbox stays alive. Recheck lifetime during long work and renew before expiry when the task still needs it. Do not recreate a sandbox or rerun its workload to perform renewal. On a renewal error, inspect the original sandbox's state and deadline before deciding what to do.

## Recover a sandbox creation

For `sandbox_create_in_progress`, `retryable=false` applies to the create POST;
`poll_create_request` means repeat the original receipt query until a terminal
state, then follow its action. Do not use the failed-instance list for this case.

Use `bohr sandbox describe --create-request-id <original-id> -o json` to query a
Sandbox create receipt. This is a general Sandbox operation, independent of
dataset mounts. It requires the server's create-request endpoint; an unsupported
endpoint or a single 404 does not prove that no sandbox was created.

The server scopes keys to the caller and rejects changed parameters for the same
key. Concurrent/repeated requests do not dispatch another runtime. A completed
key returns a conflict receipt with the existing `sandbox_id`; inspect that
sandbox instead of creating again. Only `REJECTED` with `retryable=true` and
`action=retry_same_request` permits another POST with the original key and body.
`PREPARING`, `CREATING` and `UNKNOWN` require reconciliation. Receipt state records
the creation attempt, not the sandbox's current lifecycle state.

Create, list, describe, delete, and check quota on on-demand sandboxes; plus best practices.

Every `create` below needs `--yes` when stdin is not a terminal (an agent's
always is not): since 2.5.18 the CLI refuses billable commands without
confirmation, returning `CONFIRMATION_REQUIRED` with exit code 10 and creating
nothing. See the "Billable commands" section in the top-level SKILL.md. The
listing/describe/quota commands here are unaffected.

```bash
bohr sandbox doctor -o json                                # verify control-plane configuration; verdict is data.ready
bohr sandbox quota -o json                                 # your effective caps + usage (Running+Hibernating)
bohr sandbox create my-template -o json                    # create from explicit template (personal wallet)
bohr sandbox create my-template --project-id <id> -o json  # bill the sandbox to a project budget
bohr sandbox create my-template --timeout 1800 -o json     # set sandbox auto-destroy lifetime (seconds)
bohr sandbox create my-template --never-timeout -o json    # request disabling the automatic timeout policy
bohr sandbox create my-template --mount-user-storage -o json # mount caller's personal + share disks
bohr sandbox create my-template --project-id <id> --mount-user-storage --share-subpath sub/dir -o json # mount only a sub-dir of the project share disk
bohr sandbox create my-template --env HELLO=WORLD --env DEBUG=1 -o json # inject environment variables
bohr sandbox create doc-compiler --inherit-auth -o json       # opt in to caller bohr-login inheritance
bohr sandbox create --image <image-path> -o json          # use the default 1c2g image profile
bohr sandbox create --image <image-path> --cpu 2c4g -o json # select the 2 CPU / 4 GiB profile
bohr sandbox create my-template --image <image-path> -o json # override image on an explicit template
bohr sandbox create <template> --wenyon-dataset-id <dataset-id> --wenyon-mount-path /mnt/wenyon --wenyon-user-token-env WENYON_USER_JWT --request-id <stable-request-id> -o json # instance mount
bohr sandbox create my-template --session-id <session-id> -o json # tag with a session id (shared user-storage subPath)
bohr sandbox create worker-template --parent-sandbox-id <driver_sandbox_id> -o json # record a child owned by a driver sandbox
bohr sandbox list -o json                                  # list your sandboxes
bohr sandbox list --query running                          # human parent/child tree
bohr sandbox list --query running --flat                   # legacy flat human table
bohr sandbox list --query failed -o json                   # recent failures + failed_reason/status_reason
bohr sandbox describe <sandbox_id> --with-processes -o json # metadata + running processes
bohr sandbox ps <sandbox_id> -o json                       # list processes in a sandbox
bohr sandbox exec -o json <sandbox_id> -- python script.py  # foreground command
bohr sandbox files write --source ./local_dir <sandbox_id> /workspace/dir -o json  # upload a file or directory
bohr sandbox files read <sandbox_id> /workspace/result.csv --destination ./result.csv
bohr sandbox delete <sandbox_id> -o json                   # destroy a sandbox (alias: kill)
```

**`bohr sandbox quota` — check before create; self-check after limit errors.**

`running_sandboxes.used` counts **Running + Hibernating** (same as create
enforcement) — hibernating still holds a slot. Branch on `-o json`
`remaining`; prefer reuse via `list` before creating another sandbox.

```bash
# Before create
bohr sandbox quota -o json
# remaining > 0 → create (or reuse); remaining == 0 → delete a slot first

# After 400 "running sandbox limit reached (N/M)..." — never blind-retry
bohr sandbox quota -o json
bohr sandbox list -o json
bohr sandbox delete <idle_or_hibernating_id> --force
bohr sandbox create <template> -o json
```

- `running sandbox limit reached` → `quota` → `list` → `delete` idle/hibernating → retry create.
- `template limit reached` → `quota` → `template rm` unused owned templates → retry.
- Extra ephemeral storage is a hard create-time check. The production target
  system default permits up to **50Gi extra** above the **30Gi base** (80Gi
  total), unless the user has an explicit override. A
  `max_extra_ephemeral_storage_gb.limit` of **0 is a real zero allowance**, not
  "unset"; a missing/NULL user override inherits the system default. A request
  above the effective limit returns **HTTP 400** and creates no template — it
  never silently falls back to 30Gi.
- Agents cannot raise the cap — if every sandbox is still needed, stop and ask the user.

## List all pages and probe one data-plane target

`list` remains single-page by default. Use `--all-pages` to aggregate pages and
`--max-pages N` to bound that operation. Later-page failures, contradictory or
repeated pages and an exhausted page budget retain fetched rows, return
nonzero and mark `data.incomplete=true`; the pagination reflects unconsumed
rows. A short page alone does not prove exhaustion. The current control plane
supports active/all/running/failed only; deleted-history filtering remains
blocked on server capability. The CLI does not infer deleted records or actors. With `--with-ttl`, a paging failure skips enrichment, and a TTL lookup
failure remains incomplete even if all pages were fetched.

Without new flags, `sandbox doctor` checks the control plane. Add both
`--sandbox <id> --target <http-or-https-url>` for an authorized existing instance
and a target-specific DNS/TCP/TLS/HTTP probe. The connect POST may extend the
instance lifetime. The report omits the URL path, response body and raw errors;
IP DNS and HTTP TLS are marked not applicable. Missing Python or a broken
execution channel makes the report unready rather than successful.

`dataplane.fault_stage` describes this probe's GET only: `pre_submit` means no
HTTP bytes were sent; `post_submit` means an HTTP response established a
submitted request but failed; an ambiguous send/channel failure is `unknown`.
It does not classify a user's earlier business operation. Read `data.ready`
and the layers: diagnosis completion still exits 0 even when unready.

## Payment selection

```bash
# Requires confirmed photon support for the target environment/account.
bohr sandbox create sdbxagent --payment-source photon --timeout 180 --yes
# Project budgets use cash.
bohr sandbox create sdbxagent --project-id 123 --payment-source balance --yes
```

`--payment-source` accepts exactly `photon` or `balance`; omission selects balance.
The CLI sends numeric `paymentType: 1` or `0` in the order body, not a gateway
payment header. Personal photon billing needs no project ID. Project budgets
cannot use photons. The legacy `--payment-type 0|1` remains compatible; if both
selectors are supplied they must agree. Invalid values and project/photon
conflicts fail before confirmation or creation. No failure is retried as cash.

Only use photon mode once the backend rollout and Trade/Finance support are
confirmed. Legacy backends may ignore `paymentType` and use cash. Creation
success is not proof of the channel: check the resulting list row and ledger.
No changes to image build/commit billing are implied by this flag.

## Failure diagnosis

After any create/access timeout or error, query recent failures before retrying:

```bash
bohr sandbox list --query failed -o json
```

For `sandbox_create_outcome_uncertain`, a sandbox may exist and be billable.
Retain the original request ID and any returned sandbox ID; inspect known IDs
and reconcile the list before another create. `retryable:false` and the
reconciliation hint take precedence over an upstream retry suggestion retained
as diagnostic evidence. A missing list row does not prove the POST had no effect.

Read the newest matching row's `failed_reason` and `status_reason` (`aborted-*`
is a failed create). Reasons are asynchronous; if blank, wait briefly and query
once more. After a create gateway timeout, also run `bohr sandbox list -o json`
to catch a late success before retrying.

`bohr sandbox create` POSTs to the sandbox workspace OpenAPI (v4) endpoint with
body `{"templateID": "<template-name>", "paymentType": 0}` by default.
`--project-id` is **optional**: with it, the value is sent as the
`X-PROJECT-ID` request header and the sandbox bills against that project's
cash budget. Without it, cash is the default; `--payment-source photon` explicitly
selects photons when the backend supports them. The response carries canonical
`sandbox_id` (the ID to use with `exec`, `delete`, etc.), retained compatibility
field `sandboxID`, `templateID`, `state`, `cpuCount`, `memoryMB`, and a
`metadata` dict populated by the platform.

**Sandbox lifetime.** When neither flag is set, the CLI sends `"timeout": 43200`
(12 hours) so multi-hour agent runs aren't cut off by the much shorter platform
default; the create command also prints a one-line stderr hint reminding you
the default is 12h and how to extend. Override with one of:

- `--timeout N` adds `"timeout": N` (seconds, non-negative integer) to the
  create body — the sandbox is auto-destroyed N seconds after creation.
  `--timeout 0` is the explicit "unlimited" sentinel and is passed through
  as-is, never substituted.
- `--never-timeout` adds the metadata flag
  `e2b.agents.kruise.io/never-timeout=true`, which requests disabling automatic timeout-based destruction; it is not a guarantee against other platform reclamation; no `timeout` field is sent. Pair with an explicit
  `bohr sandbox delete <sandbox_id>` when done — long-lived sandboxes keep
  billing.

**Extending a running sandbox.** To keep a sandbox alive past its deadline, run
`bohr sandbox connect <id> --timeout N`. The control plane measures `--timeout`
**from the moment you call connect**, not from creation (this matches E2B's
OpenAPI wording: "timeout in seconds from the current time after which the
sandbox should expire"), and only ever extends — the new deadline is
`max(current deadline, now + N)`, so a smaller `N` than the time still remaining
does not shrink it. Read the current deadline from `bohr sandbox describe <id>`
(`endAt`; the connect response itself does not carry `endAt`). To add N seconds
from now, pass `--timeout N` — do **not** pass "elapsed + the extra you want",
which over-extends and over-bills. This is the only client path to extend; there
is no separate renew endpoint.

`--timeout` and `--never-timeout` are mutually exclusive; passing both fails
the request before any HTTP call.

## Mount a Wenyon dataset in an instance

Use an ordinary compute template. Wenyon dataset, mount path, mode and starting
version are **sandbox create parameters**, never template defaults.

```bash
REQUEST_ID="$(uuidgen)"
bohr sandbox create <template> \
  --wenyon-dataset-id <dataset-id> \
  --wenyon-mount-path /mnt/wenyon \
  --wenyon-mode read --wenyon-version 3 \
  --wenyon-user-token-env WENYON_USER_JWT \
  --request-id "$REQUEST_ID" --yes -o json
```

`--wenyon-user-token-env` names an environment variable, not a JWT literal.
`BOHR_ACCESS_KEY` is not a Wenyon/Vouch user JWT. The named JWT is sent only in
`X-Wenyon-User-Token`; never put it in `--env`, request metadata or files.
`--dry-run` does not read the secret and redacts environment values. The CLI
checks server instance-mount capability before creating and requires the response
to acknowledge the dataset, path, mode, requested version and resolved version.
`--inherit-auth` remains incompatible until its rollback uses explicit discard.

Default mode is `read`; default version is `current`, pinned when the lease is
created. For editing, use `--wenyon-mode write`: current or a positive version
number preloads the complete baseline. Only `--wenyon-mode write
--wenyon-version none` starts empty. Missing versions and failed preload never
fall back to an empty directory. This version-selector contract supports
versioned datasets; manifest-based blob datasets are rejected.

The selected compute target must support Wenyon PVCs and have an operator-configured
storage mapping. Users do not supply cluster/PFS/PVC identifiers. Template image
readiness is separate from dataset readiness. Create waits for the lease to be
Ready before requesting a fresh sandbox with the mount.

On timeout or `sandbox_create_outcome_uncertain`, query the original creation:

```bash
bohr sandbox describe --create-request-id "$REQUEST_ID" -o json
bohr sandbox list --query failed -o json
```

Use the same profile as the create. A single lookup 404 or query failure does
not prove that no sandbox or lease was created. The CLI preserves the sanitized
original reason in `error.message` and the original message/subtype/hint in
`error.upstream`. Those are diagnostic evidence: follow the outer
`error.retryable=false` and reconciliation hint even if the upstream text says
to retry. An explicit 5xx reason alone does not establish a safe retry.
Human output also shows the upstream subtype and any known sandbox/lease IDs,
with credentials removed; the suggested action remains reconciliation.

Do not change the request ID and create another billable sandbox. Same-ID replay
returns the existing intent; changed parameters conflict. The create `--timeout`
is the sandbox lifetime, not the HTTP or data-preload waiting budget.

After success, inspect `data.wenyon_mount.resolved_version` and verify the expected
files/content, for example `findmnt -T /mnt/wenyon` followed by the probe hash.
An empty write workspace is only expected with explicit `none`.

Read mounts end without producing a version. Writable mounts require an explicit
choice; plain delete returns `WRITE_DISPOSITION_REQUIRED` without deleting:

```bash
bohr sandbox delete "$SANDBOX_ID" --commit --wait --force --yes -o json
# OR: discard instead of publishing the complete workspace
bohr sandbox delete "$SANDBOX_ID" --discard --wait --force --yes -o json
bohr sandbox describe "$SANDBOX_ID" -o json
```

`--commit` publishes the entire workspace, including deletions; an empty workspace
can publish an empty version. `--force` only skips the process pre-check and does
not choose commit/discard. Termination does not need the Wenyon JWT.
Without `--wait`, successful termination reports **accepted**, not saved.
`--wait-timeout 15m` bounds client waiting, not background settlement. `Settling`
and a preallocated version are not success; only `saved` with `produced_version`
confirms it. Discard cannot interrupt settlement already in progress.

On failed/expired sandbox execution, a writable workspace is not automatically
published. If still available, it enters `awaiting_decision` for one hour; inspect
`recovery_deadline` and choose commit/discard using the same delete command. It
holds the dataset write lock during recovery. After the deadline it is discarded.
Upstream failure or eviction can make the workspace unavailable earlier; report
that failure rather than promising recovery. Persistent mount results remain
queryable after sandbox deletion.

**Parent/child ownership.** Pass `--parent-sandbox-id <sandbox_id>` when a
driver sandbox creates a worker sandbox. The CLI sends top-level
`parentSandboxID`; it never hides ownership in arbitrary metadata. Use the
canonical sandbox ID returned by create/list (for example
`default--driver-abc12`), not the numeric database/list row ID. The control
plane validates that the parent exists, is accessible, and can own the child.
Omitting the flag preserves an independent root sandbox. This is lifecycle
ownership, not a display-only label: deleting a parent asks the service to
durably cascade cleanup through all recorded descendant generations.

**Storage mounts.** Pass `--mount-user-storage` to ask the platform to mount the
caller's personal disk and the share disks the caller has access to into the
sandbox (metadata flag `bohr.launching.io/mount-user-storage=true`). Default is
off — sandboxes start with no user storage mounted.

**Share disk sub-path.** Pass `--share-subpath <rel>` to mount only a
sub-directory inside the project share disk instead of its root. It only affects
the project **share** disk and only takes effect with `--mount-user-storage`
(+ a `--project-id` that grants a share disk). `<rel>` is confined server-side:
a leading `/` or any `..` segment is clamped back inside the share.

**Environment variables.** Pass `--env KEY=VALUE` (repeatable) to inject
environment variables at creation time. Splitting is on the first `=` only, so
values may contain `=`. The injected vars are visible to every later
`bohr sandbox exec` / `terminal` shell, e.g. `bohr sandbox exec <id> --command 'echo $HELLO'`.

**Opt-in bohr login inheritance.** `--inherit-auth` is default off and must be
used only with the caller's explicit authorization. It resolves the credential
the outer CLI would actually use, creates the sandbox, transfers a minimal
bundle through the data plane into verified `/dev/shm` tmpfs, and runs
an authenticated `sandbox list` read inside before reporting success. Before
writing secrets it polls for up to about 45 seconds so the background CLI distributor
can finish. Explicit inheritance trusts the sandbox and its installed execution
environment; injection runs as root and does not reject executable/package owners
or writable directories (including 0777). It resolves the npm wrapper to its nested
or hoisted Linux package, checks that Node and the native payload are executable,
and requires readable manifests with matching package identities and exact versions
at bohr-cli 2.6.45 or newer. These are installation/version compatibility checks.
Missing, older, or unsupported installations still fail before credential writes.
The authenticated read executes the native payload directly.
The normal create HTTP
header still authenticates the caller, but credential bytes are never added to
the create body/persisted `envs`, dry-run output, local temp files, or shell
argv. The isolated config contains only host/profile; the credential exists
only in its dedicated tmpfs file and becomes an environment variable after the
guard succeeds, so a failed guard cannot recover AccessKey from config.
AccessKey inheritance includes the effective key and host;
Vouch inheritance includes the current access token but **no refresh token**, so
it fails closed at expiry instead of risking the host session.

The CLI requires `/dev/shm` to be a root-owned tmpfs on a different device from
rootfs, accepts an initial 0777 or 1777, and normalizes
`/dev/shm` to sticky 1777 before writing. The private bundle stays root-owned
(0750 directory, 0640 files); PID 1's filesystem GID gets read/traverse but no
write permission. A root-owned,
secret-free guard in rootfs checks tmpfs ownership, modes, and a random marker
before sourcing the credential environment. The default root data-plane shell
is verified; if a provider forces PID 1 to root, the image's original non-root
user is not covered. Do not use this with an untrusted template. Both AccessKey
and Vouch inheritance require the verified npm bohr-cli 2.6.45+ layout. Older
releases are not run because they used an unsigned automatic updater, could put
AccessKeys in URLs, and lacked complete config isolation. Template images can differ by environment; the static package and authenticated-read checks catch
incompatibility after create. The pre-secret gate restores the fixed PATH after
login-profile startup so verification and subsequent commands resolve the same
installation; it does not require system directories or Node to be root-owned. `--inherit-auth` rejects `--no-wait`,
`--as`, and conflicting auth/debug/proxy/custom-CA/loader `--env` values; those ambient
variables are cleared too, so custom enterprise proxies/CAs are not supported
in this mode. `PATH` is fixed to standard system directories for consistent command resolution;
activate Conda/Mamba explicitly
afterward. It cannot satisfy an entrypoint that
needs bohr before post-create injection. If inheritance fails after a successful
create with a validated sandbox ID, the command attempts credential-file rollback
and automatically requests sandbox deletion with an independent 15-second budget,
including after caller cancellation. An unexpected Wenyon mount instead returns
`requires_settlement` and an explicit discard hint, preventing accidental write-lease
finalization by a plain DELETE. It returns non-retryable `PARTIAL_FAILURE`
and keeps the already-created sandbox ID. `data.auth_inheritance` records the phase,
allowlisted reason, credential-write/rollback facts, and sandbox cleanup state.
`credential_state=not_written` is established by the injection code, never by
matching remote error text. With explicit `--reserve-failed-sandbox`, only this
known pre-write state may be retained for debugging (`retained_for_debugging`);
inheritance still failed and billing continues. The flag cannot preserve a sandbox
when credential transmission may have started. `compatibility_check_failed`
indicates an installation, required-program, or version mismatch; executable
ownership or writable installation directories no longer cause rejection.
`delete_requested` means the DELETE was accepted; it does not confirm resource
removal or stopped billing. Use the control-plane verification hint. If cleanup
is `unconfirmed`, use the delete hint; do not inspect sandbox contents or reuse it.
A concurrent process may already have read credentials even after confirmed file
rollback. `sandbox_auth_rollback_uncertain` retains that stricter warning.
A lost/timeout/uncertain-5xx
create response returns `sandbox_create_outcome_uncertain`, `retryable=false`;
list and reconcile by request ID before considering another create. To keep the
irreversible operation's sandbox ID visible, create with `--inherit-auth`
ignores `--jq` and returns the full envelope; recovery hints preserve an
effective profile selected by the active config, `--profile`, or `BOHR_PROFILE`.
Non-default environment/active profiles must exist; inheritance fails before
create instead of silently falling back.
Hints never embed credentials or one-shot host overrides, so execute them in
the same authenticated host context when those came from environment variables.
The secret exists only in
the current runtime's tmpfs; a container recreation loses it and the guard
fails closed. There is no re-inject command in this first version.

**In-place image override.** Pass `--image <image-path>` to launch the chosen
template but replace its container image for this one sandbox. Template
selection is:

| Command shape | Selected resource template |
| --- | --- |
| `create --image <image-path>` | default CPU profile `1c2g`; currently `dynamic-image-replace` |
| `create --image <image-path> --cpu <profile>` | the profile's currently mapped public template |
| `create <template> --image <image-path>` | the explicit template's shape |
| `create --gpu [key] --image <image-path>` | the selected GPU template |

`--cpu` is a CLI profile key, not a SKU. It requires `--image` and is mutually
exclusive with an explicit template and `--gpu`; see the current
[CPU profile mapping](../platform-snapshot.md#image-override-cpu-profiles).
The metadata flag is `e2b.agents.kruise.io/image=<image-path>`; the platform
keeps the template's SKU / CPU / memory / GPU shape and only swaps the running
container image. Tell the user the selected profile/template/expected shape
before create; after success report actual `templateID`, `cpuCount`, and
`memoryMB` from the response.

Prefer an immutable tag or digest. Direct `create --image` accepts `:latest`,
but mutable or first-use uncached images can cold-pull and outlast the gateway
deadline. A timeout is not proof create failed: inspect both `sandbox list -o
json` and `sandbox list --query failed -o json` before retrying. Create still
defaults to 12 hours, personal-wallet billing, and no mounted user storage;
set those flags explicitly when needed. For a persistent image change, use
`bohr sandbox template update <name> --image <new-image>`. Find image
references with `bohr image ls`.

**Session id.** Pass `--session-id <id>` to tag the sandbox with a caller-defined
session id (metadata flag `bohr.launching.io/session-id=<id>`); the platform
uses it as the user-storage mount subPath, so sandboxes created with the same
session id share the same persisted directory (pairs with `--mount-user-storage`).

**Reserve a failed sandbox for debugging.** By default a sandbox that fails to
start is cleaned up. Pass `--reserve-failed-sandbox` to keep it around so you can
inspect why it failed (e.g. `describe` / logs) before deleting it yourself.

**Resolve the data-plane connection.** `bohr sandbox connect <id>` returns the
envd endpoint for a sandbox. The access token is redacted by default; pass
`--reveal-token` to include the sensitive envd access token in the output (only
when you actually need to hand it to another tool).

**`bohr sandbox list` — relation tree, age column and stale-sandbox warning.**
The default human view for active/running sandboxes groups descendants under
their parent. The historical `--query all` and diagnostic `--query failed`
views remain paginated flat lists; use the default active selection when a
complete live ownership tree is required. Use `--flat` for the legacy flat
active/running human table. `--quiet` and
structured formats (`-o json|table|yaml|ndjson|csv`) always remain flat for
script compatibility and preserve `parent_sandbox_id` / `relation_status`.
Rows whose linked parent is outside a filtered page say `parent not in result`;
only an authoritative backend `relation_status=orphaned` is labelled orphaned.
The human view also appends an `AGE` column showing how long each sandbox has been alive. Rows
older than 30 minutes are highlighted (yellow on a TTY, ⚠ glyph everywhere) and
a one-line footer reminds you to delete them if no longer in use. The footer is
suppressed under `--quiet` and on machine-readable output (`-o json`).
Structured output adds a `status_name` label and an `age_seconds` integer per
entry so consumers can apply their own highlight policy.

The default `--query active` preserves the legacy active-sandbox list.
`--query all|running|failed` uses the paginated status query. In particular,
`--query failed -o json` preserves the backend diagnostic fields for agents;
the human view shows the failure reason directly.

**Sandbox cost in list output.** Human output shows `COST` with a unit per row;
`-` means the cost is unavailable. Read `paymentType` alongside the amount:
`1` selects photons, `0` or missing selects CNY. Unknown payment types have an
unknown unit. Photon rows prefer `photonCost` and fall back to `cost`; cash
rows use `cost`. The CLI preserves raw JSON numbers and missing fields.

```bash
bohr sandbox list --query all --page-size 100 -o json \
  | jq '[.data.items[] | {sandbox_id,status_name,paymentType,photonCost,cost}]'
```

Follow `data.pagination.has_more`, incrementing `--page`, to cover historical
sandboxes. Missing or null amounts are unavailable, never zero. Explicit zero
is known for that response, but settlement can lag termination. Cash `cost`
is in CNY (yuan); photons are normally integers. A fractional legacy photon
`cost` is not rounded or multiplied by 100: prefer `photonCost` or verify the
ledger. `notice.billing.balance` is the account balance, not a sandbox cost.

**`describe` is read-only.** `bohr sandbox describe <id>` fetches metadata only
and does **not** wake a paused sandbox. Only `--with-processes` opens a live
connection (which wakes it) to list running processes.

**Pass the template `name`, not a SKU or numeric `id`.** Despite the request
field being called `templateID`, the backend accepts the template's string
`name` — the `name` column of `bohr sandbox template list`, or a platform
shortcut. List SKUs with `bohr sandbox machine list`.

For GPU workloads, use the `--gpu` shortcut (see [`../platform-snapshot.md`](../platform-snapshot.md)):

```bash
bohr sandbox create --gpu -o json
bohr sandbox exec <sandbox_id> nvidia-smi
bohr sandbox exec <sandbox_id> --command 'python train.py' -o json
bohr sandbox delete <sandbox_id> -o json
```

## Best practices

- **Quota** — `bohr sandbox quota -o json` before create; on `limit reached` errors run it again, free slots, then retry (never blind-retry).
- **Reuse** running sandboxes whenever possible — check `bohr sandbox list` before creating new ones.
- **Check before deleting** — run `bohr sandbox ps <sandbox_id>` (or `describe --with-processes`) to confirm no job is still running.
- **Save your work** before deleting — Sandbox-local data is permanently lost after delete. Use `bohr sandbox files read` to retrieve important files first. A Wenyon write mount follows its separate finalize contract described above. For the full retrieve-before-delete SOP for long-running jobs, see [`execution-modes.md`](./execution-modes.md).
- **Delete promptly** when done — release resources with `bohr sandbox delete <sandbox_id>` (alias `kill`). The following safety gates apply:
  - **Confirmation gate** (delete is a destructive op): on a TTY you get an interactive prompt; **non-interactively (agents, CI, piped shells) it is refused with `CONFIRMATION_REQUIRED` unless you pass `--yes` (or `--force`).**
  - **Running-process pre-check**: if the sandbox still has running processes (or they can't be verified), on a TTY they are listed for a `y/N` confirm; non-interactively the delete returns a structured `PRECONDITION_FAILED` (HTTP 412) with `bohr sandbox delete <id> --force` in `error.hint.command`.
  - Control-plane preflight failures retain their authoritative code even on a TTY: for example, an absent sandbox returns `RESOURCE_NOT_FOUND` rather than being downgraded to a warning or a plain Go error. These failures honor `-o json` / `-o yaml`.
  - `--yes` confirms the deletion but **keeps** the process pre-check — so a non-TTY delete of a *busy* sandbox with only `--yes` still fails, asking for `--force`.
  - `-f, --force` **both confirms and skips the process pre-check** (i.e. `--force` implies `--yes`), so `bohr sandbox delete <id> --force` can non-interactively delete even a busy sandbox. It does not alter service-side cleanup/finalization and must not be described as discarding Wenyon pending work.
  - **Parent cascade preview:** for an interactive non-`--force` delete, the CLI asks the service for the complete descendant set. If descendants exist it prints direct/total/running counts and asks a second `y/N` question. An incomplete preview is labelled explicitly; the service still cascades from its durable ownership relations after confirmation. Non-TTY and `--force` calls add no new prompt, but deleting a parent still performs the server-side cascade.


## Observe known deadlines

```bash
bohr sandbox describe <sandbox_id> -o json
bohr sandbox list --with-ttl -o json
bohr sandbox list --query running --with-ttl
```

Normal describe adds `data.lifetime`; list adds `data.items[].lifetime` only
with `--with-ttl`. Default list does not perform extra detail requests. Explicit
TTL listing reads each distinct ID on the current page, without connecting,
waking, renewing or automatically paging. It cannot be combined with `--quiet`.

Read `state`, UTC `end_at`, `observed_at`, `remaining_seconds`, and `reason`.
Known remaining seconds are floored and clamped to zero. A past deadline is
not evidence of deletion. Missing/invalid `endAt` is `unknown` with null deadline
and remaining seconds, even if never-timeout metadata is present. This is an
observation, not a reclamation SLA or proof that a later operation will finish.

On a detail request failure, listing stops and exits nonzero, preserving the
original items/pagination and completed observations in the failure envelope.
The failed ID is `lookup_failed`; later/unidentifiable rows are `not_queried`.
Do not interpret this partial result as a successful TTL check.
