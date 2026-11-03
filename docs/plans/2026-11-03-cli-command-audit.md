# CLI command and workflow contract audit

Scope: PR #32, public Skills at `36282ec`, all 17 domains and their complete references in both languages. Existing native-package and task-workflow specs require commands, arguments, identifiers and completion rules to match implementation.

- [x] Confirm the pinned source `bohr 2.9.19` and current published `2.9.21` against npm gitHead and release tags. Their relevant command implementations and command/flag/argument snapshots have no changes.
- [x] Inventory literal command paths/flags and review task semantics against actual command registration, validation, request builders and output paths. Keep private CLI checkouts read only.
- [x] Correct 11 confirmed source-document defect classes in both languages and record explicit deterministic source adaptations.
- [x] Extend offline published-binary checks for documented Bash examples and the corrected request contracts; run on both releases through existing CI.
- [x] Close Standards and Spec findings, inspect the final diff and record verification evidence.

Validation is offline help/dry-run plus existing isolated package/client checks. It does not submit computation, change permissions, open PTYs, upload user data or spend money. Mac work is editing and static text inspection; executable validation runs in CI.

## Verification and review

- PR #32 was merged externally during this audit. The fixes were copied onto `fix/cli-command-contracts`, created from latest `origin/main` at `adf1ee2`, and delivered as PR #33. The old worktree and its commits were preserved.
- [CI 41569055189](https://github.com/dptech-corp/bohrium-skills/actions/runs/41569055189) at `0abd7b2` passed 866 documented examples / 354 distinct normalized help invocations on each of bohr 2.9.19 and 2.9.21, all corrected dry-run probes, complete package comparison/installation, real Codex/Claude/DSH discovery, Go/Node/Python checks, npm content inspection, vet and formatting.
- Initial checker failures were traced to prepended help interfering with multi-level lookup and Node 24 loading the missing example env file before CLI help. Command paths now remain intact and the comment-only env fixture lives in a temporary cwd. No example flag coverage was removed.
- Both review axes have no remaining findings. A final review added double trailing help so a missing string-flag value cannot consume the whole guard, with unit and published-binary read-only controls. The final guard regression is included in the PR's attached check workflow.
