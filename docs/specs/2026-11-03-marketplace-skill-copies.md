# Marketplace Skill copies

## Goal

Distribute the 17 English Bohrium Skills as complete ordinary directories so Git-based consumers can install them without broken links, with the current source files kept authoritative.

## Current evidence

- The public `skills/bohrium-*` paths are symbolic links to `en/` on main. PR #35 replaces them with directories for Grok.
- Gemini CLI 0.66.0 Git installation rewrites these links to a temporary clone, leaving only `bohrium-setup` usable after the clone is removed.
- Six PR #35 Skill copies differ from the current `en/` source. `TestClaudeMarketplaceResolvesPublishedSkills` also rejects the valid setup path on both main and PR #35.

## Requirements and decisions

- Each of the 17 `skills/` copies contains exactly the complete regular files of its matching `en/` directory and is not a symlink. `en/` remains the source of truth; setup remains independent.
- The Claude marketplace contains each of the 17 English Skills and exactly one existing, valid `skills/bohrium-setup` entry. Unknown, missing and duplicate entries still fail the check.
- A Go test running in the existing `go test ./...` CI step checks complete directory inventories and bytes, including reference files, and rejects symlinks so drift or broken packaging cannot pass silently.
- Keep existing public Skill names, references, Claude/Codex manifests and the separate CLI unchanged.

## Acceptance criteria

- Before refreshing, the new parity check reports the observed stale files; after refreshing all 17 trees it succeeds.
- The marketplace check accepts setup and still requires each expected entry exactly once.
- Development-host tests and PR CI report results independently. A Gemini consumer test using this change with the Gemini manifest discovers all 18 Skills after installation and retains their references.

## Scope and dependencies

This PR fixes the existing copied Skill distribution and marketplace validation. Gemini manifest and catalog listing remain in the separate Gemini branch; successful model or business calls are outside these checks.
