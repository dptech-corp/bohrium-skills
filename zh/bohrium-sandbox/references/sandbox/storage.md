# Inspect and resize sandbox rootfs

The cluster must explicitly enable `capabilities.storage_management` for the
whole `bohr sandbox storage` group (`inspect` and `resize`). Servers enforcing
this declaration default it to disabled. `capabilities.storage` only controls
personal/shared disk mounting and does not enable this command group.
On `sandbox_storage_management_disabled` (legacy message code
`STORAGE_MANAGEMENT_DISABLED`), both commands fail without further probing or
polling. `retryable=false` and `hint.action=inspect_storage_in_sandbox`: enter
the target sandbox and use the read-only fallback below. Do not retry unchanged,
infer zero capacity, or automatically enable the capability. Enabled management still
does not guarantee resize support for every runtime or bypass quota checks.

```bash
# Run inside the target sandbox, adding the actual workload paths as needed.
df -h / /tmp
# Check inode availability if needed.
df -i / /tmp
```

The CLI does not execute this fallback automatically. Read `Avail` for each
filesystem and compare it with expected writes plus a safety margin. Paths on
the same filesystem share the same space; do not add their free space together.
For machine-readable sizing, `df -Pk / /tmp` reports 1024-byte blocks. If the
workload uses `/dev/shm`, inspect it separately; it is normally a memory-backed
filesystem. A failed read still means unknown capacity. These checks measure
usage only: they do not expand storage or establish resize support.

Before any capacity-sensitive operation, especially before proposing an image
commit, inspect the live writable rootfs:

```bash
bohr sandbox storage inspect <sandbox_id> -o json
```

The command combines control-plane rootfs values with read-only directory and
`df -Pk` probes for `/` and the selected remote temporary directory (`/tmp` by
default). Live probes can connect or wake the sandbox. Output intentionally
excludes mount relationships.

Read `rootfs.measurement.available_bytes` as current writable-root headroom;
`total_bytes`, `used_bytes`, and `used_percent` describe the same filesystem.
Check each path the workload will actually use. A `partial` report can retain a
valid rootfs measurement while a temporary-directory probe is unknown; use the
known measurement only for that path. Missing or `null` measurements are unknown:
never treat `null` as zero or as sufficient space.

`"status": "complete"` describes the evidence collected by this inspection, not
enough space for a task, support for resize, or completion of an expansion. It can
coexist with low space or unavailable resize. Human output calls it
`inspection evidence`. `ok=true` and exit code 0 only mean the query succeeded.

Before considering growth, read
`capabilities.dynamic_adjustment.available` from the same sandbox's inspection:

- `true`: the current provider API supports runtime rootfs resize; a request must
  still pass ownership, running-state and quota checks.
- `false`: do not call or retry `resize` for this sandbox. Read `reason_code`,
  `message` and `next_action` when present. If more space is required, stop that
  operation and choose an appropriate storage plan; do not automatically delete
  files, migrate data or create another sandbox.
- Missing, null or a non-boolean value: support is unknown. Do not infer support
  from the template name, cloud vendor or the generic `openkruise` provider name.

Older servers may only return `available` and `state`. The legacy state
`deferred_vendor_validation` does not prove physical storage can never expand;
`available=false` says the current platform API does not expose that operation.
Newer servers use `reason_code=sandbox_storage_resize_unavailable` and
`next_action=choose_storage_plan`; the same reason is the resize error subtype,
with `retryable=false`. Human output reports resize as available, unavailable or
unknown independently of inspection evidence.

When the inspected headroom is insufficient, `--target <N>Gi` is the total rootfs capacity after resize, not an increment. It never means "add N Gi".
Choose the smallest safe integer with:

`N = ceil((used_bytes + known_workspace_bytes + safety_margin_bytes) / 1GiB)`

Here `known_workspace_bytes` is the known rootfs space the pending operation
will add, and `safety_margin_bytes` covers uncertainty. If current
`available_bytes` already covers those two values, do not resize. Otherwise,
only after `capabilities.dynamic_adjustment.available` is explicitly true,
resize the running Pod, replacing `<N>` below with the calculated integer:

```bash
bohr sandbox storage resize <sandbox_id> --target <N>Gi --timeout 120 -o json
```

For example, `--target 50Gi` requests a 50 Gi rootfs after resize; it does not
add 50 Gi. Resize is grow-only and cannot shrink the rootfs. Never pass a
smaller target to reclaim space; an already-expanded rootfs stays expanded.
Values such as `50`, `50GB`, zero, and decimals are rejected.
`--timeout` is the CLI polling budget in seconds and defaults to 120.
The PATCH only initiates an idempotent expansion. The CLI then polls inspection
about every two seconds and fails closed unless all of these become true:

- `pod_uid` is unchanged from the PATCH response;
- the inspection state is `complete` and has no `warnings`;
- `pod_ready` is true;
- `rootfs.runtime_target_bytes` and `rootfs.runtime_actual_bytes` are at least
  the requested target;
- live `rootfs.measurement.total_bytes` from `df /` is at least 95% of the
  target, allowing filesystem metadata overhead.

After the live `df /` probe passes, the CLI performs one final control-plane
GET and rechecks the same UID and readiness gates before it can report Ready.

Only `"status": "ready"` is success. API/404 errors, a changed Pod UID, or a
timeout are failures and must stop the caller.

A satisfied repeated target can return `status=ready` and `patched=false`:
`patched` only says whether this request changed the target annotation. A timeout
stops the local wait, not an accepted backend expansion; inspect the same sandbox
before deciding what to do next.

Capacity fields have different sources:

- `control_plane.requested_bytes`: the Pod's declared ephemeral-storage request;
  it may remain 30 GiB after a successful expansion to 40 GiB.
- `control_plane.runtime_target_bytes`: the runtime expansion target; null can
  mean no target has been set, not a target of zero.
- `control_plane.runtime_actual_bytes`: the runtime-reported provisioned amount.
- `measurement.total_bytes` / `available_bytes`: filesystem measurements, including
  filesystem overhead. These are not the same as the configured request.

The legacy `diskSizeMB` on create/describe is not a live free-space reading. A
zero there does not establish that the sandbox has no disk; use storage inspection
for capacity decisions. `resize --target` is a total; template creation's
`--extra-ephemeral-storage-gb` is an increment above its base. Do not interchange them.

## Agent decision contract

1. Inspect the actual write paths before another capacity-sensitive operation;
   if the command group is disabled, use the sandbox-local `df` fallback above.
2. Compare known headroom with the workload's expected peak plus a safety margin;
   if sufficient, continue without resizing. Unknown workload size is not zero.
3. If insufficient, require explicit resize availability before calculating a
   target and calling `resize`; unavailable or unknown support requires a different plan.
4. Continue only when resize returns `"status": "ready"`; otherwise stop.
5. Re-inspect after later writes when the capacity decision may be stale.

For `bohr image commit`, use the proactive failure-data formula and
unknown-delta stop rule in [`../images.md`](../images.md). Inspect immediately
before commit and never wait for commit to cross a usage threshold: its apply
burst may reach ENOSPC before reactive expansion converges.

The CLI still does not calculate commit delta, preflight commit automatically,
or submit/retry commit from `storage resize`; the Skill is the Agent sequencing
contract.


## Low watermarks and temporary directories

```bash
bohr sandbox storage inspect <sandbox_id> --temp-dir /scratch -o json
```

`--temp-dir` selects an absolute directory inside the sandbox; it does not read
the caller's local TMPDIR or discover the workload's real scratch path. Select
the workload's actual directory explicitly. Missing, inaccessible or non-directory
targets are unknown; inspection never creates directories or changes permissions.

Read `rootfs.health` and `temporary_directory.health` independently:

- `low`: available bytes are strictly below 1 GiB, or available/total is strictly below 10%.
- `healthy`: neither low watermark is crossed; this is not a guarantee of enough space for the next operation.
- `unknown`: no valid measurement. The corresponding measurement is null, and the report is partial.

`temporary_directory.measurement` uses the same capacity fields as rootfs.
`low_watermark` reports the client thresholds; `advisories` identifies each low
scope and reason. A complete measurement may be low: advisories do not make it
partial. Conversely, a temporary-directory failure preserves a known rootfs
measurement and adds a scoped warning. Capacities must not be added together:
the two paths may share the same underlying filesystem.

This inspection only reports. It does not clean files, reserve space, block
commands or resize storage. Low capacity alone exits successfully with advisories.
A control-plane request failure exits nonzero; an unavailable live probe returns
a partial report, as before. Cancellation stops remaining probes. Resize keeps
its existing rootfs-only capacity and readiness checks.

Human output labels `healthy` as a clear low watermark with task capacity not
assessed. For example, 8 GiB free can clear the fixed thresholds but still be
insufficient for a workload expected to write 28 GiB. The JSON health values are
unchanged for compatibility. Live probes can extend a near-expiry sandbox lifetime
through connect; they are not a lifecycle-neutral metadata-only query.
