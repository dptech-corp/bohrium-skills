# LKM-style workflow implementation

Spec: [task workflow requirements](../specs/2026-11-02-task-workflows.md).

- [x] Inspect the native LKM intent/ID/async/result model and the other 16 main documents; pin the published implementation to v2.9.19.
- [x] Trace actual output-to-command handoffs for compute/storage and research domains; distinguish raw envelopes from internal Go models.
- [x] Add task routes and one concrete recipe per domain in both languages, using existing detailed references for specialized branches.
- [x] Record explicit source adaptations and extend representative offline command-contract checks.
- [ ] Review the final content against implementation contracts and run the existing full CI, including real client consumption, on the resulting commit.

No runtime business calls, local test/build execution, release or private CLI edits are included.

## Implementation evidence

The contract review uses private CLI source at tag `v2.9.19`, read only. Key boundaries are implemented in `pkg/cmd/job/submit/submit.go` (config `out_files` → request `outFiles`, original success envelope), job describe/download and file transfer (Bohr ID and local receipt); `pkg/cmd/node/{create,get}` (requested node versus machine identity); `pkg/cmd/image` and sandbox image commands (local push versus remote build status); `pkg/cmd/dataset` (ready version and verified archive); `pkg/cmd/file` (scope/transfer receipts); and `pkg/cmd/kb/{upload,document}` (submitted versus indexed versus retrieved). Research commands that print the original upstream envelope do not gain a fixed JSON schema from human-renderer structs.

Public source corrections found while composing workflows: Tools free listing returns taxonomy; Web JSON preserves multiple upstream containers; PDF JSON can omit the handoff token; Mentor success does not check answer non-emptiness or a server completion event. The recipes preserve these limitations rather than inventing fields or retries.
