---
name: bohrium-project
description: "Use the bohr CLI to inspect or manage Bohrium projects, members, and quotas. Trigger on Bohrium projects tasks; read the installed CLI's bohr-project guide for detailed commands and execution boundaries."
---

# Bohrium Projects

Use `bohr project` for this task. The installed CLI's `bohr-project` guide owns the detailed workflow.

## 1. Check the CLI

```bash
bohr --version
```

When installation or upgrade is needed, follow the environment's permissions and the user's instructions:

```bash
npm install -g @dptech-corp/bohr-cli@latest
```

This Skill targets the latest stable CLI; the verified baseline is **2.9.19**.

## 2. Read version-matched guidance

Before business operations, read:

```bash
bohr guide read bohr-project -o json
bohr guide list bohr-project/references -o json
bohr guide read bohr-project references/commands.md -o json
bohr project --help
```

These read commands return JSON. Check `ok=true`, then read the Markdown in `data.content` as the instructions. Use `data.name`, `data.version`, and `data.path` to identify the guide. Directory listings expose `data.items`.

Relative references inside that guide belong to the CLI's embedded resources. Read
`references/…` with `bohr guide read bohr-project references/path-to-file -o json`,
rather than looking for it in the local Skill directory. Read other referenced files as the task requires,
and confirm flags through the specific subcommand's `--help`.

## 3. Complete the task

Resolve the project, membership permissions, and intended change. Apply membership, quota, and deletion changes only to the requested target and read back the result. Report operations absent from the CLI as capability gaps.

- Before account access, run `bohr auth status -o json`. If signed out, use `bohr auth login` and have the user complete browser authorization. Let the CLI manage credentials; keep credential files and secrets out of output.
- Follow the guide for current prices, payment source, and write effects. Add `--yes` only when the user has authorized the specific spending or target operation. Preview with `--dry-run` where supported and retain CLI confirmation controls.
- Use `-o json` for automated business commands. Check the exit status, `ok`, error details, and business state. Follow asynchronous requests to a verifiable result; report missing fields and unknown states accurately.
- If the CLI lacks a guide or capability, report its version and the missing item and suggest an upgrade. Stay within supported CLI operations rather than constructing HTTP requests or restoring the old Python clients.
