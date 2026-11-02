---
name: bohrium-sandbox
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Operate Bohrium cloud sandboxes with the `bohr sandbox` CLI: sandbox
  lifecycle, quota self-query, command execution, interactive PTY sessions,
  file transfer, per-sandbox cost inspection, templates, container image
  commit/build, Wenyon dataset mounts, and installing software inside sandboxes. Use when the user
  mentions bohr sandbox, sdbx, a cloud sandbox, sandbox cost or spend, quota,
  or any Bohrium sandbox operation; also when sandbox access times out or
  fails, or create reports "running sandbox limit reached" / "template limit
  reached".
---

# Bohrium Sandbox (`bohr sandbox`)

CLI for managing **on-demand cloud sandboxes** on the Bohrium platform via the
`bohr sandbox` command group (open-platform v4). Covers sandbox lifecycle
(`create`, `exec`, `delete`, `list`, `describe`), quota self-query (`quota`),
interactive PTY sessions, rootfs inspection and resize, file transfer, template management, image
commit/build, and software installation inside sandboxes.

This is **not** the E2B SDK: it does not need `E2B_API_KEY` and does not call
`api.e2b.dev`. Everything runs on the Bohrium platform, authenticated with a
Bohrium AccessKey. Creating a Sandbox with a Wenyon instance mount additionally
needs a Wenyon/Vouch user JWT from a named environment variable; the AccessKey
cannot substitute for that JWT.

> **Proxy toggle.** Sandboxes default to no outbound HTTP proxy. For overseas
> access (PyPI `pypi.org`, GitHub, HuggingFace, etc.), toggle the
> `proxy.example.com` proxy on with the snippet in
> [`references/sandbox/network.md`](references/sandbox/network.md), then toggle
> it off when done. Domestic access via the image-level Aliyun mirror is
> unaffected.

> **Default-behavior pitfalls (read first for ML / long-running jobs).**
> Several defaults are fine for quick interactive use but bite training
> workloads: the overlay disk has a **30Gi base** and any extra must be
> requested when creating the template (SKU has no disk field),
> per-user **running-sandbox quota** (Running+Hibernating — check
> `bohr sandbox quota -o json` before create; on `limit reached` free slots
> then retry), `exec` has no execution deadline (network failures and sandbox lifetime still apply), `files read` of a
> binary is safe only via byte mode, `delete` is **irreversible** and a
> stopped sandbox's files can't be read, the outbound proxy is **off**, and a
> template image with a mutable `:latest` tag is **rejected** (use an
> immutable tag). See the consolidated list in
> [`references/sandbox/pitfalls.md`](references/sandbox/pitfalls.md) before
> driving a sandbox through an agent.

## Agent rules

- On an uncertain create or a lost response, do not rerun it blindly: query the original request ID with `bohr sandbox describe --create-request-id`, preserving the profile and identity. A single lookup 404 does not establish that no sandbox was created. For template preparation failures, inspect template status; only inspect a failed-instance record when an instance is actually identified.
- Pass `--wenyon-dataset-id`, `--wenyon-mount-path`, `--wenyon-mode` and `--wenyon-version` to **sandbox create**, using an ordinary compute template. Default read/current pins existing data; write/current preloads it; only write/none starts empty. `--wenyon-user-token-env WENYON_USER_JWT` names an environment variable, not a Wenyon/Vouch user JWT literal; `BOHR_ACCESS_KEY` cannot substitute. Do not combine with `--inherit-auth`. End writable mounts with `delete --commit` or `delete --discard`; use `--wait` and verify `produced_version` for a saved result. On uncertain create, use `describe --create-request-id` before any retry. See lifecycle.md for the recovery deadline and no-secret rules.
- Pass `sandbox create --inherit-auth` only when the user explicitly authorizes propagating the caller's login into that sandbox. It is default off. Root and any process granted read access to the private tmpfs bundle can use the inherited credential; both AccessKey and Vouch require a statically verified npm bohr-cli 2.6.45+ layout. Vouch never includes a refresh token and expires with its access token. If injection returns `PARTIAL_FAILURE`, delete the returned sandbox without inspecting/reusing it and do not rerun create; subtype `sandbox_auth_rollback_uncertain` makes deletion immediate. For `sandbox_create_outcome_uncertain`, list and reconcile by request ID before another create.
- If `template create --shm-size` reports that the server did not echo the requested `shm_size`, the template may already exist. Do not retry blindly; inspect the template first, then delete or recreate it only after the backend is upgraded or corrected.
- Treat `/share` and `/personal` as potentially network-backed. Do not repeatedly run recursive `find`, `du`, `grep -R`, or content searches over them. For repeated filename lookup, build one private, path-only local index for the narrowest useful subtree, refresh it on writes/mount changes or after 10 minutes, and verify live paths before mutations. See [`references/sandbox/remote-storage-index.md`](references/sandbox/remote-storage-index.md).
- To inspect per-sandbox spend, `bohr sandbox list --query all` shows `COST` with a unit per row. Read `paymentType`: `1` is photons, `0` or missing is CNY. Photon rows prefer `photonCost`, falling back to `cost`; cash rows use `cost`. Missing `cost` means unavailable unless a photon row has `photonCost`; never invent zero. An explicit zero is known at query time. JSON preserves both fields and their absence. These are job costs, not `notice.billing.balance`; settlement can lag lifecycle changes.
- Commands and new PTYs default to `root`. Use `--user` only when explicitly requested; an existing PTY keeps its original user.
- Positional `exec` arguments preserve argv boundaries. Pass shell pipelines, expansion and redirection through `-c/--command`; use `--` before remote argv to stop local flag parsing. See exec.md for outcome-unknown recovery and retained scripts.
- Before every non-Dockerfile snapshot through canonical `bohr image commit`, run `bohr sandbox storage inspect <sandbox_id> -o json` immediately before submit. If the command group is disabled, use sandbox-local `df -Pk` on the actual write paths as described in the storage guide; this measures capacity but does not enable resize. Commit apply can consume rootfs in a short burst, so never start commit and wait for an 80% threshold to trigger expansion. Estimate total capacity from live `used_bytes` plus a defensible uncompressed commit delta; unknown delta is not zero. If growth is needed, require `capabilities.dynamic_adjustment.available=true` before resize; false or unknown needs another storage plan, not automatic retries. Proceed only after resize `"status": "ready"`; inspect `complete` only describes evidence, and a PATCH alone is not Ready. Resize is grow-only and cannot shrink a rootfs. This is an Agent workflow, not automatic CLI preflight, and does not apply to `bohr image build --dockerfile`. Read the formula and stop conditions in [`references/images.md`](references/images.md) and Ready evidence in [`references/sandbox/storage.md`](references/sandbox/storage.md).

## Prerequisites

```bash
export BOHR_ACCESS_KEY=<your-access-key>
bohr sandbox doctor -o json   # verify local auth and control-plane connectivity
```

For an existing sandbox and an authorized HTTP(S) target, add
`--sandbox <id> --target <url>` to test that data-plane path. This connects
the instance and may extend its lifetime. See lifecycle.md for layer and
submission-stage semantics.

The verdict is `data.ready`, not the exit code: `sandbox doctor` exits 0
whenever the diagnosis itself ran, ready or not. Check `data.ready` (and
`data.checks[]` for which item failed) before creating anything.

## Billable commands need `--yes` when you are not a human at a terminal

`sandbox create`, `image build` and `image commit` create billable resources.
Since 2.5.18 the CLI refuses them outside an interactive terminal unless you
confirm — and an agent's piped stdin is never a terminal, so this is your
default situation:

```json
{"ok": false, "error": {"code": "CONFIRMATION_REQUIRED", "subtype": "billing_resource",
 "hint": {"action": "confirm", "command": "bohr sandbox create --yes"}}}
```

Exit code is **10** and nothing ran — no sandbox, no charge. This is not a
transient failure; re-running the identical command gets the identical refusal.

- Authorized to spend? Re-run the same command with `--yes`.
- Not authorized? Report the refusal to the user and stop. Do not add `--yes`
  on your own initiative.
- Keep confirmation enabled. Change persistent billing confirmation only when
  the user explicitly requests that configuration change; it is not a workaround
  for a refused operation. Destructive operations keep their own confirmation.
- `--dry-run` is never gated. Check prices with `bohr billing pricing`.
- Before creating, preview the cost with `bohr sandbox create <template> --estimate-cost -o json`: it reads the template's SKU rate into the dry-run note (read-only, two extra lookups). The note includes the per-hour rate and an approximate total for the effective lifetime (12 hours by default); `--timeout 0` and `--never-timeout` have no bounded total. Failed price lookups explain why an estimate is unavailable. With `--payment-source photon` it is still the cash rate, not your photon charge.

## Default sandbox templates

`bohr sandbox create` with no template argument and no `--image`/`--gpu`
creates a sandbox from the `sdbxagent` template, which ships the image-level
Aliyun PyPI mirror, so domestic `pip install` is fast out of the box. The
outbound HTTP proxy is
**off by default**; for overseas access toggle it on demand via the
`proxy on` / `proxy off` snippets in
[`references/sandbox/network.md`](references/sandbox/network.md).

For an image override, omit the positional template and optionally pass
`--cpu <profile>`. Omitting `--cpu` selects `1c2g`, currently backed by
`dynamic-image-replace` (1 CPU / 2 GiB, no GPU). `--cpu` requires `--image`
and is mutually exclusive with a positional/`--template` selector and
`--gpu`. The profile selects CPU/memory resources; `--image` still replaces
only the container image. Before create, report the selected profile/current
template/expected shape; after success, report the response's actual
`templateID`, `cpuCount`, and `memoryMB`. See the current
[CPU profile mapping](references/platform-snapshot.md#image-override-cpu-profiles).

For GPU work, pass `--gpu` as a shortcut: bare `--gpu` resolves to the default
GPU template (`4090`); `--gpu 5090` and `--gpu l20` pick the other two. It is
mutually exclusive with the positional `template` argument and `--cpu`.

## Mount a Wenyon dataset in an instance

Use an ordinary compute template. The dataset, mount path, mode, and starting
version are **sandbox create parameters**, not template defaults. Check
`bohr sandbox create --help` and the local
[lifecycle reference](references/sandbox/lifecycle.md#mount-a-wenyon-dataset-in-an-instance).

```bash
bohr sandbox create sdbxagent \
  --wenyon-dataset-id <dataset-id> \
  --wenyon-mount-path /mnt/wenyon \
  --wenyon-mode read --wenyon-version current \
  --wenyon-user-token-env WENYON_USER_JWT \
  --request-id "$REQUEST_ID" --yes -o json
```

`read/current` is the default and pins the current version when the lease is
created. `write/current` or a positive version preloads that complete baseline;
only `write/none` starts empty. Missing versions and failed preload do not fall
back to an empty directory. Confirm the dataset, mount path, requested/resolved
version, and actual access mode in the response and verify the mounted data.

`--wenyon-user-token-env` names an existing environment variable holding a
Wenyon/Vouch user JWT; it never takes a token literal. AccessKey authentication
and `--as user` do not supply this JWT. Do not copy `bohr auth token` output when
it reports `auth_method=access_key`. Keep tokens out of command arguments,
metadata, dry-run output, logs, and chat. Do not combine this mount with
`--inherit-auth`.

A writable mount requires an explicit settlement choice: `sandbox delete
--commit --wait` publishes the entire workspace, including deletions;
`--discard --wait` discards it. `--force` only skips the process check and does
not choose settlement. Check the returned `produced_version` after commit.
Without `--wait`, acceptance is not evidence that the data was saved.

If writable execution fails or expires and the workspace remains available,
inspect `recovery_deadline` and choose settlement during the one-hour
`awaiting_decision` window. The write lock remains held; after the deadline the
workspace is discarded. Upstream eviction can make recovery unavailable sooner.
Keep the same request ID while reconciling an uncertain create instead of
creating another instance. A sandbox lifetime limit is not an HTTP or mount
readiness timeout, and an image `READY` state does not prove dataset readiness.

## Command surface

| Command | Purpose |
| --- | --- |
| `bohr sandbox` | Operate cloud sandboxes |
| `bohr image` | Catalog images; commit / build sandbox images |

## Sandbox commands

| Subcommand | Description |
| --- | --- |
| `doctor` | Inspect sandbox configuration and readiness |
| `quota` | Caps + usage (before create / after `limit reached`) |
| `storage inspect <id>` | Read-only rootfs capacity, usage, and available space |
| `storage resize <id> --target <N>Gi` | Expand a running sandbox rootfs and wait for verified Ready |
| `create [template]` (`run`) | Launch a new sandbox (default: `sdbxagent` template); `--parent-sandbox-id` records ownership by another sandbox |
| `list` (`ls`) | List active sandboxes as a parent/child tree in human output (`--flat` opts out); structured/quiet output stays flat; `--query failed` diagnoses recent failures |
| `describe <id>` (`get`) | Read-only sandbox details (does not wake a paused sandbox) |
| `connect <id>` | Resolve the data-plane connection for a sandbox |
| `exec <id> <cmd>` | Run a command as root inside a sandbox (foreground / background) |
| `ps <id> [--pid <pid>]` | List processes, or inspect one exact numeric PID with the full existing projection |
| `process connect <id> <pid>` | Observe only output received on this exact-PID connection stream |
| `process kill <id> <pid>` | After exact-PID resolution and confirmation, request SIGKILL and report RPC acceptance only |
| `delete <id>` (`kill`) | Terminate a sandbox; deleting a parent also cascades its recorded descendants |
| `files write --source <src> <id> <path>` | Upload a file or directory |
| `files read <id> <path> [--destination <dst>]` | Read or download a file |
| `terminal attach <id>` | Attach a local TTY to a PTY (interactive, `docker exec -ti` style) |
| `terminal create <id>` | Open an interactive PTY session as root |
| `terminal send <id> <pid> <input>` | Send input to a PTY session |
| `terminal kill <id> <pid>` | Terminate a PTY session |
| `machine list` (`machine ls`) | List available sandbox machine SKUs |
| `image commit` | Snapshot a running sandbox (`--wait` for terminal status) |
| `image build` | Build an image from a Dockerfile (`--wait` for terminal status) |
| `image build-log <id>` | Fetch or follow build logs |
| `image dockerfile <id>` | Print the stored Dockerfile source; `-o json` preserves raw base64 |
| `image get <id>` | Alias of `bohr image get` (catalog) |
| `image list` (`image ls`) | Alias of `bohr image list` (catalog) |
| `template create` | Create a new sandbox template (`--visibility 1`=public / `2`=private, default private); legacy Wenyon descriptor flags do not configure instance mounts |
| `template list` (`template ls`) | List available templates |
| `template update` | Update a template's image |
| `template delete` (`template rm`) | Delete a template |
| `template share` | Share a private template with one user or project principal |
| `template unshare` | Revoke one user or project principal grant |
| `template grants` | List a private template's user and project grants |

Process selectors are exact positive uint32 PIDs only. `process connect` promises
only output received after its stream attaches: do not claim historical replay
or completeness. An observed end event propagates its exit code; a missing end
event is `sandbox_process_outcome_unknown`, `retryable=false`. `process kill`
uses SIGKILL after a List preflight and deferred confirmation; `signal_accepted`
means only that SendSignal was accepted, and `termination_confirmed` is always
false. A transport failure after the signal request is outcome-unknown and must
not be blindly retried. PID reuse between List and the operation cannot be ruled
out, and the data-plane credential can target any PID visible in that sandbox.
Recheck with `ps --pid`; use `terminal attach` or `terminal kill` for PTY
sessions and the manual TERM flow for workloads that need graceful shutdown.

## Discovering commands

```bash
bohr sandbox --help
bohr sandbox <subcommand> --help
```

## Version contract

This Skill is embedded in the `bohr` binary and uses the same version contract
as the commands it documents. Read the bundled copy directly:

```bash
bohr guide read bohr-sandbox
bohr guide list bohr-sandbox/references
```

`bohr guide` does not fetch from the Bohrium Skill Hub. Upgrade `bohr` to get
guidance for a newer command surface.

## Reference docs

Deep usage docs under `references/`. Load only what's relevant.

**Sandbox lifecycle & usage:**

- [`references/sandbox/pitfalls.md`](references/sandbox/pitfalls.md) — default-behavior traps (30Gi base disk and create-time extra, running-sandbox quota, timeouts, Wenyon readiness/JWT, binary downloads, delete, proxy, billing) + an ML-training checklist
- [`references/sandbox/lifecycle.md`](references/sandbox/lifecycle.md) — create / list / describe / delete / quota, plus best practices
- [`references/sandbox/execution-modes.md`](references/sandbox/execution-modes.md) — foreground vs background vs PTY, plus the retrieve-before-delete SOP for long-running jobs
- [`references/sandbox/exec.md`](references/sandbox/exec.md) — `bohr sandbox exec` usage, including `--background`
- [`references/sandbox/terminal.md`](references/sandbox/terminal.md) — PTY sessions for REPLs / TUIs / Ctrl-C interaction
- [`references/sandbox/files.md`](references/sandbox/files.md) — upload / download files and directories
- [`references/sandbox/remote-storage-index.md`](references/sandbox/remote-storage-index.md) — avoid repeated remote scans with a private, expiring path index
- [`references/sandbox/storage.md`](references/sandbox/storage.md) — rootfs evidence, runtime resize, and the pre-image-commit agent rule
- [`references/sandbox/templates.md`](references/sandbox/templates.md) — list, create, update, and delete sandbox templates, including Wenyon dataset descriptors
- [`references/sandbox/network.md`](references/sandbox/network.md) — on-demand `proxy.example.com` HTTP proxy toggle

**Container images (sandbox snapshots & builds):**

- [`references/images.md`](references/images.md) — commit/build images, inspect status/logs, and recover stored Dockerfiles

**Software install (inside sandboxes):**

- [`references/software-install/README.md`](references/software-install/README.md) — pre-installed inventory + domestic mirrors for sandbox installs

**Platform snapshot:**

- [`references/platform-snapshot.md`](references/platform-snapshot.md) — pinned CPU/GPU template shortcuts and SKU discovery

## Output formats

`bohr sandbox` commands accept `-o json|human|table|ndjson|csv|yaml`. For scripting and
agent use pass `-o json`, optionally with `-q <jq expression>` to filter.

## Resource payment boundary

`sandbox create --payment-source photon` selects photon order billing;
`--payment-source balance` selects balance. Omission keeps balance as the default.
Personal photon billing needs no `--project-id`: the caller is identified by AK.
Project budgets cannot use photons: omit `--project-id` or choose balance.
Invalid sources and conflicting selectors fail before confirmation or creation.
`--yes` is still required for non-interactive billable creation. A rejected
creation never retries with another payment channel.

The CLI maps the choice to JSON `paymentType: 1` (photon) or `0` (balance),
without sending the gateway payment-source header for this order operation.
The legacy `--payment-type 0|1` remains compatible. When both flags are given,
they must select the same channel. Image build/commit retain balance-only behavior.

Photon creation requires the target environment and account to route to a
photon-capable sandbox backend with Trade/Finance support. A legacy backend
can ignore the field and charge cash: do not use photon mode before backend
support and rollout are confirmed. A successful creation response alone does
not prove the payment channel or debit; inspect the sandbox list and billing
ledger. Fractional legacy photon `cost` values are preserved without scaling
or rounding; use `photonCost` or the ledger to confirm consumption.
