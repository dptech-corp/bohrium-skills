# Sandbox exec

`bohr sandbox exec` usage, including `--background` mechanics.

For the foreground vs background vs PTY mental model and the
retrieve-before-delete SOP, see [`execution-modes.md`](./execution-modes.md).

## exec command passing

Positional command arguments preserve each argument, including spaces, empty
strings, quotes and shell metacharacters. A standalone `--` stops local flag parsing
when the remote argv contains flags. Shell operators, expansions and pipelines
belong in the explicit `-c/--command` string:

```bash
bohr sandbox exec <sandbox_id> -- python3 -c 'import sys; print(sys.argv[1:])' '' 'a b'
bohr sandbox exec --command 'cd /workspace && python train.py' <sandbox_id>
bohr sandbox exec --command 'cat log.txt | grep ERROR | wc -l' <sandbox_id>
bohr sandbox exec --command 'echo hello > /tmp/out.txt' <sandbox_id>
```

Commands run as `root` by default, including background commands. Agents should
use that default for sandbox work. Pass `--user <user>` only when an explicit
non-root identity is required. You can also select a working directory and
inject environment variables:

```bash
bohr sandbox exec --user <user> --cwd /workspace --env FOO=bar --command 'env | grep FOO' <sandbox_id>
```

Instead of positional args you may pass the whole command via `-c/--command`.
This is the safest form when the command depends on nested quoting:

```bash
bohr sandbox exec -c 'python train.py --lr 0.1' <sandbox_id>
```

Cobra parses bohr flags before or after `<sandbox_id>` until a standalone `--`.
Use `--` when remote tokens could be interpreted as local flags; positional
argv boundaries are preserved independently of that separator. Prefer flags before the ID in
examples so the local/remote split is obvious.

## Local script files (`--file`)

```bash
bohr sandbox exec --file ./script.sh <sandbox_id>
```

`--file` opens a local regular file without following a symlink in the final
path component, uploads its bytes to `/tmp/bohr-exec-<uuid>.sh`, then runs
`bash <remote_path>`. Parent-directory links are outside this guarantee; on
Windows all reparse points are rejected, including some deduplicated/cloud
files, and runtime behavior is unverified. Do not modify the file concurrently:
the opened file is not a frozen content snapshot. `--file` is mutually exclusive
with `-c/--command` and positional commands. Use LF; CRLF is not converted.
The upload has a 90-second total request limit, even if progress continues.

Before Start, upload failure attempts cleanup. Once Start has been attempted,
if its outcome or the subsequent wait/detach is unknown, the script is retained
and a non-retryable receipt identifies the sandbox, known PID and script path;
inspect before retrying or removing it. A bounded preview of received stdout/stderr
is retained with byte counts and truncation state; known credentials are redacted. Unknown receipts ignore `--jq`; after Start
or an observed end, an invalid filter falls back to the full receipt. The raw end-event error is withheld because the runtime may echo script contents or credentials. An observed end
event triggers best-effort foreground cleanup, including nonzero exits.
Background success retains the script and prints its path on stderr, including
with `--jq`. Cleanup failure does not replace the exit code. Dry-run validates the file and shows the short command, path and
byte count without connecting, uploading, or exposing the script contents.

The CLI rejects a final command longer than **96 KiB (98304 UTF-8 bytes)**,
including in dry-run. This is a conservative local guard, not an authoritative
server limit or an ARG_MAX probe. Large scripts passed through `--file` use a
short command and are not limited by the script's byte count.

The npm launcher forwards arguments to the native binary. `E2BIG` or
`ENAMETOOLONG` before that binary starts is a host OS/shell launch failure; it
cannot reach the CLI guard and does not mean Go shortened the arguments.
Windows shell-specific limits, including the reported 8191-character case,
remain unverified on a Windows machine. Use `--file` for long scripts.

## Exit codes

The command's exit code becomes the CLI's own exit code (like `ssh` /
`docker exec`), so `bohr sandbox exec <id> 'false'` exits non-zero. Structured
output (`-o json`) carries `data.exit_code`, `data.stdout`, `data.stderr`, and
`data.error`; an observed nonzero exit has `ok:false` while retaining those
fields and the same process exit code. A broken execution channel instead
returns `sandbox_exec_outcome_uncertain`, `retryable:false`, and known IDs/PID.
Script executions retain the published `sandbox_exec_outcome_unknown` receipt, partial output and script path. Explicit argument/authentication rejections retain their failure category and do not retain an unused uploaded script.
Foreground reports the sandbox ID and confirmed PID on stderr before waiting. On interrupted ordinary exec, the failure receipt retains bounded known stdout/stderr and indicates truncation; non-UTF-8 bytes are replaced, not used to discard later valid output. Inspect the process and any task receipts before another execution; an absent PID does not prove the command never ran.

## Background jobs (`--background`)

`--background` returns a confirmed `pid` and disconnects the client stream
without explicitly killing the remote process. Exec has no execution deadline;
sandbox lifetime and network disconnections remain independent. Foreground
announces the sandbox ID and confirmed PID on stderr before waiting. The old
`--timeout` flag is deprecated; during transition only `0` is accepted.
Re-check state with `bohr sandbox ps <sandbox_id>` or by reading output files;
use the recorded PID to stop a process explicitly (see below).

```bash
bohr sandbox exec --background --command 'python train.py > /tmp/out.log 2>&1' <sandbox_id>
bohr sandbox ps <sandbox_id> -o json
bohr sandbox files read <sandbox_id> /tmp/out.log
```

Background outputs live only inside the sandbox — pull anything you need with
`bohr sandbox files read` before `bohr sandbox delete`.

## Identify and stop one process

An exec runs inside a login `bash -l -c` whose arguments contain the entire
command string. `pgrep -f` and `pkill -f` can match that ancestor wrapper,
even when the search tool excludes its own PID. A match alone does not prove
that the workload is still running; a broad kill can stop the control shell
or another task.

Record the PID returned by `--background`, together with the sandbox ID and
the task it belongs to. Inspect only that exact PID:

```bash
bohr sandbox ps <sandbox_id> --pid 4242 -o json
```

This keeps the existing full projection, including raw `cmd`, `args`, `envs`,
and the convenience `command_line` that removes the bash wrapper. The current
`ps` projection can contain sensitive environment values; do not quote it into
a public report. In images with procps, also record and compare the process
start time before sending a signal:

```bash
bohr sandbox exec -c 'ps -p 4242 -o pid=,ppid=,lstart=,args=' <sandbox_id>
```

Replace `4242` with the recorded PID. To observe output from the current
connection stream, run:

```bash
bohr sandbox process connect <sandbox_id> 4242 -o json
```

This does not promise replay of output produced or buffered before attachment.
If an end event arrives, the envelope carries the observed exit code, and a
nonzero exit has `ok=false` while the CLI exits with that same code. No end event
or a disconnected stream is outcome-unknown: inspect the original execution,
do not blindly retry. Exec has no execution deadline; the control-plane lifetime
hint remains fixed at 300 seconds and may extend a near-expiry sandbox.
Process output is unredacted and may contain sensitive information; the new
process summary withholds untrusted tag, program, and working-directory values.

For a workload-specific graceful stop, confirm the identity first and retain
the existing TERM flow:

```bash
bohr sandbox exec -c 'kill -TERM -- 4242' <sandbox_id>
bohr sandbox ps <sandbox_id> --pid 4242 -o json
```

The platform also exposes exact-PID SIGKILL. It is destructive and confirms
only after resolving the target:

```bash
bohr sandbox process kill <sandbox_id> 4242 --yes -o json
```

`signal_accepted=true` means only that the SendSignal RPC was accepted;
`termination_confirmed` is always false. `(false,nil)` is not-found with no
signal sent. A transport error is outcome-unknown because the signal may already
have been accepted, so do not retry automatically. Signal receipts ignore
`--jq` to preserve the complete outcome. Recheck with `ps --pid`.

PIDs can be reused. If the recorded process has disappeared or its identity
changed, stop cleanup rather than signaling its replacement. Inspection and
signaling are not atomic: workloads needing a guaranteed identity or cleanup
of all descendants need an explicit remote supervisor. A PID is not a process
group; signaling the wrapper alone does not guarantee its children stopped.
The data-plane token can target any PID visible in the sandbox.
Use `sandbox terminal attach` or `sandbox terminal kill` for known PTY
sessions, not as a general arbitrary-process or process-group operation.

A disappeared PID is not proof of success, and a zombie is not an executing
workload. After a wrapper is signaled, its children can keep running under
PID 1, and `sandbox ps` can keep listing the wrapper's PID until they exit.
Check the state before treating a listed PID as running:

```bash
bohr sandbox exec -c 'ps -p 4242 -o pid=,stat=,args=' <sandbox_id>
```

A `Z` state (`[bash] <defunct>`) means the wrapper itself has exited, not that
its children stopped.

Background exec does not retain a durable exit code after the process ends.
Save the workload's exit status and logs in task-specific files when
completion evidence is required. See [execution modes](./execution-modes.md)
for retrieving outputs before destroying the sandbox.

## Environment files and safe summaries

Use `--env-file path` for explicit local UTF-8 `KEY=VALUE` files (repeatable,
maximum 1 MiB per regular file; final-component symlinks and Windows reparse
points are rejected). Blank lines and lines beginning with # after
leading whitespace are ignored; CRLF works. Keys use letters, digits and
underscores, starting with a letter or underscore. Empty values are valid.
Values are literal: quotes, spaces, #, = and $ remain unchanged. Shell `export`,
interpolation and dotenv unquoting are not supported.

Files apply in order, then repeated `--env` overrides them (last value wins).
Only these explicit settings are sent; the host environment is not inherited.

```bash
bohr sandbox exec --env-file ./process.env --env MODE=check --dry-run <sandbox_id> true
bohr sandbox exec --env-file ./process.env -o json <sandbox_id> true
```

Dry-run and JSON foreground/background receipts include `environment.keys`,
`count`, `file_count` and `source: client_request`. Human execution writes this
summary to stderr. Summaries omit values and describe the client request;
verify actual process settings using assertions inside the sandbox. Commands
that print their own environment can still reveal values in their stdout.
For a local allowlisted support bundle, see `bohr support capture --help`.
