---
name: bohrium-setup
description: >-
  Install and sign in to the Bohrium CLI (`bohr`), verify with `bohr doctor`,
  then list available guides. Use when onboarding after installing the Bohrium
  plugin or skills, when `bohr` is missing or outdated, when auth fails, or
  before the first paid/read-write Bohrium command.
compatibility: "Installs or upgrades npm package @dptech-corp/bohr-cli; verified baseline 2.9.19. Sign-in needs network and a browser or device-code flow."
---

# Bohrium Setup

Run this skill after installing the Bohrium plugin/skills, or whenever `bohr` is missing, outdated, or not signed in. Other Bohrium skills assume a working CLI.

## Goal

1. Confirm `bohr` is on `PATH` at a supported version.
2. Sign the user in.
3. Pass `bohr doctor`.
4. Point the next task at the right skill / `bohr guide`.

## Workflow

### 1. Check the CLI

```bash
bohr --version
```

- If the command is found and the version is **≥ 2.9.19**, skip to step 3 unless the user asked to reinstall.
- If the command is missing, or the version is older than 2.9.19, install or upgrade (step 2).
- Never invent a version string; only trust command output.

### 2. Install or upgrade

Prefer npm (official package). On macOS/Linux, Homebrew is fine when a `dptech-corp` formula is available; otherwise use npm.

```bash
npm install -g @dptech-corp/bohr-cli@latest
bohr --version
```

If npm is unavailable, tell the user which package manager is missing and stop; do not silently switch to an unofficial binary.

### 3. Sign in

```bash
bohr auth login
```

- Interactive desktops: follow the CLI browser flow.
- SSH / headless / agent environments: check `bohr auth login --help` and prefer device code (`--device --no-wait`, then `--resume`) when those flags exist.
- Do **not** put access keys, tokens, or passwords into Skills, the repo, or chat. Use the CLI’s own credential store.

Confirm:

```bash
bohr auth status
```

### 4. Doctor

```bash
bohr doctor
```

Fix anything the doctor reports (PATH, auth, network) before running job/file/billing commands.

### 5. Hand off to the right skill

```bash
bohr guide list
```

Then map the user’s task to a Bohrium skill (for example `bohrium-job`, `bohrium-paper-search`, `bohrium-sandbox`) or `bohr guide read <name>`.

## Safety and billing

- Prefer `--dry-run` on commands that create or spend before a real run.
- Commands that charge or mutate usually need an explicit `--yes` (or equivalent); never pass `--yes` unless the user clearly approved the spend or change.
- Missing price fields in a dry-run do **not** mean the real call is free — confirm with `bohr billing pricing` when cost matters.

## Failure checklist

| Symptom | What to try |
|---|---|
| `bohr: command not found` | Re-run npm global install; ensure npm’s global bin dir is on `PATH` |
| Auth loop / not logged in | `bohr auth login` again; for headless, device-code + `--resume` |
| Doctor network errors | Check proxy/VPN; company networks may need HTTP(S)_PROXY |
| Old CLI, new skill flags | Upgrade with `npm install -g @dptech-corp/bohr-cli@latest` |
