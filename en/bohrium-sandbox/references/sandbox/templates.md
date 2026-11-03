# Sandbox templates

List, create, update, and delete sandbox templates.

List templates you've already created, or create a new one.

```bash
bohr sandbox template ls              # table of your templates
bohr sandbox template ls -o json      # machine-readable JSON
bohr sandbox template ls --project-id 29496 # include templates granted to this project
bohr sandbox template ls --page 2 --page-size 20 -o json
bohr sandbox template ls --quiet      # names only (pipe into `bohr sandbox create`)
```

Show a single template by exact name:

```bash
bohr sandbox template get <name>          # human-readable summary incl. the network policy
bohr sandbox template get <name> -o json  # bounded network summary plus other template fields
```

`get` uses the backend's exact lookup endpoint (`/templates/lookup?name=`) and
returns `NOT_FOUND` (non-zero exit) when no template has that exact name.


`-o json` always gives the same shape, paginated or not: `data.items` holds the
rows and `data.pagination` holds `page`, `page_size`, `total` and `has_more`.
Passing `--page` or `--page-size` switches to the paginated backend endpoint but
does not change the shape. When paginating, the default page size is 100.

Creating a template requires an image reference and a SKU name:

```bash
bohr image ls                     # find an image path (paste into --image)
bohr sandbox machine list         # find a SKU name
bohr sandbox template create <name> \
    --image <image-path> \
    --sku-name <sku>
```

Additional `template create` flags:

- `--visibility <1|2>` sets who can see and use the template — `1` = public
  (visible to everyone), `2` = private (only you, unless explicitly shared).
  See [Visibility](#visibility) below.
- `--project-id <id>` binds the template to a project's shared storage.
- `--provider <name>` selects a specific compute provider. It is whitelist-gated
  — leave it unset unless the platform told you to use a specific provider.
  The exception is a Wenyon template, which must explicitly pass the exact
  value `--provider volcengine-openkruise`.
- `--wenyon-dataset-id <id>` and `--wenyon-mount-path <absolute-path>` opt the
  template into Wenyon mounting. They must be supplied together.
- `--extra-ephemeral-storage-gb <N>` adds quota-gated GiB above the 30Gi base
  overlay disk. This is fixed at template-create time.
- `--shm-size <quantity>` sets `/dev/shm` at template-create time using a
  positive Kubernetes quantity such as `8Gi`.
- `--network-allow <entry>` sets template allow entries. Repeatable, with
  comma-separated lists accepted. No mode flag is needed. Capacity is enforced
  by the backend; see [Network policy](#network-policy).
- `-f, --file <req.json>` supplies the create request as a JSON file; any
  explicit flags you also pass override the matching fields from the file.

## Visibility

Every template has a **visibility** that controls who can see and launch
sandboxes from it. Only two values are accepted at creation:

| Value | Meaning | Who can use it |
| --- | --- | --- |
| `1` | **public** | Everyone on the platform |
| `2` | **private** | Only you (unless explicitly shared) |

```bash
bohr sandbox template create shared-tpl \
    --image <image-path> --sku-name <sku> \
    --visibility 1                       # publish to everyone
bohr sandbox template create my-tpl \
    --image <image-path> --sku-name <sku> \
    --visibility 2                       # private (same as omitting the flag)
```

Rules an agent should know:

- **Default is private.** Omitting `--visibility` (or the `visibility` field in
  `--file`) creates a **private** (`2`) template — the backend defaults an unset
  value to private, so nothing is published by accident.
- **Only `1` or `2` are valid.** The CLI rejects any other value before the
  request is sent.
- Extra overlay storage is quota-gated independently of visibility; public and
  private templates use the same effective extra-storage quota.
- Prefer **private** unless the user explicitly asks to publish/share a template
  with others.

## Historical Wenyon dataset template

Historical API reference only. New sandbox instance mounts never read these
descriptors. Use the ordinary compute template and the sandbox create parameters
in [lifecycle.md](./lifecycle.md).

A Wenyon template stores a dataset descriptor for later Sandbox creates. It
does not create a dataset lease or PVC and does not need a Wenyon user JWT.
Create it with all three Wenyon-specific inputs visible to the caller:

```bash
bohr sandbox template create wenyon-work \
  --image <immutable-image> \
  --sku-name <sku> \
  --provider volcengine-openkruise \
  --wenyon-dataset-id <dataset-id> \
  --wenyon-mount-path /mnt/wenyon \
  --visibility 2 \
  -o json
```

The current create response does not acknowledge the Wenyon descriptor or
Provider binding. Therefore dry-run proves only the local request, and the CLI
cannot distinguish an older server that ignored unknown fields. Use a known
supported environment, keep the warning visible, and verify the configured
mount on the first Sandbox; never infer support from HTTP 2xx or a template
name alone.

The current contract is deliberately strict:

- `--provider` is required and must equal `volcengine-openkruise` exactly
  (case-sensitive, after trimming). Missing it fails locally with subtype
  `wenyon_provider_required`; any other value fails with
  `wenyon_provider_invalid`. The CLI does not guess, correct, or silently
  replace the Provider.
- `--wenyon-dataset-id` and `--wenyon-mount-path` are a pair and must both be
  non-empty. The mount path must be a clean absolute POSIX path, cannot be `/`,
  and cannot conflict with another volume mount path.
- The template name must be non-empty and at most 31 characters. The image
  should use an immutable tag or digest, and `--sku-name` must be supported by
  this Provider.
- A Wenyon template is forced private (`visibility=2`) and uses `replicas=0`.
  Public visibility or non-zero replicas are rejected rather than silently
  rewritten. It cannot be published, shared through `template share`, or used
  to keep prewarmed Sandbox instances.
- Template creation never reads `WENYON_USER_JWT` and must not receive
  `--wenyon-user-token-env`; the Wenyon/Vouch user JWT is needed only when a
  Sandbox is created from this template.

The same descriptor can come from `--file`:

```json
{
  "name": "wenyon-work",
  "image": "registry.example.com/team/workload:v1",
  "sku_name": "<sku>",
  "provider": "volcengine-openkruise",
  "visibility": 2,
  "replicas": 0,
  "wenyon_mount": {
    "dataset_id": "<dataset-id>",
    "mount_path": "/mnt/wenyon"
  }
}
```

Explicit flags retain the normal precedence over matching file fields, but the
final merged request must satisfy every invariant above.

## Publication status: create/update succeeded ≠ template Ready

`template create` and `template update` are asynchronous: a 2xx means the
request was **accepted**, while the desired revision's publication (target
selection, image preparation, activation) completes in the background. The
success output names the query entry point; do not treat acceptance as
readiness and do not re-issue the create to "push it along".

```bash
bohr sandbox template status <name> -o json            # one-shot status read
bohr sandbox template status <name> --watch            # block until a terminal state
bohr sandbox template status <name> --watch --interval 10s --timeout 10m

bohr sandbox template create <name> --image <image> --sku-name <sku> --wait
bohr sandbox template update <name> --image <image> --wait
```

Reading the result:

- `publication.state` is the verdict for **this** create/update (the desired
  revision): `READY` / `FAILED` are terminal; `PREPARING` / `ACTIVATING` /
  `RETRY_WAIT` / `UNKNOWN` keep going. `ready` also carries
  `phase`/`reason`/`message`/`retryable`/`action` and per-target detail —
  preserve them instead of re-classifying.
- `available` and `preparation_status` describe the **currently active**
  revision. An old revision staying Ready while the new one fails means
  `available=true` **and** `publication.state=FAILED`: older sandboxes keep
  working, the update failed. Never read the active revision's READY as this
  publication's success.
- Older services have no `publication` object: the wait still does not trust
  `preparation_status=READY` when the desired revision differs from the active
  one, and a legacy `FAILED` stops the wait with whatever reason is available.
- Exit semantics under `--watch`/`--wait`: READY → exit 0; FAILED →
  `TEMPLATE_PUBLICATION_FAILED`, exit 1, with the full sanitized diagnostic in
  `data`; timeout → `TEMPLATE_WAIT_TIMEOUT`, retryable — that is the **client**
  stopping, the server-side publication continues; cancellation →
  `TEMPLATE_WAIT_CANCELLED`. `--wait` only polls the read endpoint; it never
  repeats the POST/PATCH.
- **The wait pins a revision.** `--wait` pins the revision the accepted
  response named; a bare `--watch` pins the first desired revision it
  observes. If the observed revision differs from the pinned one — stale
  snapshot and a different update look identical — or a READY/FAILED comes
  without revision identity, the wait exits 1 with
  `TEMPLATE_PUBLICATION_UNCONFIRMED`: neither success nor failure, and the
  message names both revision IDs. If the accepted response itself carries no
  revision identity, `--wait` reports unconfirmed immediately without polling
  and explicitly says the write must not be re-issued; the only follow-up is
  the read-only status query. Query the status and reconcile the two
  revisions before acting.
- Machine output carries the acceptance fact as envelope warning
  `TEMPLATE_PUBLICATION_PENDING` (with the status query command), not only as
  a stderr line.
- When `bohr sandbox create` fails because the template has not published
  (e.g. `TEMPLATE_PUBLISHING` or subtype `template_publication_failed`), the
  error hint points at `bohr sandbox template status <name>`, not at the
  failed-sandbox list — no sandbox exists yet at that stage.

Image preparation and Wenyon data readiness are independent. Template image
states `PREPARING`, `READY`, and `DEGRADED` describe only the image cache:
`PREPARING` means wait for image preparation, `READY` is the most reliable
image path, and `DEGRADED` may cold-pull from the registry. None of those
states means the Wenyon dataset or PVC is Ready; the per-Sandbox PVC is created
only during Sandbox creation.

The descriptor requests a **write** lease. Its mount starts as a new empty
output directory and does not copy or seed the dataset's current version.
Wenyon prewarm/read-open only serves read/warm replicas; it never fills this
write mount. The current `bohr sandbox` Wenyon flow does not support a read
lease. If existing content must be preserved, arrange a separate Wenyon read
flow and explicitly copy/write that content into the write mount before the
Sandbox reaches a terminal state.

## Share a private template

Grant access to exactly one typed principal at a time:

```bash
bohr sandbox template share <name> --user-id 26745
bohr sandbox template share <name> --project-id 29496
bohr sandbox template grants <name>
bohr sandbox template grants <name> -o json

bohr sandbox template unshare <name> --user-id 26745
bohr sandbox template unshare <name> --project-id 29496
```

`--user-id` and `--project-id` are mutually exclusive and IDs must be positive.
A Wenyon template cannot be shared; the commands in this section apply only to
ordinary private templates.
A project grant stores the project as the principal; it does not copy or expand
the project's current member list. Both discovery and creation must carry the
same project context:

```bash
bohr sandbox template list --project-id <id>
bohr sandbox create <name> --project-id <id>
```

The list flag only sends `X-Project-Id` for project membership/access checks;
it does not modify or bind the template. On sandbox create, `--project-id`
retains its existing billing and project-share-storage context. The service
re-checks the caller's current project membership. Removing a member therefore
removes their access without rewriting the template grant.
Only the template owner/authorized manager can list or change grants; backend
404/403/5xx responses remain failures rather than being treated as success.

> **`:latest` image tags are rejected.** `template create` and
> `template update --image` refuse any image whose tag ends in `latest` (e.g.
> `:latest`, `:dev-latest`) and ask for an immutable tag. Why: the backend
> pre-warms an image cache keyed by the image tag, and a mutable `:latest`
> cannot be reused — it is deleted-and-rebuilt on every create/update pointing
> at it, which churns the cloud image-cache quota and makes affected templates'
> new sandboxes cold-pull the image (slow — can hit the create timeout). Use an
> **immutable / unique tag**: a version, or the unique tag that `image build` /
> `image commit` already produces.

## Disk size (extra ephemeral storage)

A sandbox's overlay (writable root) disk defaults to **30Gi**. The SKU only
sizes CPU/memory/GPU — it does **not** include disk. If models, datasets, or
Python envs fill the default disk you'll hit `OSError: No space left on device`
(e.g. a 7B + 1.5B model plus torch/transformers easily exceeds 30Gi).

Disk size is fixed per template at creation time (it cannot be changed on
`bohr sandbox create`). Request extra overlay storage with
`--extra-ephemeral-storage-gb` when creating the template:

```bash
bohr sandbox template create <name> \
    --image <image-path> \
    --sku-name <sku> \
    --extra-ephemeral-storage-gb 50      # 30Gi default + 50Gi extra = 80Gi total
```

Rules:

- **Quota-gated and fail-closed.** Check `bohr sandbox quota -o json` →
  `max_extra_ephemeral_storage_gb.limit`. The production target system default
  is **50Gi extra**, so the normal maximum is **30Gi base + 50Gi extra = 80Gi
  total**. An explicit user override can be lower or higher. A limit of **0 is
  valid** and means no extra is permitted; it is not "unset". A missing/NULL
  user override inherits the system default.
- If `N` exceeds the effective limit, template creation returns **HTTP 400**
  and creates nothing. It does not succeed with a silent 30Gi fallback. Reduce
  `N` or ask an administrator to change the quota; do not blind-retry.
- `--extra-ephemeral-storage-gb 0` is an explicit request for no extra and
  therefore keeps the 30Gi base. Omitting the field has the same disk-size
  result while preserving "unset" request provenance.
- Alternatively, mount a persistent personal/share disk with
  `bohr sandbox create <template> --mount-user-storage` to keep large model
  weights and datasets off the overlay disk entirely.

## Shared memory (`/dev/shm`)

MPI and multi-process scientific workloads can exceed Kubernetes' small
default `/dev/shm`. Set a template-specific size at creation:

```bash
bohr sandbox template create mpi-vasp \
    --image <immutable-image-path> \
    --sku-name c32_m64_cpu \
    --shm-size 8Gi
```

`--shm-size` maps to request field `shm_size`; `--file` may also contain
`{"shm_size":"8Gi"}`, and an explicit flag overrides the file. Blank explicit
values are rejected by the CLI. The backend performs the authoritative
positive Kubernetes-quantity and provider/memory-limit validation.

After a successful create response, the CLI requires response field
`shm_size` to equal the explicit request. A missing or different echo is an
uncertain outcome: an old backend may have silently ignored the field, but the
template may already have been created. The CLI therefore fails closed and
marks the result as unsafe to retry. Inspect the template first with
`bohr sandbox template list -o json`; do not repeat the same create command
blindly.

Shared memory is **create-time only** for the template. The current
`template update` contract is image-only and cannot change `shm_size`; recreate
the template to change it. When omitted, a Provider with managed shared-memory
policy may calculate a value; an unmanaged Provider retains its historical
Kubernetes default. This automatic shared-memory policy does not infer or
resize the overlay disk.

## Update a template's image or network policy

Repoint an existing template's image without recreating it. This is atomic and
image-only — `sku`/`cpu`/`mem`/`gpu`, extra disk, and `shm_size` stay unchanged.

```bash
bohr sandbox template update <name> --image <new-image>          # repoint the image
bohr sandbox template update <name> --image <new-image> -o json  # machine-readable result
```

## Network policy

Use the existing template commands; do not read and rewrite a full list for a
small change. The server atomically merges, normalizes, deduplicates and retries
concurrent version conflicts. Duplicate additions and absent removals are no-ops.

```bash
bohr sandbox template create <name> --image <image> --sku-name <sku> --network-allow api.example.com
bohr sandbox template update <name> --network-add docs.example.com --network-remove old.example.com
bohr sandbox template update <name> --network-file changes.json
bohr sandbox template get <name> --network-file policy.json
bohr sandbox template update <name> --network-reset
```

`--network-add` and `--network-remove` can be combined, repeated, and contain
comma-separated entries. `--network-file` accepts a JSON object (or `-` for stdin):

```json
{"add":["api.example.com"],"remove":["old.example.com"]}
```

Use `{"replace":["api.example.com","10.20.0.0/16"]}` for full replacement.
`{"replace":[]}` keeps a restricted empty allowlist (DNS baseline remains), while
`{"reset":true}` explicitly removes this template's restrictions. Only internal
IP/CIDR entries means only those internal targets are allowed. Coverage is IPv4.

File and inline operations are mutually exclusive; reset cannot combine with
add/remove; image updates cannot combine with any network update. Optional
`--expected-generation N` makes the write conditional (409 on mismatch), and
overrides any `expected_generation` recorded in a file. Ordinary additions and
removals need no version flag. An exported file includes the complete `replace`
(or `reset`) operation and the observed generation, ready to edit and replay.

`get` and `create` bound network output to `restricted`, `entry_count`,
`entries_preview` (at most 20), `truncated`, and the existing generation field.
`update` returns `template/added/removed/total/generation/changed/restricted/state/max_entries`.
The backend owns the configurable capacity; the CLI only validates formats.
Dry-run is a bounded preview and does not claim server quota validation.

Allowed entry forms: exact hostname (`api.example.com`, lowercase/trailing dot
normalized), IPv4 address (`203.0.113.10`), canonical IPv4 CIDR (`10.20.0.0/16`).
URLs, paths, ports, wildcards, IPv6, empty entries and CIDRs with host bits set
are rejected with a structured validation error. Create's existing `--file` can
provide `network_policy:{"allow":[...]}`; omitted mode defaults to allowlist,
explicit `allow:[]` stays restricted, and a missing policy keeps the old behavior.

Network policy is mutable template configuration and does not rebuild an image
revision. Updates cover matching instances already running the policy agent;
older Pods without it require recreation. `pending`/HTTP 202 means intent saved;
`submitted` means accepted by the cluster, not a network connectivity test.

## Delete a template

Delete a template by name. TTY callers see a confirmation prompt; non-TTY
callers must pass `--force`. The call is idempotent — a missing template comes
back as `skipped` (exit 0).

```bash
bohr sandbox template rm <name>            # interactive confirm on TTY
bohr sandbox template rm <name> --force    # skip prompt (required for non-TTY)
bohr sandbox template rm <name> -o json    # {"deleted": [...], "skipped": [...], "failed": []}
```

For concrete `<image-path>` and `<sku>` discovery, see
[`../platform-snapshot.md`](../platform-snapshot.md). Use `bohr image ls` to get
image paths.

Network-policy `pending/submitted` is submission progress, not proof of traffic enforcement. `template update --wait` waits for image publication only; network updates reject it and require generation/state inspection plus traffic verification.
