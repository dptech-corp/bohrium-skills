# Skill distribution downstream correctness

## Goal

The CLI-native packages remain complete during installation and updates and are accepted by their documented downstream consumers. Instructions reflect the published CLI implementation rather than treating native documentation as authoritative when it contradicts code.

## Evidence

- `RealRunner.DetectInstallMethod` treats any executable path containing `bohrium-skills-cli` as npm; the documented manual binary has that name. `syncForVersion` selects the updated child through PATH without checking its version.
- `syncer.Sync` uses a second-resolution backup directory; `copyDir` deletes a same-name backup. npm updates can trigger postinstall sync and a second explicit sync within one second.
- `syncOneSkill` removes the active tree before renaming the staged tree. Partial deletion can leave a broken installation; rollback errors are discarded.
- The published bohr 2.9.19 paper and patent commands register `--page` but omit it from outgoing bodies. Increasing it is not supported pagination.
- Local installed client baselines are Codex 0.164.0, Claude Code 2.1.309, and DSH 0.1.5-rc.2. Codex sources parse SKILL.md and agents/openai.yaml and search both user roots; DSH's published filesystem provider parses the installed bundle and supplies a resource base.

## Decisions

- Keep the existing public commands and release authority. Fix the public repository's installer/update implementation; do not edit the separate Bohrium CLI or consumer repositories.
- Classify npm only from the actual package layout and package identity. A manual executable name does not imply npm. After npm installation, resolve the installed package's binary explicitly and verify the requested version before synchronizing Skills. Never label an older child snapshot as the requested new version.
- During self-update, suppress the postinstall sync and perform one verified sync. Each synchronization owns a unique backup directory, even with a fixed clock. Retain the original tree and user additions in backups.
- Publish complete staged trees by moving the prior tree aside, then promoting the new tree. Publication failure restores the old tree. If restoration fails, preserve the previous tree and report both errors and its recovery location. Cleanup failure must not erase or misidentify the complete active tree.
- For bohr 2.9.19 paper/patent queries, document that page is not forwarded. Repeated queries may be billable and must not be described as expanding page coverage. Verify page 1/2 dry-run bodies and correct both language packages and source adaptations.
- Exercise the real published Codex local skills/list path, DSH filesystem provider list/get, and Claude's offline plugin validation/management where available. Isolate homes/configuration, use no credentials, and make no model calls. A consumer's unavailable/closed-source boundary is a stated limit rather than an inferred pass.
- Default full-set installation targets `~/.agents/skills` and `~/.claude/skills`. Back up and retire selected official legacy `~/.codex/skills` copies only after canonical installation, preserving unrelated custom Skills. Explicit custom target lists retain their requested destinations.
- Restrict the installer npm package to runtime files so CI-only check scripts are not shipped.

## Acceptance criteria

1. Manual and npm layouts are classified correctly, including misleading names and symlinked launchers.
2. PATH shadowing or a wrong-version child cannot silently synchronize an old snapshot as a new release. Successful npm update, including same-version force reinstall, synchronizes the verified fresh binary's content.
3. Same-clock consecutive syncs retain distinct intact backups, including original API files and user additions.
4. Replacement failures leave a complete old or complete new tree; rollback failure preserves recovery data and both errors.
5. Published CLI dry-runs confirm the page limitation, and both main/reference documents warn against fake pagination.
6. Downstream loaders can identify each selected public Skill, preserve interface identity/resource paths, and load full main documents after supported installation routes. Duplicate/override behavior is checked and material integration defects are corrected without overwriting unrelated user Skills.
7. Go regression tests, native source/package checks, consumer integration checks, vet, formatting and npm packing inspection pass in CI. No Mac tests/install/build or production/model call is used.

## Scope and limits

This Standard follow-up is authorized by the request to inspect downstream code and fix migration-related problems. It keeps the existing release workflow and does not publish, merge main, change vendor code, or introduce an independent updater service. The existing release version/package version consistency issue is recorded as a publication prerequisite; no release is claimed here.
