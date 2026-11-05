# Bohrium Skills

Agent Skills for the Bohrium scientific computing platform, powered by the [Bohrium CLI (`bohr`)](https://www.bohrium.com/bohr-cli/intro).

[中文](README.md)

This repository provides 17 complete CLI-native Skill packages in Chinese and English. Main documents contain task workflows, command examples, billing, and recovery rules; local `references/` and source-provided `agents/openai.yaml` files ship with each package. The source is the published `bohr 2.9.19` artifact, recorded in [cli-skills.lock.json](cli-skills.lock.json).

## 1. Install and sign in to the Bohrium CLI

```bash
npm install -g @dptech-corp/bohr-cli@latest
bohr --version
bohr auth login
bohr doctor
```

Supports macOS, Linux, and Windows on x64 / arm64. Installing a Skill and installing `bohr` are separate steps; the Skill does not bundle the CLI. The verified baseline is **2.9.19**; use the latest stable release.

The CLI guides the user through browser authorization. On servers or in non-interactive environments, check `bohr auth login --help` for supported sign-in methods. Provide credentials through the environment's established CLI configuration; keep secrets out of Skills, the repository, and conversations.

## 2. Install a Skill

`bohr skill install` is included in the Bohrium CLI. No separate Skill installer is needed. Choose one method:

### Gemini CLI extension

```bash
gemini extensions install https://github.com/dptech-corp/bohrium-skills
```

The extension provides the English Skills and `bohrium-setup`. Install and sign in to `bohr` separately as described above.

### Claude Code plugin

Install the English Skills:

```text
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Agent Plugins / Codex marketplace

This repo is also an [Agent Plugins 1.0](https://agent-plugins.org/) package (`plugin.json` + `skills/` → English skills). For Codex, add the repo marketplace then install:

```bash
codex plugin marketplace add dptech-corp/bohrium-skills
codex plugin add bohrium-skills@bohrium
```

Or clone locally and point Codex at the repo root (contains `.agents/plugins/marketplace.json`).

After install, run the **bohrium-setup** skill (Codex/ChatGPT onboarding skill) so the agent installs `bohr`, signs in, and runs `bohr doctor` before other Bohrium skills.

### Native DSH plugin

DSH users can install one native bundle containing all 17 workflow Skills and `bohrium-setup`. The independent package in `skills/` reuses the complete content and does not run the legacy installer. Supports **DSH 0.2.x from 0.2.1-rc.5**, including Desktop 0.2.2-rc.1.

See the [DSH guide](docs/dsh.md#english) for installation, index setup, updates, and removal. Install and sign in to `bohr` separately, then run `bohrium-setup`. Users must add the public index themselves: https://dptech-corp.github.io/bohrium-skills/index.json .

### Codex, Claude Code, or DSH

Clone this repository and install the Skill you need, such as English paper search:

```bash
git clone https://github.com/dptech-corp/bohrium-skills.git
cd bohrium-skills
bohr skill install ./en/bohrium-paper-search --target codex --scope user
```

Supported `--target` values are `codex`, `claude`, and `dsh`; replace `en/` with `zh/` for Chinese. Choose one language per Skill and follow the receipt's instructions to reload your agent. Embedded CLI guidance is available directly through `bohr guide`; external agents still need the installation above to discover these Skills.

Then describe your task directly to the agent, for example:

> Find papers published since 2020 on SEI in lithium-metal batteries and organize their titles, years, DOIs, and abstracts. Show me the estimated cost before any paid searches.

See the [detailed guide](docs/cli.md) for package structure, full-set installation, and the older installer.

## Skill catalog

Public Skill names remain `bohrium-*`; embedded CLI guides use `bohr-*`.

| Skill | Command entry point | Embedded guide | Purpose |
|-------|---------------------|----------------|---------|
| [bohrium-setup](skills/bohrium-setup/SKILL.md) | `bohr` install / `auth` / `doctor` | — | Install CLI, sign in, verify, then hand off |
| [bohrium-job](en/bohrium-job/SKILL.md) | `bohr job` | `bohr-job` | Compute jobs, logs, and results |
| [bohrium-node](en/bohrium-node/SKILL.md) | `bohr node` | `bohr-node` | Interactive development nodes |
| [bohrium-dataset](en/bohrium-dataset/SKILL.md) | `bohr dataset` | `bohr-dataset` | Dataset uploads, version lookup, and resumable downloads |
| [bohrium-file](en/bohrium-file/SKILL.md) | `bohr file` | `bohr-file` | Personal/share file storage |
| [bohrium-database](en/bohrium-database/SKILL.md) | `bohr database` | `bohr-database` | Scientific databases and table schemas |
| [bohrium-image](en/bohrium-image/SKILL.md) | `bohr image` | `bohr-image` | Container images, builds, and snapshots |
| [bohrium-project](en/bohrium-project/SKILL.md) | `bohr project` | `bohr-project` | Project discovery, creation/deletion, and member lookup |
| [bohrium-knowledge-base](en/bohrium-knowledge-base/SKILL.md) | `bohr kb` | `bohr-knowledge-base` | Knowledge bases, document upload, and retrieval |
| [bohrium-paper-search](en/bohrium-paper-search/SKILL.md) | `bohr paper` | `bohr-paper-search` | Paper and patent search |
| [bohrium-pdf-parser](en/bohrium-pdf-parser/SKILL.md) | `bohr pdf` | `bohr-pdf-parse` | PDF parsing and result retrieval |
| [bohrium-scholar-search](en/bohrium-scholar-search/SKILL.md) | `bohr scholar` | `bohr-scholar-search` | Scholar search and profiles |
| [bohrium-sciencepedia](en/bohrium-sciencepedia/SKILL.md) | `bohr wiki` | `bohr-wiki` | Scientific articles, courses, and graphs |
| [bohrium-tools](en/bohrium-tools/SKILL.md) | `bohr wiki tools` | `bohr-tools` | Scientific tool discovery and details |
| [bohrium-web-search](en/bohrium-web-search/SKILL.md) | `bohr search` | `bohr-web-search` | Public web search |
| [bohrium-sandbox](en/bohrium-sandbox/SKILL.md) | `bohr sandbox` | `bohr-sandbox` | Cloud sandboxes, execution, and files |
| [bohrium-lkm](en/bohrium-lkm/SKILL.md) | `bohr lkm` | `bohr-lkm` | Knowledge, claims, reasoning, and evidence |
| [bohrium-mentor](en/bohrium-mentor/SKILL.md) | `bohr agents mentor` | `bohr-mentor` | Literature-grounded scientific Q&A |

## Execution and billing

- Read the installed task workflow and relevant local references, then confirm flags through the specific command's `--help`.
- Prices and examples belong to the pinned snapshot. Query current prices through `bohr billing pricing` as documented and confirm payment source and spending authorization. Missing billing amounts do not establish an actual charge.
- Use `--yes` only for authorized operations. Preview with `--dry-run` where supported. For automated business calls, use `-o json` and check the exit status, `ok`, and business state.
- Request acceptance does not prove asynchronous completion. Verify the terminal state and results. Report missing CLI capabilities instead of automatically switching to direct API calls.

## Migrating from the API version

All 17 Skill names, language directories, and Claude installation entry points are preserved. Skill content now uses the CLI. The bundled Python API clients, example request files, API smoke/demo scripts, and `requirements.txt` have been retired. Automation invoking those scripts must move to CLI commands and consult their current help.

Update the same-named Skills through the GitHub installation path, then prepare `bohr` using the installation and sign-in steps above. Old scripts remain available in Git history. Historical API verification reports do not establish current CLI acceptance. The [OpenAPI specification](docs/api/README.md) remains independently maintained; its interface definitions and published path are unchanged by this migration.

## Source and verification

`cli-skills.lock.json` records the source npm package, published version, gitHead, public-name/guide mapping, and complete file inventory. Each language contains 64 files: 17 main documents, 31 reference documents, and 16 source-provided agent metadata files.

English main documents and references come from the same published CLI. Distribution adaptations include identities, runtime compatibility, the proxy placeholder, and corrections/task recipes individually recorded in `patches`. Following LKM, the other 16 main documents add intent routes, actual ID/result handoffs, completion evidence and recovery steps verified against the same release's implementation, retaining the detailed native references. Known corrections cover instance Wenyon mounts, output formats, paper pagination, Tools taxonomy versus tool keys, and provider-dependent Web result containers.

For updates, export the selected guides' complete trees from a published npm artifact, update the source lock, and review Chinese main documents through the existing check/release process. CI compares complete English packages with the real published CLI, validates references and agent metadata, bilingual command examples, local links/anchors, key workflow boundaries, and byte-identical full-package installation in isolated directories. These checks do not call business APIs or establish real-agent business acceptance. See [.github/workflows/check.yml](.github/workflows/check.yml).

## Help

- [Bohrium CLI introduction](https://www.bohrium.com/bohr-cli/intro)
- [CLI installation](https://docs.bohrium.com/docs/bohrctl/install/)
- [CLI pricing](https://docs.bohrium.com/docs/bohrctl/pricing)
- [CLI npm package](https://www.npmjs.com/package/@dptech-corp/bohr-cli)

## License

MIT
