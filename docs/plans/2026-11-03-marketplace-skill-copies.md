# Marketplace Skill Copies Implementation Plan

**Goal:** Keep ordinary marketplace Skill directories complete and synchronized with English sources.

**Architecture:** Compare each distributed tree with the existing embedded English source through a Go test, then synchronize the 17 directories. Extend the existing marketplace inventory check to include the setup path.

**Tech Stack:** Go, Git, Gemini CLI.

**Spec:** `docs/specs/2026-11-03-marketplace-skill-copies.md`

## Global Constraints

- The English tree is the source of truth; setup remains independent.
- Tests and consumer installation run in an isolated development-host directory.
- Preserve all 17 existing public names and the Claude and Codex manifests.

### Task 1: Verify and refresh marketplace copies

**Files:** `bundle_cli_test.go`, the 17 `skills/bohrium-*` directories except setup.

- [ ] Add a test that obtains the English inventory via `OfficialSkills("en")`, rejects links and special files, compares each file byte with `EmbeddedFS`, and checks for missing and extra files.
- [ ] Run the new test on the development host and observe it fail on stale copies.
- [ ] Synchronize the 17 trees from their corresponding `en/` directories and rerun the check; retain setup unchanged.
- [ ] Extend the Claude marketplace expected-path set with `./skills/bohrium-setup` and validate its frontmatter name and presence; run the existing marketplace test.
- [ ] Run `go test ./... -count=1` and the repository's CI checks on the development host or use matching CI results. Check the final diff, commit task files, and push PR #35's existing branch.
- [ ] Validate Gemini installation using the exact updated PR contents plus the Gemini manifest and report any network, model or gallery limitations.
