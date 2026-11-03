# Job command reference

## Job discovery and detail

| Command | Current inputs |
|---|---|
| `bohr job list` | `--job_group_id/-j`, `--number/-n` default `10`, phase filters, `--quiet` |
| `bohr job describe` | repeatable `--id/-i` Bohr IDs or `--job_group_id/-j` |

Phase filters: `--fail`, `--pending`, `--running`, `--finish`, `--scheduling`,
`--stopping`, and `--stopped`. Keep the original numeric status fields for
diagnostics, but automate from the normalized `phase`/`terminal` contract.

## Submission

```text
bohr job submit [--config_file <json>] [flags]
```

Current flags: `--input_directory`, `--job_group_id`, `--project-id`,
`--job_name`, `--image_address`, `--machine_type`, `--command`, `--result_path`,
`--log_file`, `--max_reschedule_times`, `--max_run_time`, and `--nnode`
(default `1`). `--result_path` must start with `/personal` or `/share`.

Submission creates a billable resource from balance:

```bash
bohr billing pricing --resource job -o json
bohr machine list --scene job --choose-type all -o json
```

Add `--yes` only after the exact configuration and spending are authorized.
Input upload and scheduling can fail independently; preserve request IDs and do
not report a Job as running from submission acceptance alone.

## Lifecycle and local retrieval

| Command | Identifier/effect |
|---|---|
| `bohr job terminate <bohr_id>...` | Running-job stop; default terminal confirmation |
| `bohr job kill <bohr_id>...` | Queued/running cancellation; destructive |
| `bohr job delete <bohr_id>...` | Destructive record/result deletion |
| `bohr job download <bohr_id>... --out <dir>` | Write result files locally |
| `bohr job log <bohr_id>... --out <dir>` | Write log files locally |

All accept positional Bohr IDs; lifecycle and download commands also accept
repeatable `--id/-i`. Stop commands support `--confirm-timeout` default `1m`.
Global `--no-wait` skips terminal confirmation. For queued jobs, kill may return
after acceptance unless global `--wait` is explicitly supplied.

## Job groups

| Command | Inputs/effect |
|---|---|
| `bohr job-group create` | `--name/-n`, `--project-id` |
| `bohr job-group list` | `--number`, `--project-id`, `--job_group`, date range, sort, quiet |
| `bohr job-group terminate <id>...` | Destructive group finish |
| `bohr job-group delete <id>...` | Destructive group deletion |
| `bohr job-group download` | `--job_group_id/-j`, `--number`, `--out` |

Group list date filters require both `--start` and `--end` in `yyyy-mm-dd`.
Use `jobGroupId` for immediate terminate/delete after submit. Use
`bohrJobGroupId` as `job submit --job_group_id`; group lifecycle can resolve it
after indexing. Preserve both fields instead of renaming either to “group ID.”

Use canonical commands rather than hidden `+submit`, `+watch`, or `+cancel`
aliases.

```bash
bohr guide read bohr-job
bohr guide read bohr-job/references/commands.md
```
