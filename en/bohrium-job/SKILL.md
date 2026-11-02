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
