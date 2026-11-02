# Native CLI Skill packages for public distribution

## Goal

The existing 17 public Bohrium Skills include the CLI-native task workflows, local references, and available agent interface metadata. An installed Skill can guide its domain task directly from its package.

## Current evidence

- The public repository preserves 17 `bohrium-*` names in `zh/` and `en/`, plus `bohrium-skills@bohrium` for Claude.
- The published `@dptech-corp/bohr-cli@2.9.19` embeds full Skill trees: 17 selected main documents, 31 reference documents, and 16 `agents/openai.yaml` files. Sandbox does not include agent interface metadata in this release.
- CLI guide reads expose each file's content, canonical guide name, version, and path; guide directory listings enumerate the embedded tree.
- The current public branch contains standalone routing instructions and no local references. Existing installer checks assume one file per Skill and must account for complete packages.

## Decisions

- Keep the existing 17 public names, both language directories, marketplace names/paths, and installer commands. The selected guide mapping and source version live in `cli-skills.lock.json`.
- Use the published CLI artifact as the authoritative source. English main documents and all reference files retain the complete upstream content. CI compares the distributed files with the real published CLI, not an unshipped source branch.
- Make explicit distribution adaptations: rename the frontmatter Skill name, rename `$skill` references in agent default prompts, add a short CLI setup/compatibility note, and replace the environment-specific sandbox proxy host with a documented placeholder. Preserve command names, arguments, result fields, costs, failure conditions, and lifecycle rules. Explicitly record source-document corrections and implementation-verified task recipes in the lock; the [task workflow requirements](2026-11-02-task-workflows.md) govern the latter. Corrections include the published Wenyon instance-mount interface, supported output formats, paper pagination limitation, and separately authorized persistent confirmation changes.
- Translate Chinese main documents completely. Preserve the English CLI reference documents in both language packages and document that choice. Keep fenced command examples identical between English and Chinese main documents. Localize existing agent UI metadata while preserving the public Skill identifier.
- Prefer local `references/` links in installed packages. `bohr guide` remains available for version inspection and cross-guide lookup, with JSON for agent calls and Markdown in `data.content`.
- Keep the separate CLI source, installer package version, tag/release workflow, OpenAPI schema, and existing published releases unchanged. Public installer/update corrections and canonical default directories follow [the downstream correctness spec](2026-11-02-skill-downstream.md).

## Behavior scenarios

- For PDF parsing, the installed main document supplies local-file/URL selection, pricing authorization, submission flags, `data.token`, bounded polling, terminal/failure states, and result-shape options. Its detailed reference resolves locally.
- For paper/patent search, the package supplies command examples, tier selection, per-page spending limits, payment-source behavior, and evidence-quality boundaries.
- For sandboxes, the package includes lifecycle, execution, files, network, storage, templates, and recovery references. Unknown create outcomes and credentials retain their upstream safeguards. Public network examples require an environment-approved proxy rather than a company-specific endpoint.

## Acceptance criteria

1. Both languages retain the same 17 public names. All 128 selected package files are present: 64 per language, including source-provided references and agent metadata.
2. Every English package is byte-identical to the published CLI after the documented deterministic distribution adaptations. Both languages contain exactly the CLI's file inventory.
3. Every Chinese main document includes its domain's complete native workflow, with identical fenced examples and working local links. Source references are complete and unabridged except for the documented public proxy adaptation.
4. Source-provided `agents/openai.yaml` files identify the public Skill in default prompts; sandbox retains the source's absence of this optional file.
5. Both existing full-set installation targets and published `bohr skill install` preserve every package file byte for byte in isolated destinations.
6. CI checks source/version parity, package inventories, local links, fenced examples, key PDF/search/sandbox workflow boundaries, command help, Go tests, vet, and formatting against the actual published CLI.

## Scope and verification limits

This Standard migration follows the explicit request to use CLI-native Skills and include their complete resources. It adds no new public CLI command, marketplace listing, scheduler, release pipeline, or npm/tag release. Verification runs in CI; the editing workstation is used for source inspection, editing, and synchronization. CLI source files are inspected only at the published version, and CI's real artifact comparison decides parity. Authenticated business flows and paid operations are separate from package/content verification and are not claimed by passing CI.
