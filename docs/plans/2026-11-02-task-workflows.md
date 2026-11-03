# LKM-style workflow implementation

Spec: [task workflow requirements](../specs/2026-11-02-task-workflows.md).

- [x] Inspect the native LKM intent/ID/async/result model and the other 16 main documents; pin the published implementation to v2.9.19.
- [x] Trace actual output-to-command handoffs for compute/storage and research domains; distinguish raw envelopes from internal Go models.
- [x] Add task routes and one concrete recipe per domain in both languages, using existing detailed references for specialized branches.
- [x] Record explicit source adaptations and extend representative offline command-contract checks.
- [x] Review the final content against implementation contracts and run the existing full CI, including real client consumption, on the resulting commit.

No runtime business calls, local test/build execution, release or private CLI edits are included.

## Implementation evidence

The contract review uses private CLI source at tag `v2.9.19`, read only. Key boundaries are implemented in `pkg/cmd/job/submit/submit.go` (config `out_files` → request `outFiles`, original success envelope), job describe/download and file transfer (Bohr ID and local receipt); `pkg/cmd/node/{create,get}` (requested node versus machine identity); `pkg/cmd/image` and sandbox image commands (local push versus remote build status); `pkg/cmd/dataset` (ready version and verified archive); `pkg/cmd/file` (scope/transfer receipts); and `pkg/cmd/kb/{upload,document}` (submitted versus indexed versus retrieved). Research commands that print the original upstream envelope do not gain a fixed JSON schema from human-renderer structs.

Public source corrections found while composing workflows: Tools free listing returns taxonomy; Web JSON preserves multiple upstream containers; PDF JSON can omit the handoff token; Mentor success does not check answer non-emptiness or a server completion event. The recipes preserve these limitations rather than inventing fields or retries.

## Review and initial verification

- Standards/contract review of the eight compute/storage domains found no actionable contract error. Node readiness waiting now has an explicit budget and retains its ID on timeout.
- Spec review of the eight research domains found no P1/P2 and one P3: new tool examples used a deprecated top-level alias. Examples, navigation, mapping and the earlier source adaptation now consistently use canonical `bohr wiki tools`.
- CI [41438837650](https://github.com/dptech-corp/bohrium-skills/actions/runs/41438837650) passed Go/Node/Python tests, native package parity and all three real consumer checks. The new Job preview exposed an incorrect checker expectation: result directories normalize to a trailing slash. The assertion now checks the actual normalized `/personal/results/` while retaining exact output-file, log-file, project and input-tree checks. The full check is rerun on the corrected content.

## Final verification

- Content/checker commit `c4fc48d1d34973d18ab5da01352a05773dfbf0b1` passed the full existing [check workflow](https://github.com/dptech-corp/bohrium-skills/actions/runs/41439151350): Go tests, three Node tests, five Python checker tests, all 17 source package comparisons, bilingual links/examples, isolated complete installation, the new published-CLI Job/Sandbox previews, real Codex/Claude/DSH consumption, npm pack inspection, vet and formatting.
- Standards review has no remaining findings. Spec review confirmed the tool-entry P3 is closed; no remaining findings. Both languages retain the source inventory and all 16 task additions; LKM itself remains unchanged.
- No authenticated business or paid task was executed. The branch is pushed for review; main merge and release are separate.
