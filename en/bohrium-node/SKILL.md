---
name: bohrium-node
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Create, list, inspect, connect to, restart, stop, or delete Bohrium interactive
  development nodes with `bohr node`. Use for 开发机, long-lived CPU/GPU nodes,
  machine pricing, image selection, disk size, auto-shutdown, SSH connection,
  and node lifecycle. Use Job for non-interactive batch compute.
---

# Bohrium Node

## Choose the task

| Goal | Route |
|---|---|
| Open an interactive development environment | Resources/cost → create → wait for running → connect |
| Reuse a known environment | Resolve Node ID → inspect state → authorized restart if stopped |
| End a session | Save work → stop → verify stopped; delete only for explicit disposal |

## Workflow: create, use, and stop a development node

1. Select project/image/machine, disk and auto-shutdown using the cost preflight below. After authorized create, preserve `data.id` as the Node ID. From a list, use `data.items[].nodeId`.
2. Inspect the same Node ID within the agreed waiting budget until `data.status=2`, then connect; if the deadline expires, retain this ID and report that readiness is unconfirmed:

```bash
bohr node get <node_id> -o json
bohr node connect <node_id>
```

3. `get` preserves the lookup ID in `data.requestedNodeId`; its detail `nodeId`/`machineId` may identify the underlying machine. Continue using the original Node ID for lifecycle/SSH operations. `connect` launches interactive SSH even if global JSON output is selected; use an interactive terminal and keep connection secrets private.
4. Save the user's work, then stop when requested:

```bash
bohr node stop <node_id> -o json
bohr node get <node_id> -o json
```

5. Completion requires the requested environment to be usable, and after an authorized stop, observed `status=-1` from detail/list. Stop acceptance alone is insufficient. If startup or SSH fails, inspect the original node before creating another billable one. Restart may resume charges; delete is a separate destructive decision.

Nodes are interactive, potentially long-lived compute resources. Inspect
available machines and pricing before create or restart:

```bash
bohr billing pricing --resource node -o json
bohr machine list --scene node --choose-type all -o json
bohr node resources -o json
```

Create and restart are balance-only resource operations. Confirm project,
machine, image, disk, auto-shutdown, and spending before adding `--yes`:

```bash
bohr node create --name dev-node --project-id <project_id> \
  --image_path <image> --machine_type <machine> --disk_size 40 \
  --turnoff_after 24 --yes -o json
```

Disk size is required and must be one of **20, 40, 80, 100** GB.
`--turnoff_after` defaults to `24` hours; common values are `1`, `8`, `24`, and
`-1` for never. Never-ending resources require especially explicit cost intent.
A JSON config file is an alternative input; inspect it before submission.

Always preserve the ID returned by create as the requested Node ID. A later detail
response may also expose a different internal machine/node ID; do not silently
replace the create ID in lifecycle commands.

Track and connect by node ID:

```bash
bohr node list --started -o json
bohr node get <node_id> -o json
bohr node connect <node_id>
```

Connect opens an interactive local SSH process. Keep credentials and connection
details out of logs and summaries.

Lifecycle commands:

```bash
bohr node stop <node_id> -o json
bohr node restart <node_id> --turnoff_after 24 --yes -o json
bohr node delete <node_id> --yes -o json
```

Stop is a write; restart may resume billing; delete is destructive. Verify the
returned state instead of treating request acceptance as lifecycle completion.
After delete, confirm the unique Node name/ID is absent from `bohr node list`.
`bohr node get` may briefly return stale detail after deletion, so a stale detail
record alone is not evidence that the Node remains active.
The current CLI has no separate start, rename, or dataset-binding subcommand.

Read [`references/commands.md`](references/commands.md) for every current
command, filter, resource input, price boundary, and lifecycle guardrail.

## Resource payment boundary

Resource actions marked balance-only keep their existing balance billing. The
global `--payment-source` option does not enable photon billing for these
resources; explicitly selecting `photon` on a balance-only action is rejected
before execution. Use the existing resource pricing/authorization workflow.
