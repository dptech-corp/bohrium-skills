# Project command reference

| Command | Inputs | Effect |
|---|---|---|
| `bohr project list` | none | List visible projects |
| `bohr project get <id>` | project ID | Read project detail |
| `bohr project members <id>` | project ID | Read member list |
| `bohr project create` | flags below | Create project |
| `bohr project delete <project_id>` | project ID or `--project-id` | Destructive delete |

Create flags:

| Flag | Default | Meaning |
|---|---:|---|
| `--name/-n` | `default` | Project name |
| `--total_cost_limit/-t` | `0` | Total project cost-limit input |
| `--month_cost_limit/-m` | `0` | Monthly project cost-limit input |

The CLI sends these values at creation. Confirm intended limit semantics from
the returned project or platform policy; do not invent currency, reset dates,
or unlimited behavior.

All current Project commands are free. Create is a write. Delete is destructive
and requires `--yes` for non-interactive execution after the exact project ID
and impact are confirmed.

Project members are read-only in this command surface. The member list does not
authorize mutation or reveal that a user can be safely removed.

```bash
bohr guide read bohr-project
bohr guide read bohr-project/references/commands.md
```
