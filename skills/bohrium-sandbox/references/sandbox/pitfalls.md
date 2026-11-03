# Default-behavior pitfalls

Every item below is a **default** that is safe for quick interactive use but
bites long-running / ML-training workloads. Read this once before driving a
sandbox through an agent; each row links to the deep doc.

| # | Default | When it bites | Do this instead |
|---|---------|---------------|-----------------|
| 1 | Overlay (writable root) capacity is independent of CPU/mem/GPU SKU | Large images or writable data can exhaust rootfs | Inspect actual write paths with `storage inspect <id> -o json`; use each known measurement and never treat `null` as zero. `complete` is evidence completeness, not sufficient space. If growth is needed, require `capabilities.dynamic_adjustment.available=true` before `storage resize <id> --target <N>Gi`; require resize `status=ready`. False or missing availability needs another plan, not an automatic retry. Do not infer disk size from compressed image layers. See [`storage.md`](./storage.md) and [`templates.md`](./templates.md). |
| 2 | `files read` defaults to auto format on stdout | Redirecting base64 stdout to a model/archive filename does not produce the original binary; text rejects invalid UTF-8 | Download with `--destination` to preserve original bytes. For encoded stdout use `--format bytes`; inspect `content_base64` / `byte_count` in JSON. See [`files.md`](./files.md). |
| 3 | `exec` has no execution deadline | Network disconnection can still interrupt the stream; this does not prove remote termination | Save the sandbox ID and confirmed PID; inspect logs and outputs before rerunning. See [`execution-modes.md`](./execution-modes.md). |
| 4 | `exec --background` returns a sandbox PID, not task completion | Detach guarantees neither remote survival nor termination | Persist logs and inspect the recorded PID. Deprecated `--timeout` only accepts 0. See [`exec.md`](./exec.md). |
| 5 | `delete` is **irreversible**; a stopped sandbox's files cannot be read | You `delete` before pulling outputs → the contents of `/workspace/` are gone for good | Always `files read` everything you need **before** delete. Convention: write outputs to `/workspace/out/`. See [`execution-modes.md`](./execution-modes.md). |
| 6 | Outbound HTTP proxy is **off** | Overseas installs (`pypi.org`, GitHub, HuggingFace) hang / fail | Toggle the `proxy.example.com` proxy on for the overseas step, off afterwards. Domestic Aliyun mirror is unaffected. See [`network.md`](./network.md). |
| 7 | A mutable `:latest` template image is **rejected** | `template create/update --image foo:latest` fails | Use an immutable / unique tag (a version, or the unique tag `image build` / `image commit` produces). Direct `create --image` accepts `:latest`, but immutable references are preferable for reproducibility and avoiding first-use cold pulls. See [`lifecycle.md`](./lifecycle.md) and [`templates.md`](./templates.md). |
| 8 | `delete` needs explicit confirmation and a running-process pre-check; deleting a parent can cascade to all recorded descendants | In an agent / CI / piped shell, confirmation returns `CONFIRMATION_REQUIRED`; after `--yes`, a busy/unverifiable sandbox returns `PRECONDITION_FAILED`. On a TTY, a parent with descendants also gets a cascade preview and second confirmation | Non-interactively pass `--yes`; use `-f/--force` only to confirm and skip the process pre-check. It does not change backend finalization semantics: for Wenyon it is **not** a discard/end-only path. Server-side cascading still applies without an interactive preview. See [`lifecycle.md`](./lifecycle.md). |
| 9 | Per-user **running-sandbox quota** (Running + Hibernating) | `create` → 400 `running sandbox limit reached (N/M)...` | Before create **and** on that error: `quota -o json` → `list` → delete an idle/hibernating sandbox → retry. Never blind-retry. A writable Wenyon delete requires explicit `--commit` or `--discard`; `--force` does not choose either. See [`lifecycle.md`](./lifecycle.md). |
| 10 | A new sandbox does **not** inherit the caller's bohr login | Inner bohr-cli starts another authorization flow, or an agent copies a secret through `--env` | Only with explicit user authorization, use `create --inherit-auth`. It injects after create through the data plane and requires a statically verified npm bohr-cli 2.6.45+ for both credential kinds; Vouch has no refresh token. On `PARTIAL_FAILURE`, delete the returned sandbox without inspection/reuse instead of retrying create; reconcile `sandbox_create_outcome_uncertain` by request ID first. See [`lifecycle.md`](./lifecycle.md). |
| 11 | `create --image` without `--cpu`, an explicit template, or `--gpu` uses CPU profile `1c2g` (**1 CPU / 2 GiB, no GPU**) | A training/GPU image is mistaken for a request for more resources, then OOMs or has no accelerator | `--image` only swaps the image. Select another supported shape with `--cpu`, use an explicit template outside the profile list, or `--gpu`; verify `templateID`, `cpuCount`, and `memoryMB`. See [`lifecycle.md`](./lifecycle.md). |
| 12 | Wenyon mounting is specified per instance | Expecting template defaults to supply data | Pass dataset/path to sandbox create. Default read/current pins data; write/current preloads it; write/none alone starts empty. |
| 13 | `--wenyon-user-token-env` takes an **environment-variable name**; `BOHR_ACCESS_KEY` is not the Wenyon/Vouch user JWT | An Agent passes a secret literal/AccessKey, treats `bohr auth token --as user` as Vouch even though it reports `auth_method=access_key`, retries an auth error, or assumes `--timeout` controls the HTTP wait | Export the JWT through a secret/session as `WENYON_USER_JWT`, then pass `--wenyon-user-token-env WENYON_USER_JWT`; stored login is not read implicitly. The JWT must not enter body/metadata/output. `sandbox create --timeout` is only Sandbox lifetime; fix auth/readiness errors before retrying. See [`lifecycle.md`](./lifecycle.md). |
| 14 | A writable Wenyon workspace needs an explicit end decision | Treating delete --force as saving or discarding | Choose --commit or --discard; --wait observes final settlement. Failed execution never automatically publishes. Inspect recovery_deadline. |

## ML-training checklist

Before launching a training job through an agent:

1. **Quota.** Slot available? `bohr sandbox quota -o json` — if
   `running_sandboxes.remaining == 0`, `delete` an idle/hibernating sandbox
   first.
2. **Storage.** Run `bohr sandbox storage inspect <id> -o json`; do not infer
   disk size from compressed image layers or treat unknown values as zero.
   If the storage command group is disabled, enter the target sandbox and run
   `df -h / /tmp` with the actual write paths; use `df -i` for inode availability.
   This read-only fallback measures usage and does not enable resize.
   Check the actual output and temporary paths against expected writes; complete
   evidence or a healthy low watermark is not a workload-space guarantee. If
   space is insufficient, require `capabilities.dynamic_adjustment.available=true`
   before resize and `status=ready` after it. False or unknown support must not
   trigger automatic resize, cleanup or replacement.
3. **Lifetime.** Long job? Create with `--timeout 0` or `--never-timeout`
   (and remember to `delete` manually).
4. **Execution mode.** Choose foreground or `exec --background` based on the workload; save the confirmed sandbox PID and outputs.
5. **Outputs.** Write to `/workspace/out/`; `files read` them **before** delete.
6. **Network.** Overseas datasets/models? Toggle the proxy on, then off.
7. **Binary downloads.** Pull checkpoints with `files read --format bytes`.
8. **Wenyon semantics.** Choose read/current, write/current or explicit write/none
   on sandbox create. End writes with --commit or --discard; verify the final
   produced_version after commit instead of assuming accepted means saved.
