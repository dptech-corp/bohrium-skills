# Gemini CLI extension

## Goal

Let Gemini CLI users install the existing Bohrium Skills from the public GitHub repository and discover the extension through the Gemini CLI gallery.

## Current evidence

- `skills/` contains 17 links to the English packages under `en/` and the `bohrium-setup` directory. `plugin.json` identifies the existing bundle as version `0.1.1`.
- Gemini's extension format uses a root `gemini-extension.json` and discovers skills under `skills/`. Its gallery indexes public GitHub repositories with the `gemini-cli-extension` topic after validation.
- A development-host experiment with Gemini CLI 0.66.0 validated a minimal manifest and listed all 18 skills after local installation. That installation linked back to its source directory; a GitHub installation remains to be verified.

## Behavior scenario

Given Gemini CLI and Git, a user installs the extension using the public GitHub URL. Gemini lists the 17 English Bohrium Skills and setup skill, including their local reference files. The user separately installs and authenticates `bohr` before using a Bohrium workflow.

## Requirements

- Add a root Gemini extension manifest identifying `bohrium-skills`, with its version aligned with `plugin.json`.
- Document the GitHub installation command and `bohr` prerequisite in both README languages.
- Preserve existing Skills, symbolic links, plugin manifests and existing installation paths.

## Decisions

Reuse the existing English Skills tree and setup skill. The extension contains no new MCP server, hooks or credentials. Publish from the existing public repository once the default branch includes the manifest; gallery indexing is an external asynchronous step.

## Acceptance criteria

- Gemini validates the extension manifest and installs a specific pushed GitHub branch in an isolated configuration directory.
- The installed extension lists all 18 expected Bohrium skills and its installed paper-search Skill and references remain readable independently of the source checkout.
- README instructions identify the correct repository installation command and the separate `bohr` prerequisite.
- Any unverified gallery indexing, model invocation or business execution is reported separately.

## Scope and dependencies

The GitHub repository and Gemini CLI are external dependencies. Model authentication, real business operations, Windows symlink behavior and Gemini gallery acceptance are outside the local installation check and require separate evidence. No CLI, backend or WorkBuddy changes are included.
