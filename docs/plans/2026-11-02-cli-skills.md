# CLI-backed Skills Implementation Plan

> **For agentic workers:** Execute inline through the host-native edit → remote/CI verification → self-review workflow. Preserve unrelated work. No sub-agent dispatch is required.

**Goal:** Convert the existing bilingual public Skills to the published `bohr` CLI.

**Architecture:** Keep the existing Skill names and install paths. Route each Skill to its installed, version-matched `bohr guide` and domain help. Remove the old direct-API clients.

**Tech Stack:** Markdown, existing Go embed/installer, GitHub Actions, published npm CLI.

**Spec:** `docs/specs/2026-11-02-cli-skills.md`

## Global Constraints

- Exactly 17 Skills per language; preserve `bohrium-skills@bohrium` and marketplace paths.
- Use published CLI 2.9.19 for verification; install no dependencies or run tests on the editing workstation.
- No business API calls, paid operations, CLI source changes, npm/tag release, or marketplace submission.
- Keep secrets and machine-specific development details out of the public repository.

### Task 1: CLI content and distribution contract

**Files:** `zh/*/SKILL.md`, `en/*/SKILL.md`, `bundle_cli_test.go`, `.claude-plugin/marketplace.json`, `README.md`, `README_EN.md`, `docs/cli.md`, `docs/api/README.md`, `tests/VERIFICATION_REPORT.md`, `.github/workflows/check.yml`.

**Consumes:** Existing `OfficialSkills(lang) ([]Skill, error)`, `EmbeddedFS`, and published CLI guide/help commands.

**Produces:** The same installed names with CLI-only instructions, current documentation, and regression checks for the actual distributed content.

- [x] Add Go checks for all 17 names, bilingual mappings, guide references, absence of API clients in bundled files, and marketplace path resolution. Runtime checks invoke the binary supplied through `BOHR_SKILLS_TEST_BIN` and inspect its returned guide metadata and text.
- [x] Rewrite all 34 Skills. Example guide sequence for paper search:

```bash
bohr --version
bohr guide read bohr-paper-search -o human
bohr guide list bohr-paper-search/references -o json
bohr guide read bohr-paper-search references/commands.md -o human
bohr paper --help
```

- [x] Remove every non-`SKILL.md` file under the 34 Skill directories, `requirements.txt`, `tests/smoke_test.py`, and `tests/run_all_demos.py`. Mark retained API verification evidence as historical. Keep the separately maintained OpenAPI schema and its distribution path unchanged.
- [x] Update bilingual README installation, login, guide/reference use, catalog mapping, dynamic pricing, and migration/released-installer caveats. Keep marketplace names/paths and package/release versions unchanged.
- [x] Add CI on push/pull request. Use Go 1.23 and Node 24, install `@dptech-corp/bohr-cli@2.9.19` under the job's temporary prefix, then run:

```bash
go test ./...
go vet ./...
```

Use the existing installer test seam `run` with `install --home <temporary-directory> --lang zh|en --json`; inspect its JSON receipt and every installed Skill in all three target directories.
- [ ] Review final diff and run text/whitespace checks. Record remote/CI results and any unverified boundary. Deliver the task branch without triggering a tag/npm release.

## Self-review

The single task covers spec acceptance criteria 1–6. Runtime tests validate the actual published CLI, while installer tests cover the embedded content. No package/version bump or cross-repository implementation is needed.
