# Native CLI Skill Distribution Implementation Plan

> **For agentic workers:** Execute inline using the host-native edit → CI verification → self-review workflow. Preserve unrelated work and use the existing task branch.

**Goal:** Distribute complete CLI-native Skill packages under the existing public names.

**Architecture:** Import all files for the selected 17 guides from the published CLI snapshot. Keep local references and agent metadata, adapt public identities, and translate Chinese main documents without changing examples. Verify against the published binary in CI.

**Tech Stack:** Markdown, YAML metadata, JSON source lock, existing Go embed/installer, GitHub Actions.

**Spec:** `docs/specs/2026-11-02-cli-skills.md`

## Global constraints

- Source package is `@dptech-corp/bohr-cli@2.9.19`; gitHead is `efb5ba5707427708c2a82b794598868160a87ed0`.
- Exactly 17 public Skills per language; preserve `bohrium-skills@bohrium` and all existing public names.
- Full inventory is 64 files per language: 17 main documents, 31 reference documents, and 16 agent metadata files.
- Public adaptations preserve CLI behavior and replace only Skill identity, setup/compatibility text, and the environment-specific proxy hostname.
- All tests/installations run in CI. No authenticated business/paid operation, CLI source edit, npm/tag release, or marketplace submission.

### Task 1: Package content and localization

**Files:** `.gitignore`, `cli-skills.lock.json`, `en/bohrium-*/**`, `zh/bohrium-*/**`, `README.md`, `README_EN.md`, `docs/cli.md`.

**Consumes:** The selected published guide names and complete source trees. `bohr guide list <guide[/directory]> -o json` returns `data.items`; `bohr guide read <guide> <path> -o json` returns the source text in `data.content`.

**Produces:** Complete locally readable Skill packages, a pinned source mapping, and accurate installation/update documentation.

- [x] Record the 17 public-name/canonical-guide/command mappings and published source version in `cli-skills.lock.json`.
- [x] Copy the complete English source trees. Change only `name:` and agent `$skill` identities, add the common runtime setup note, and replace the company-specific proxy host with `proxy.example.com` plus a clear placeholder instruction in the network reference.
- [x] Translate all 17 Chinese main documents. Preserve source examples verbatim, copy complete reference documents, and localize the 16 existing agent interface files. Keep sandbox's source inventory unchanged.
- [x] Update both README files and the installer guide to describe packaged workflows, local references, source version, localization, source-update verification, and old published installer snapshots.

### Task 2: Native-source and installation regression checks

**Files:** `bundle_cli_test.go`, `cmd/bohrium-skills-cli/main_test.go`, `scripts/check_native_skills.py`, `scripts/check_native_skills_test.py`, `.github/workflows/check.yml`.

**Consumes:** `OfficialSkills`, `EmbeddedFS`, the source lock, and the installed published CLI supplied through `BOHR_SKILLS_TEST_BIN`.

**Produces:** Proof that the distributed trees preserve source content and install complete resources.

- [x] Replace single-file/routing-only checks with package inventory, valid local-link, source workflow, and bilingual fenced-example checks.
- [x] Recursively list/read all selected trees through the real published CLI. Compare inventories, full English adapted files, shared references, source-provided metadata identities, and guide metadata. Retain default JSON, command-help, and isolated `bohr skill install` checks.
- [x] Expand the existing full-set installer test to compare every source file recursively for both languages and all three destinations; preserve missing/extra-file detection.
- [ ] Run `go test ./... -count=1`, `python3 -m unittest discover -s scripts -p '*_test.py'`, `python3 scripts/check_native_skills.py --cli "$BOHR_SKILLS_TEST_BIN"`, `go vet ./...`, and formatting checks through the existing `check` workflow, installing the CLI version selected by the source lock.
- [ ] Inspect the final diff and CI result. Report the actual tested boundary and leave main/npm/tag publication as separate steps.

## Self-review

Task 1 implements spec acceptance criteria 1–4; Task 2 verifies criteria 1–6. The public package is self-contained for instructions and references while `bohr` remains the runtime dependency. The existing release flow is retained and no independent update/release service is introduced.
