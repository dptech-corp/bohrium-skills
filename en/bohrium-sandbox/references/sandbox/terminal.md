# Sandbox terminal sessions (PTY)

PTY sessions for REPLs, TUIs, and Ctrl-C interaction.

> **For "run a command, get its stdout/stderr/exit_code" use `bohr sandbox exec`,
> not `terminal`.** `exec` returns structured output in one round-trip and is
> the agent-ready path. `terminal` exists for the rare cases that genuinely need
> a TTY: REPLs, TUIs (`htop`, `vim`), sending Ctrl-C to a stuck foreground
> process, programs that detect an interactive terminal and refuse otherwise.

Open long-lived PTY shells inside a sandbox for interactive workloads. Four
subcommands:

```bash
bohr sandbox terminal attach <sandbox_id>                       # interactive TTY (docker exec -ti style)
bohr sandbox terminal create <sandbox_id> -o json               # open pty, default timeout=0 (long-lived)
bohr sandbox terminal create <sandbox_id> --cwd /workspace --env K=V -o json
bohr sandbox terminal send   <sandbox_id> <pid> 'echo hi\n'     # write bytes to stdin (newline executes)
bohr sandbox terminal kill   <sandbox_id> <pid> -o json         # kill the pty pid only
```

## Interactive attach (`docker exec -ti`)

`bohr sandbox terminal attach` is the real interactive path: it bridges your
local terminal to a sandbox PTY in **both** directions, so keystrokes go in and
output comes back live — exactly like `docker exec -ti <container> bash`. This
is the one PTY subcommand that echoes output; `create` / `send` do not.

```bash
bohr sandbox terminal attach <sandbox_id>                 # fresh shell, sized to your terminal
bohr sandbox terminal attach <sandbox_id> --cwd /workspace
bohr sandbox terminal attach <sandbox_id> --pid 42        # re-attach to an existing pty
```

Semantics:

- **Detach vs exit.** Press `Ctrl-]` to detach and leave the PTY running (the
  command prints its pid so you can `attach --pid <pid>` again later). Exit the
  shell normally (`exit` / `Ctrl-D`) to end the PTY.
- **New vs existing.** Without `--pid` a fresh PTY is created (default
  `--timeout 0`, long-lived) as `root`, sized to your local window. Pass
  `--user <user>` only to request another identity. With `--pid`, attach
  connects to an already-running PTY and cannot change that PTY's user.
- **Resize** is mirrored automatically, so full-screen TUIs reflow when you
  resize the window.
- **Requires an interactive POSIX TTY.** If stdin is not a terminal (piped,
  redirected), `attach` refuses and points you at `exec`. For scripted,
  structured output always prefer `bohr sandbox exec`.

Key semantics:

- `terminal send` writes bytes **as-is** — there is no implicit newline. Append
  `\n` to make a shell run the line; send raw control bytes (e.g. `$'\x03'` for
  Ctrl-C) to interrupt a running foreground command.
- `terminal send` joins multiple text arguments after `<pid>` with spaces.
  Pass quoting-sensitive terminal input as one argument. Use `exec` when the
  operation needs its argv-preserving command interface instead of terminal text.
- `terminal kill` terminates **only** the pty pid. The sandbox keeps running;
  use `bohr sandbox delete <sandbox_id>` to destroy the sandbox itself. For a
  non-PTY process, use `bohr sandbox ps <id> --pid <pid>`,
  `bohr sandbox process connect <id> <pid>`, or the destructive
  `bohr sandbox process kill <id> <pid>`. Those arbitrary-process operations
  do not prove historical output or confirmed termination.
- `--timeout` defaults to `0` (no auto-destroy). Always `terminal kill` the pid
  when finished — long-lived terminals continue to consume sandbox resources.
- `--cols` / `--rows` set the PTY size for a freshly created terminal (defaults
  `80` x `24`; `attach` auto-detects your local window). Pass them explicitly
  when you need a fixed size for a headless/non-TTY `create`, e.g.
  `terminal create <id> --cols 120 --rows 40`.

## Capturing PTY output

`terminal send` returns only `sent_bytes` — the PTY's stdout is **not** echoed
back. The PTY is a streaming session inside the sandbox, but each `bohr`
invocation is a one-shot RPC, so there is no per-process buffer to read from
afterwards. Two patterns work:

- **Preferred — use `exec`:** `bohr sandbox exec <sandbox_id> --command '<cmd>'` already
  returns `stdout` / `stderr` / `exit_code` in a single call.
- **PTY-bound case — redirect, then read:** if the workload genuinely needs a
  TTY, redirect each command's output to a file inside the PTY and read it back.
  Redirect per-command rather than once at the top of the session.

  ```bash
  bohr sandbox terminal send <sandbox_id> <pid> $'cmd > /tmp/out 2>&1\n'
  bohr sandbox files read    <sandbox_id> /tmp/out
  ```
