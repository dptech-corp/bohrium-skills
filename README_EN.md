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

## 2. Install these Skills

### Claude Code

The existing marketplace and plugin names are preserved. This installs the English Skills:

```text
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Codex, Claude Code, or DSH: install one Skill

Get the current GitHub content and select a directory under `zh/` or `en/`. For example, install English paper search:

```bash
git clone https://github.com/dptech-corp/bohrium-skills.git
cd bohrium-skills
bohr skill install ./en/bohrium-paper-search --target codex --scope user
```

Supported `--target` values are `codex`, `claude`, and `dsh`. Choose one language per Skill. Other clients can use the selected directory in their supported Skill location. See `bohr skill install --help` for installation options and conflict handling.

### Install all 17 Skills

With Go 1.23+, use the existing installer source from the cloned repository:

```bash
go run ./cmd/bohrium-skills-cli install --lang en
# For Chinese, use --lang zh
```

This syncs to `~/.agents/skills` (shared by Codex and DSH) and `~/.claude/skills`, backing up replaced Skills. Same-named official copies from the old installer's `~/.codex/skills` location are backed up completely and retired after canonical installation, preventing duplicate Codex discovery. Other custom Skills are preserved. Shared-root aliases that cannot be migrated safely are preserved and reported as conflicts. See the [full-set installer guide](docs/cli.md).

**Published `bohrium-skills-cli` npm packages and Release binaries embed the Skill snapshot from their release.** Updating GitHub does not change old packages. Get the current CLI-backed content through the GitHub plugin, `bohr skill install`, or this repository's installer source. This migration does not publish a new installer release.

## 3. Use the complete Skill package

For example, an installed paper-search package contains:

```text
bohrium-paper-search/
├── SKILL.md                 # Complete workflow and command examples
├── references/commands.md   # Locally packaged detailed reference
└── agents/openai.yaml       # Source-provided agent interface metadata
```

The agent reads `SKILL.md` and follows local references as needed. It does not need to fetch the entire workflow again through `bohr guide`. Chinese packages translate the main document and source-provided UI metadata; reference documents retain the CLI's original English text, with local links and section anchors preserved. Sandbox has no `agents/openai.yaml` in the source release and retains that inventory.

The CLI remains the runtime dependency. When the installed version differs from this snapshot or cross-module guidance is needed, inspect the installed CLI:

```bash
bohr --version
bohr guide list -o json
bohr guide read bohr-paper-search -o json
bohr paper --help
```

For `guide read`, the JSON instructions are in `data.content`; `data.name`, `data.version`, and `data.path` identify the source. Check `ok=true` first. Use command help and current pricing for business operations.

`bohr guide` reads local embedded guidance; `bohr agents skills` manages the cloud Skill Hub; `bohr skill install` installs a complete local Skill package.

## Skill catalog

Public Skill names remain `bohrium-*`; embedded CLI guides use `bohr-*`.

| Skill | Command entry point | Embedded guide | Purpose |
|-------|---------------------|----------------|---------|
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
| [bohrium-tools](en/bohrium-tools/SKILL.md) | `bohr tools` | `bohr-tools` | Scientific tool discovery and details |
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
