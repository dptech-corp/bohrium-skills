# Execution modes

Foreground vs background vs PTY: which `bohr sandbox` mode survives a local
interruption, and how to retrieve outputs without losing them.

## The three modes

| Mode | Invocation | Local connection | Where it runs | Client timeout and remote lifetime | Output retrieval |
|------|------------|------------------|---------------|----------------------|------------------|
| Foreground | `bohr sandbox exec <id> --command '<cmd>'` | blocks until done | sandbox | no execution deadline; network disconnect remains possible; known partial output is retained in an unknown receipt | inline `stdout` / `stderr` from the same call |
| Background | `bohr sandbox exec --background <id> --command '<cmd>'` | returns immediately with `pid` | sandbox | no execution deadline; detach does not guarantee remote survival or termination | `bohr sandbox process connect` observes only output received after attach; exact-PID state via `ps --pid` / `describe --with-processes`; persistent logs still need files |
| Terminal (PTY) | `bohr sandbox terminal create / send / kill` | persistent PTY pid | sandbox | PTY pid lives until `terminal kill <pid>` or sandbox destruction (per-PTY `--timeout` available) | redirect each command (`cmd > /tmp/out 2>&1\n`) and read the file with `bohr sandbox files read` |

In all three cases the work happens inside the sandbox; what differs is the
lifetime of the local connection and how the caller gets the output back.

Foreground/background commands and newly created PTYs run as `root` by default.
Use `--user <user>` only when an explicit non-root identity is required. An
existing PTY keeps its original user when re-attached with `--pid`.

## When to use which

- **Foreground** — quick, deterministic commands that finish well under a minute
  (`pwd`, `nvidia-smi`, `pip show <pkg>`, a single test). The caller blocks, the
  result comes back in one round-trip.
- **Background** — anything that might exceed your local connection's reliable
  window: model training, long downloads, setup, dataset preprocessing, long
  test suites. Detaching closes the client stream without explicitly killing
  the remote process; it guarantees neither continued execution nor termination.
  Omit `--timeout`; during the deprecation period only `--timeout 0` is accepted.
  `process connect <id> <pid>` can observe output after it attaches, but it is not a historical log.
- **Terminal (PTY)** — interactive workloads that genuinely need a TTY: REPLs,
  TUIs (`htop`, `vim`), or sending Ctrl-C to a stuck process. Not a substitute
  for `exec` when you just want "run this, give me stdout". For a real
  `docker exec -ti` experience (live two-way I/O in your own terminal), use
  `bohr sandbox terminal attach <id>` — it streams output back, unlike
  `terminal send`. See [`terminal.md`](./terminal.md).

## Long-running jobs: the retrieve-before-delete SOP

This is the safe sequence for any `--background` job whose outputs you need:

1. **Decide where outputs go.** Pick a persistent path inside the sandbox —
   convention is `/workspace/out/`. Bake it into the command; do not rely on the
   default cwd.
2. **Launch with `--background`.**
   ```bash
   bohr sandbox exec --background --command 'mkdir -p /workspace/out && python train.py > /workspace/out/run.log 2>&1' <id>
   ```
   The command returns a `pid`. Record it.
3. **Poll until done.**
   ```bash
   bohr sandbox ps <id> --pid <pid> -o json
   bohr sandbox describe <id> --with-processes -o json
   ```
   The `pid` disappears when the command finishes. You can tail the durable
   log or observe output from a new connection:
   ```bash
   bohr sandbox files read <id> /workspace/out/run.log
   bohr sandbox process connect <id> <pid> -o json
   ```
   `process connect` only reports output received on that connection stream; an
   end event supplies the observed exit code, while a missing end event is an
   unknown outcome. A missing PID does not establish success; persist the
   workload's exit status with its outputs. For PID reuse, wrapper matching and
   single-process termination, follow
   [Identify and stop one process](./exec.md#identify-and-stop-one-process).
4. **Retrieve outputs BEFORE delete.** Pull every file you care about locally:
   ```bash
   bohr sandbox files read <id> /workspace/out/run.log --destination ./run.log
   bohr sandbox files read <id> /workspace/out/model.bin --format bytes --destination ./model.bin
   ```
5. **Verify locally.** Confirm what you needed is on disk (size > 0, expected
   lines present, checksum matches).
6. **Then delete.**
   ```bash
   bohr sandbox delete <id>
   ```

**Why this matters.** The platform does not support reading files from a stopped
sandbox. A `--background` job can continue remotely, subject to
sandbox lifetime and other failures. If you `delete` the sandbox before step 4, the contents of `/workspace/` are gone
for good. Treat the sandbox as an ephemeral compute node: outputs you want to
keep have to be pulled out before teardown.

## Common mistakes

- **Assuming foreground cannot disconnect.** Exec has no execution deadline,
  but network failures can still interrupt the stream. For long work use
  `--background`, save the PID and persist logs. Foreground reports a confirmed
  PID to stderr before waiting.
- **`--background` then immediate `delete` with no `files read` in between.**
  The remote command may have succeeded — you still lost its output. Always
  retrieve first.
- **Using exec timeout as a remote kill timer.** Exec sets no execution deadline;
  nonzero deprecated `--timeout` is rejected. Inspect the recorded PID and stop
  it explicitly when needed (see [`exec.md`](./exec.md)).
- **Treating `process connect` as a replay log.** It promises only output
  received on the new stream, not bytes emitted or buffered before attachment.
- **Treating `process kill` acceptance as termination.** The RPC acceptance is
  not delivery or termination confirmation; recheck the exact PID with `ps`.
- **Treating `terminal send` as run-and-capture.** `send` returns only
  `sent_bytes` — the PTY's stdout is not echoed back. For "run a command, get
  its output" use `exec`. When the workload truly needs a TTY, redirect each
  command (`cmd > /tmp/out 2>&1\n`) and read the file with `files read`.

## See also

- [`exec.md`](./exec.md) — `bohr sandbox exec` usage including `--background`
- [`terminal.md`](./terminal.md) — PTY sessions
- [`lifecycle.md`](./lifecycle.md) — `delete` semantics and best practices
- [`files.md`](./files.md) — `files read` / `files write`


## Verified export before delete

Use `sandbox delete <id> --export /workspace/result.csv --export-dir ./result
--yes` to export only explicitly selected paths before the existing delete
pre-check and DELETE. Repeat `--export` for additional paths. Stop writers first:
there is no snapshot or post-delete recovery.

The local `delete-export-receipt.json` is mode 0600 and records only the selected
paths, size/hash/status and export gate. It contains sensitive metadata, so keep
the export directory private. `gate:passed` proves the export stage passed; it
does not prove deletion was accepted or completed. Confirm the sandbox becomes
`RESOURCE_NOT_FOUND` separately.

A download/hash/receipt failure blocks DELETE. A completed but unverified item,
including any explicit directory export, also blocks deletion by default with
`EXPORT_VERIFY_REQUIRED`. `--export-allow-unverified` accepts those completed
unverified exports only; it never bypasses transfer errors or hash mismatch.
Without `--export`, the original delete flow remains. Dry-run validates and
shows mappings without creating local exports or contacting the service.
