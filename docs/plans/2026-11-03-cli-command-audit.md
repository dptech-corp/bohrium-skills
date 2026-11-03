# CLI command and workflow contract audit

Scope: PR #32, public Skills at `36282ec`, all 17 domains and their complete references in both languages. Existing native-package and task-workflow specs require commands, arguments, identifiers and completion rules to match implementation.

- [x] Confirm the pinned source `bohr 2.9.19` and current published `2.9.21` against npm gitHead and release tags. Their relevant command implementations and command/flag/argument snapshots have no changes.
- [x] Inventory literal command paths/flags and review task semantics against actual command registration, validation, request builders and output paths. Keep private CLI checkouts read only.
- [x] Correct 11 confirmed source-document defect classes in both languages and record explicit deterministic source adaptations.
- [ ] Extend offline published-binary checks for documented Bash examples and the corrected request contracts; run on both releases through existing CI.
- [ ] Close Standards and Spec findings, inspect the final diff and record verification evidence.

Validation is offline help/dry-run plus existing isolated package/client checks. It does not submit computation, change permissions, open PTYs, upload user data or spend money. Mac work is editing and static text inspection; executable validation runs in CI.
