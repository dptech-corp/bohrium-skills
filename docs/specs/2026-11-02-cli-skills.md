# CLI-backed Bohrium Skills

## Goal

Agents installing this repository use the published `bohr` CLI for Bohrium operations and read guidance from the installed CLI version.

## Current evidence

- At `fc4ee753c55fc0fde339c85c5fc5b630d2f6d4fb`, `zh/` and `en/` each contain 17 named Skills. Most describe direct HTTP calls and some bundle Python clients.
- `bundle.go` embeds both language trees; the existing installer discovers `bohrium-*` directories and validates frontmatter names. Its tests expect 17 Skills per language.
- `.claude-plugin/marketplace.json` installs the English Skills as `bohrium-skills@bohrium`.
- The published `@dptech-corp/bohr-cli@2.9.19` provides `bohr guide list`, `bohr guide read`, and domain command help.

## Behavior scenarios

- An agent receiving a paper-search task checks `bohr`, reads `bohr-paper-search` with `bohr guide read`, and follows its pricing and command instructions before searching.
- An agent without `bohr` provides the npm installation command and waits for installation authorization. Interactive login is completed by the user; account credentials are handled by the CLI.
- When the installed CLI lacks a guide or command, the agent reports the version and missing capability, then offers an upgrade. It does not reconstruct the API client.

## Decisions

- Keep all existing 17 directory/frontmatter names in each language and the Claude marketplace/plugin names and paths.
- Each standalone Skill contains a domain-specific trigger, setup steps, an exact embedded-guide name, command-help entry points, and execution/result boundaries. Detailed command/reference content remains in `bohr`.
- Require the latest stable CLI and record 2.9.19 as the verified baseline. Guide reads explicitly use human output to emit Markdown; directory listing uses JSON.
- Distribute only `SKILL.md` within each Skill directory. Retire the direct-API Python helpers, their example payloads, Python dependency file, and API smoke/demo runners. Keep dated API verification evidence clearly marked as historical. Preserve the separately maintained OpenAPI schema and its published path; clarify that CLI Skills no longer serve as its interface documentation.
- Keep installer source, package version, tag/release workflow, and published releases unchanged. The npm installer embeds its Skill content at release time; changing GitHub does not update previously released installers.

## Acceptance criteria

1. Both language trees expose the same existing 17 names and resolve to guides available in the published CLI.
2. Every published Skill has CLI setup and domain-specific help; none contains raw HTTP/API fallback, credential-copy instructions, legacy CLI installation, or bundled API scripts.
3. Every guide/reference and help command used by a Skill succeeds offline with the verified CLI. Guide references are read via `bohr guide read`, not relative filesystem paths.
4. The existing installer still installs all 17 Skills into isolated destinations. Claude marketplace entries still resolve to their English directories.
5. README files describe CLI-backed Skills, installation, login, dynamic pricing, current `guide` commands, migration effects, and the distinction between GitHub content and released installer snapshots.
6. Existing Go tests and the new distribution checks pass in a remote development or CI environment. No business API or paid operation is required for these checks.

## Scope and non-goals

This is a Standard content/distribution migration authorized by the request to convert the GitHub repository to the CLI version. It does not change public CLI commands, add a generator, expand the Skill catalog, publish npm/tag releases, or submit marketplace listings.

## Verification

Run `go test ./...` and `go vet ./...` in CI. Install the exact published CLI in a job-owned prefix and run runtime guide/help checks plus an isolated installer smoke test. Inspect the final diff, bilingual mapping, removed resources, and public documentation links. Business API behavior and authenticated/paid flows are outside this migration's verification.
