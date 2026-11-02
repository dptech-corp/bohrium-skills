---
name: bohrium-job
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Submit, list, inspect, stop, delete, download, or group Bohrium compute jobs
  with `bohr job` and `bohr job-group`. Use for BOHR_ID, machine/image selection,
  scientific batch computation, job status, logs/results, queued cancellation,
  running termination, or job-group lifecycle.
---

# Bohrium Job

## Choose the task

| Goal | Ordered route |
|---|---|
| Run a batch calculation | Resolve resources/cost → submit → track Bohr ID → verify terminal outcome → download |
| Recover an existing calculation | Describe/log the original Bohr ID; a new submit starts another job |
| Stop work | Resolve phase → terminate running / kill queued → confirm terminal state |

## Workflow: submit a calculation and retrieve its outputs

1. Resolve the project, image, machine and authorized resource budget using the preflight below. Confirm input and result paths before submit.
2. For an output-producing run, put the collection contract in `job.json` (these are config keys, not public flags):

```json
{
  "out_files": ["result.csv"],
  "log_files": ["run.log"]
}
```

The selected image must provide Python; `./input/run.py` must write `result.csv` in the job working directory. Submit once after filling the resolved resource values:

```bash
bohr job submit -i ./job.json --project-id <project_id> --job_name analysis \
  --image_address <image> --machine_type <machine> \
  --command "python run.py > run.log 2>&1" --input_directory ./input \
  --result_path /personal/results --yes -o json
```

`--result_path` is remote storage; download's `--out` is local. Capture identifiers from the actual successful response:

| Returned value | Next consumer |
|---|---|
| `data.bohrJobId` | `job describe --id`, `job log`, `job download`, lifecycle commands |
| `data.bohrJobGroupId` | Later `job submit --job_group_id` |
| `data.jobGroupId` | Immediate job-group lifecycle operations |
| `data.jobId` | Internal platform ID; retain as evidence rather than substituting it for Bohr ID |

```bash
bohr job describe --id <bohr_job_id> -o json
bohr job download <bohr_job_id> --out ./results -o json
```

3. Replace `<bohr_job_id>` with `data.bohrJobId` from submission (or the selected public Bohr ID from listing). For a single describe, inspect `data.phase`, `data.terminal`, and `data.exitCode`. Download the requested results after `phase=completed` and exit code `0`; a queued/running or accepted response is not completion.
4. For a failed terminal job, inspect `errorInfo` and logs for that same Bohr ID before deciding whether new input/compute is needed. Preserve failed-run diagnostics; another submit is another resource action.
5. Complete with the terminal outcome and actual local artifacts from download's `data.files[].path`/`bytes`. A completed job with failed or missing download is compute success with an unfinished delivery. `JOB_FILES_EMPTY` means no downloadable artifacts were found; check the collection paths and preserve the run evidence. A monitoring timeout retains the Bohr ID for resumption; it does not stop the job.

## Resolve resources and cost before submission

Job compute is balance-only and priced by machine type and actual usage. Inspect
the catalog and available machines before authorizing a submission:

```bash
bohr billing pricing --resource job -o json
bohr machine list --scene job --choose-type all -o json
```

Confirm the project, machine, image, command, node count, runtime limits, local
input directory, remote result path, and spending authorization. Then submit:

```bash
bohr job submit --project-id <project_id> --job_name <name> \
  --image_address <image> --machine_type <machine> --command <command> \
  --input_directory ./input --result_path /personal/results --yes -o json
```

Omit `--input_directory` for empty input. A config file can supply complex
settings, but explicit flags and file values must not conflict.

## Track jobs by Bohr ID

The `BOHR_ID` column from `bohr job list` is the public job identifier for
describe, terminate, kill, delete, download, and log. The CLI resolves the
internal platform job ID automatically:

```bash
bohr job list --running -o json
bohr job describe --id <bohr_id> -o json
```

Use normalized `phase` and `terminal` fields when present. An accepted stop
request or a non-terminal status is not completion.

Running jobs use `bohr job terminate`; queued jobs use `bohr job kill`. By
default the CLI confirms a terminal phase, bounded by `--confirm-timeout`.
`--no-wait` returns after acceptance and must be reported as unconfirmed.

`kill` and `delete` are destructive and require exact-target authorization and
`--yes`. `terminate` is a non-destructive stop request but still requires clear
user intent. Logs and results write local files; confirm `--out` before download.

## Job groups

`bohr job submit` can return both `jobGroupId` and `bohrJobGroupId`:

- `--job_group_id` on a later submit consumes `bohrJobGroupId`.
- Group terminate/delete should prefer `jobGroupId` immediately after submit;
  `bohrJobGroupId` becomes resolvable only after jobs are indexed.

Use canonical `bohr job-group` commands for group create/list/terminate/delete
and result download. Group termination and deletion are destructive.

Read [`references/commands.md`](references/commands.md) for every current
command, flag, ID type, pricing boundary, status, and stopping rule.

## Resource payment boundary

Resource actions marked balance-only keep their existing balance billing. The
global `--payment-source` option does not enable photon billing for these
resources; explicitly selecting `photon` on a balance-only action is rejected
before execution. Use the existing resource pricing/authorization workflow.
