---
name: bohrium-knowledge-base
description: "Use the bohr CLI to manage knowledge bases, literature, tags, notes, and retrieval. Trigger on Bohrium knowledge bases tasks; read the installed CLI's bohr-knowledge-base guide for detailed commands and execution boundaries."
---

# Bohrium Knowledge Bases

Use `bohr kb` for this task. The installed CLI's `bohr-knowledge-base` guide owns the detailed workflow.

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
bohr guide read bohr-knowledge-base -o text
bohr guide list bohr-knowledge-base/references -o json
bohr guide read bohr-knowledge-base references/commands.md -o text
bohr kb --help
```

Relative references inside that guide belong to the CLI's embedded resources. Read
`references/…` with `bohr guide read bohr-knowledge-base references/path-to-file -o text`,
rather than looking for it in the local Skill directory. Read other referenced files as the task requires,
and confirm flags through the specific subcommand's `--help`.

## 3. Complete the task

Distinguish knowledge-base management, retrieval within a library, and public paper search. Resolve the library and literature identifiers before upload, organization, or retrieval. Preserve sources and confirm the scope of deletion.

- Before account access, run `bohr auth status -o json`. If signed out, use `bohr auth login` and have the user complete browser authorization. Let the CLI manage credentials; keep credential files and secrets out of output.
- Follow the guide for current prices, payment source, and write effects. Add `--yes` only when the user has authorized the specific spending or target operation. Preview with `--dry-run` where supported and retain CLI confirmation controls.
- Use `-o json` for automated business commands. Check the exit status, `ok`, error details, and business state. Follow asynchronous requests to a verifiable result; report missing fields and unknown states accurately.
- If the CLI lacks a guide or capability, report its version and the missing item and suggest an upgrade. Stay within supported CLI operations rather than constructing HTTP requests or restoring the old Python clients.
