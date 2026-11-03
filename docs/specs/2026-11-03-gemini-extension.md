# Gemini CLI extension

## Goal

Let Gemini CLI users install the existing Bohrium Skills from the public GitHub repository and discover the extension through the Gemini CLI gallery.

## Current evidence

- `skills/` contains 17 links to the English packages under `en/` and the `bohrium-setup` directory. `plugin.json` identifies the existing bundle as version `0.1.1`.
- Gemini's extension format uses a root `gemini-extension.json` and discovers skills under `skills/`. Its gallery indexes public GitHub repositories with the `gemini-cli-extension` topic after validation.
- A development-host experiment with Gemini CLI 0.66.0 validated a minimal manifest and listed all 18 skills after local installation. A Git installation of the same commit through an isolated local mirror exposed that Gemini rewrites the 17 symlinks to its temporary clone, which it then deletes. PR #35 replaces these links with real, source-matched directories.

## Behavior scenario

Given Gemini CLI and Git, a user installs the extension using the public GitHub URL. Gemini lists the 17 English Bohrium Skills and setup skill, including their local reference files. The user separately installs and authenticates `bohr` before using a Bohrium workflow.

## Requirements

- Add a root Gemini extension manifest identifying `bohrium-skills`, with its version aligned with `plugin.json`.
- Document the GitHub installation command and `bohr` prerequisite in both README languages.
- Preserve existing Skill names, plugin manifests and installation paths. PR #35 provides the real directories required for working Git installation.

## Decisions

Reuse the existing English Skill contents and setup skill. The extension contains no new MCP server, hooks or credentials. Publish from the existing public repository only after PR #35 and this manifest reach the default branch; gallery indexing is an external asynchronous step.

## Acceptance criteria

- Gemini validates the extension manifest and installs a specific pushed GitHub branch in an isolated configuration directory.
- The installed extension lists all 18 expected Bohrium skills and its installed paper-search Skill and references remain readable independently of the source checkout.
- README instructions identify the correct repository installation command and the separate `bohr` prerequisite.
- Any unverified gallery indexing, model invocation or business execution is reported separately.

## Scope and dependencies

The GitHub repository, PR #35 and Gemini CLI are dependencies. Model authentication, real business operations, Windows installation and Gemini gallery acceptance are outside the local installation check and require separate evidence. No CLI, backend or WorkBuddy changes are included.
