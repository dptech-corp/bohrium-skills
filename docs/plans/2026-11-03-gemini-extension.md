# Gemini CLI Extension Implementation Plan

**Goal:** Make the existing Bohrium Skills installable as a Gemini CLI extension.

**Architecture:** Add the root manifest required by Gemini and reuse the existing English Skills tree. Document installation in both README languages; verify actual consumer behavior using an isolated development-host Gemini configuration.

**Tech Stack:** Gemini CLI, GitHub, JSON, Markdown.

**Spec:** `docs/specs/2026-11-03-gemini-extension.md`

## Global Constraints

- Keep `plugin.json` version `0.1.1`, existing symlinks and skill sources unchanged.
- Do not modify the CLI, backend, other repositories, credentials or existing Gemini configuration.
- Run Gemini validation and installation on the development host in the task-owned directory.

### Task 1: Add manifest and installation instructions

**Files:** Create `gemini-extension.json`; modify `README.md` and `README_EN.md`.

- [ ] Add a manifest with `name` = `bohrium-skills`, `version` = `0.1.1`, and description `Bohrium scientific workflows using the bohr CLI.`
- [ ] In both READMEs document `gemini extensions install https://github.com/dptech-corp/bohrium-skills` and explain that `bohr` must be installed and authenticated separately.
- [ ] On the development host validate the manifest and install from the proposed source; list all 18 Skills, and read the installed paper-search Skill plus references without relying on the source checkout.
- [ ] Inspect the final diff, commit only task files, push the feature branch, and install the exact pushed branch by GitHub URL. Record any consumer or external indexing limitation.
