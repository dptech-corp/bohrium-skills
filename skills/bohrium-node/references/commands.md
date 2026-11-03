# Node command reference

## Read and connect

| Command | Inputs |
|---|---|
| `bohr node list` | filters `--waiting`, `--pending`, `--started`, `--paused`, `--quiet` |
| `bohr node get <id>` | positional ID or `--node_id` |
| `bohr node resources` | no command-specific flags |
| `bohr node connect <node_id>` | positional ID or `--node_id/-n` |

List/get/resources are free reads. Connect is a free write-class operation that
starts a local interactive SSH client; it may block until the session closes.

## Create

```text
bohr node create [--file <config.json> | flags]
```

Current flags: `--name/-n`, `--image_path/-i`, `--project-id/-P`,
`--turnoff_after/-t` default `24`, required `--disk_size/-d`,
`--machine_type/-m`, and `--file/-f`. JSON accepts current camel/snake forms
documented by command help; do not add fields the CLI does not validate.

Disk size is one of `20, 40, 80, 100` GB. Quote machine names containing
spaces. Create is balance-only and requires billing confirmation.

## Lifecycle

| Command | Inputs/risk |
|---|---|
| `bohr node restart <id>` | optional `--node_id`, `--turnoff_after`; balance-only resource |
| `bohr node stop <node_id>` | positional ID or `--node_id/-n`; free write |
| `bohr node delete <node_id>` | positional ID or `--node_id/-n`; destructive |

Check pricing and machines with:

```bash
bohr billing pricing --resource node -o json
bohr machine list --scene node --choose-type all -o json
```

Create/restart add `--yes` only after spending authorization. Delete adds
`--yes` only after exact-target destructive authorization. Stop does not use
the destructive gate, but still requires explicit lifecycle intent.

Keep the ID returned by create as the requested ID. Detail may include a
different internal `nodeId`/machine ID and a mismatch warning. After deletion,
verify the uniquely named record is absent from `bohr node list`; `get` can
temporarily return stale detail with an inactive status.

The current registered command tree has no separate start command and no
rename, dataset-mount management, or credential-display command.

```bash
bohr guide read bohr-node
bohr guide read bohr-node/references/commands.md
```
