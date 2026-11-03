# Public Skills versus published CLI commands

Audit baseline: PR #32 at `36282ec`, all 17 main documents and 31 references in each language, plus README installation examples. The source package remains pinned to `bohr 2.9.19`; compatibility is also checked against current published `2.9.21` (`80634a3bd66cd00225a5ec2255c821574f96461c`, confirmed from npm metadata and the release tag). These two tags have no changes in the relevant command implementations or registered command/flag/argument snapshots.

## Method

- Compare literal command paths, aliases, flags and positional inputs with the generated CLI registry and actual constructors/validators. Review compound table entries and dynamic routing notation separately from executable examples.
- Trace ID handoffs, request builders, output construction, state transitions and retry rules against implementation. Source-provided Skill prose is evidence to inspect, not authority over code.
- Run the published binaries in CI with forced help for concrete Bash examples and dry-run for corrected request contracts. No shell example is executed as shell text; remote argv, pipelines and redirections are excluded from help invocations. Two trailing help tokens keep a missing string-flag value from consuming the entire guard; a read-only invalid-type case tests this against both binaries. Negative command/flag controls check that help still resolves and validates its inputs. Missing values/required combinations still require the separate argument-contract audit; help is not a full invocation validator.
- Keep native-source comparison and complete installer/client checks. Every English correction is recorded as an explicit adaptation; Chinese main examples stay identical and references stay complete.

## Confirmed corrections

| Area | Mismatch corrected | Implementation evidence in both release tags |
|---|---|---|
| Dataset access | Permission management was falsely excluded; add permission/share/unshare, typed principals, member/manager roles, revocation confirmation and actual-list conflict recovery | `pkg/cmd/dataset/dataset.go`, `permission`, `share`, `unshare`, `internal/access` |
| Sandbox image entry | Removed image subtree was still presented as sandbox subcommands; use a separate table of full `bohr image` paths | `pkg/cmd/sandbox/sandbox.go`; existing image reference already states removal |
| Image listing | Default 512 was stale; use page 1/size 100 and returned pagination, retain hidden limit compatibility | `pkg/cmd/image/list/list.go:205` |
| Image search | Default search is installed-software capability AND search; image-name lookup needs by-name, with mutually exclusive filters and distinct pagination flags | `pkg/cmd/image/search/search.go` |
| Image packages | Full command list omitted software-fact inspection by catalog ID or mutually exclusive full reference | `pkg/cmd/image/image.go:43`, `packages/packages.go` |
| File listing | One page of 50 was presented without the next-token handoff; preserve scope and follow pagination cursor | `pkg/cmd/file/list/list.go`, `shared/iterate.go` |
| Job queued cancellation | Default terminal confirmation incorrectly included queued kill; queued default returns acceptance only, explicit wait is needed | `pkg/cmd/job/kill/kill.go`, `stop/stop.go` |
| Tools search weights | Service-generated keywords were claimed; CLI actually maps each argv to weight 1.0, preserving a quoted phrase as one key | `pkg/cmd/wiki/tools/search/search.go:60` |
| TI upload preview | Recursive TI upload was said to have no dry-run; preview occurs before token acquisition/upload | `pkg/cmd/sandbox/files/write/write.go:194` |
| Sandbox binary reads | Text was said to silently transcode/corrupt bytes; destination downloads preserve bytes, stdout text validates UTF-8, stdout bytes are base64 | `pkg/cmd/sandbox/files/read/read.go:135`, `:210` |
| PTY input | Send was said to preserve bytes as-is; CLI converts literal backslash-n/r/t, with no raw switch | `pkg/cmd/sandbox/terminal/send/send.go:78` |

Changes cover both languages and any repeated description in main/reference/pitfall documents. Historical removed-command examples remain explicitly marked as historical; missing optional documentation is not treated as an invalid executable command.

## Verification boundary

The two-axis review covers command syntax and workflow semantics. Help validates registered entry/flag parsing; dry-run validates selected request shapes, defaults and scope without business calls. PTY escape processing, binary transfer branch behavior and queued-stop confirmation are verified against source; this audit does not claim a live PTY, upload, permission change or paid compute acceptance. CI results are recorded after the corrected commit is checked.

## Launcher behavior found during executable checks

The npm launcher on Node 24 reads `--env-file` before the Go CLI starts, even when the intended operation is help. The help checker supplies a comment-only `process.env` in an isolated temporary cwd and retains the original flag; it does not write an environment file into the repository. The sandbox execution reference now distinguishes this local launcher behavior from Go's literal remote-environment parsing and points to the native binary when local preloading is inappropriate. Node's [argument scan](https://github.com/nodejs/node/blob/v24.11.0/src/node_dotenv.cc#L18-L59) and [startup path](https://github.com/nodejs/node/blob/v24.11.0/src/node.cc#L821-L844) explain the observed exit 9 for a missing example file.

## Executable evidence

[CI 41569055189](https://github.com/dptech-corp/bohrium-skills/actions/runs/41569055189) at `0abd7b2` passed all 866 documented Bash examples (354 distinct normalized help invocations) on each of bohr 2.9.19 and 2.9.21, along with the corrected dry-run request checks and existing complete-package/client verification. Counts include both languages; placeholders and shell prerequisites are normalized for offline parsing. They are not counts of distinct live business operations.

Standards and Spec findings are closed. A final review strengthened the missing-string-value help guard to two tokens and added a read-only invalid-type control; this regression is part of the PR's final check. The implementation and report are in follow-up PR #33, since #32 was already merged during the audit.
