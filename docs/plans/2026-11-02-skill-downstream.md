# Downstream Skill correctness implementation plan

> **For agentic workers:** Use the native inline test-first → CI → fix → review flow; preserve other work. Review agents are read-only.

**Goal:** Correct the observed update/backup/publication defects and verify actual consumer loading.

**Architecture:** Keep the existing embed, sync and update modules. Validate installation provenance and updated binary version, stage complete trees with recoverable replacement, and run real vendor parser/discovery paths in isolated CI.

**Tech stack:** Existing Go standard library, Node 24, Python standard library, pinned published consumer packages.

**Spec:** `docs/specs/2026-11-02-skill-downstream.md`

## Constraints and seams

- Observe `RealRunner.DetectInstallMethod` from relocated helper processes; `Update` receipts and fresh-child synchronization; `Sync`'s actual active/backup file trees; replacement failure/rollback filesystem outcomes; and real consumer discovery/read responses.
- Run all tests/builds/installations in CI, not on the editing Mac. Source inspection and syntax/text checks are local only.
- Baselines: bohr 2.9.19, Codex 0.164.0, Claude Code 2.1.309, DSH filesystem provider 0.1.5-rc.2. No model/paid operation or credentials.
- Keep public names and commands; no cross-repository writes, tag/npm publication or main merge.

## Tasks

1. Add regressions for misleading manual executable paths, same-clock backups, and partial-delete behavior; push an explicit failing CI run before the fix. Add wrong-version/PATH and failed-publication/rollback cases through process/filesystem seams.
2. Fix npm detection, suppress postinstall's duplicate sync during self-update, resolve and verify the fresh npm binary, and delay updated/synced success receipts until synchronization succeeds.
3. Allocate unique backup directories. Replace destructive active-tree deletion with staging/previous-tree promotion and faithful rollback errors. Verify complete native references/metadata and legacy user edits survive backup/recovery.
4. Correct the paper/patent paging guidance in English/Chinese and both references. Record deterministic source adaptations and compare bohr page 1/2 dry-run bodies without gateway calls.
5. Install pinned consumer tools in a job-owned prefix. Exercise Codex app-server skills/list, DSH's published FileSystemSkillProvider list/get and resource bases, and Claude plugin validation/management. Inspect actual duplicate discovery and preserve user-managed overrides when choosing any fix.
6. Narrow npm runtime file packaging and inspect npm pack output. Run the existing native checks, complete Go regression suite, consumer checks, vet and formatting. Review final diff and close concrete findings.

## Review

The source/read traces resolve the technical choices without a new public interface. Each test asserts an independently observable process, filesystem, request body or consumer result; it does not merely echo the content generator. CI results are attached to each commit. Vendor/model execution and authenticated business acceptance remain separate boundaries.
