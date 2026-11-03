# Gemini CLI Extension Implementation Plan

**Goal:** Make the existing Bohrium Skills installable as a Gemini CLI extension.

**Architecture:** Add the root manifest required by Gemini and reuse the existing English Skills tree. Document installation in both README languages; verify actual consumer behavior using an isolated development-host Gemini configuration.

**Tech Stack:** Gemini CLI, GitHub, JSON, Markdown.

**Spec:** `docs/specs/2026-11-03-gemini-extension.md`

## Global Constraints

- Keep `plugin.json` version `0.1.1`, Skill names and sources unchanged; PR #35 supplies real directories for Git installation.
- Do not modify the CLI, backend, other repositories, credentials or existing Gemini configuration.
- Run Gemini validation and installation on the development host in the task-owned directory.

### Task 1: Add manifest and installation instructions

**Files:** Create `gemini-extension.json`; modify `README.md` and `README_EN.md`.

- [ ] Add a manifest with `name` = `bohrium-skills`, `version` = `0.1.1`, and description `Bohrium scientific workflows using the bohr CLI.`
- [ ] In both READMEs document `gemini extensions install https://github.com/dptech-corp/bohrium-skills` and explain that `bohr` must be installed and authenticated separately.
- [ ] On the development host validate the manifest and install the combined PR #35 and Gemini content; list all 18 Skills, and read installed Skill files and references independently of the source checkout.
- [ ] Inspect the final diff, commit only task files and push the feature branch. Record GitHub connectivity, consumer and external indexing limitations.
