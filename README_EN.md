# Bohrium Skills

Agent Skills for the Bohrium scientific computing platform, powered by the [Bohrium CLI (`bohr`)](https://www.bohrium.com/bohr-cli/intro).

[中文](README.md)

This repository provides 17 Skills in Chinese and English. Each Skill identifies the task, checks the CLI, and loads its embedded guidance. Detailed commands, flags, billing, and recovery rules stay aligned with the installed CLI version.

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

This syncs to `~/.agents/skills`, `~/.claude/skills`, and `~/.codex/skills`, backing up replaced Skills. See the [full-set installer guide](docs/cli.md).

**Published `bohrium-skills-cli` npm packages and Release binaries embed the Skill snapshot from their release.** Updating GitHub does not change old packages. Get the current CLI-backed content through the GitHub plugin, `bohr skill install`, or this repository's installer source. This migration does not publish a new installer release.

## 3. Read version-matched guidance

```bash
bohr guide list -o json
bohr guide read bohr-paper-search -o human
bohr guide list bohr-paper-search/references -o json
bohr guide read bohr-paper-search references/commands.md -o human
bohr paper --help
```

Relative links inside a guide refer to CLI-embedded resources. Read them with `bohr guide read <guide> <path> -o human`; they are not files in this repository's Skill directories. If a command or guide is missing, check `bohr --version` and install the latest stable release.

`bohr guide` reads local embedded guidance; `bohr agents skills` manages the cloud Skill Hub; `bohr skill install` installs a local Skill package.

## Skill catalog

Public Skill names remain `bohrium-*`; embedded CLI guides use `bohr-*`.

| Skill | Command entry point | Embedded guide | Purpose |
|-------|---------------------|----------------|---------|
| [bohrium-job](en/bohrium-job/SKILL.md) | `bohr job` | `bohr-job` | Compute jobs, logs, and results |
| [bohrium-node](en/bohrium-node/SKILL.md) | `bohr node` | `bohr-node` | Interactive development nodes |
| [bohrium-dataset](en/bohrium-dataset/SKILL.md) | `bohr dataset` | `bohr-dataset` | Datasets, versions, and files |
| [bohrium-file](en/bohrium-file/SKILL.md) | `bohr file` | `bohr-file` | Personal/share file storage |
| [bohrium-database](en/bohrium-database/SKILL.md) | `bohr database` | `bohr-database` | Scientific databases and table schemas |
| [bohrium-image](en/bohrium-image/SKILL.md) | `bohr image` | `bohr-image` | Container images, builds, and snapshots |
| [bohrium-project](en/bohrium-project/SKILL.md) | `bohr project` | `bohr-project` | Projects, members, and quotas |
| [bohrium-knowledge-base](en/bohrium-knowledge-base/SKILL.md) | `bohr kb` | `bohr-knowledge-base` | Knowledge bases, literature, and retrieval |
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

- Read the embedded guide before business operations, then confirm flags through the specific command's `--help`.
- Use `bohr billing pricing --help` to discover the current pricing query. Follow the guide for cost, payment source, and spending authorization; this repository does not duplicate the price catalog.
- Use `--yes` only for authorized operations. Preview with `--dry-run` where supported. For automated business calls, use `-o json` and check the exit status, `ok`, and business state.
- Request acceptance does not prove asynchronous completion. Verify the terminal state and results. Report missing CLI capabilities instead of automatically switching to direct API calls.

## Migrating from the API version

All 17 Skill names, language directories, and Claude installation entry points are preserved. Skill content now uses the CLI. The bundled Python API clients, example request files, API smoke/demo scripts, and `requirements.txt` have been retired. Automation invoking those scripts must move to CLI commands and consult their current help.

Update the same-named Skills through the GitHub installation path, then prepare `bohr` using the installation and sign-in steps above. Old scripts remain available in Git history. Historical API verification reports do not establish current CLI acceptance. The [OpenAPI specification](docs/api/README.md) remains independently maintained; its interface definitions and published path are unchanged by this migration.

## Content maintenance and verification

`zh/<name>/SKILL.md` and `en/<name>/SKILL.md` each provide a task entry point. Detailed reference files live in `bohr`; this repository references guide names and embedded resource paths.

GitHub CI checks all Skills, the Claude manifest, the installer, and guide/help commands against the real published CLI without calling business APIs. See [.github/workflows/check.yml](.github/workflows/check.yml) for test commands and version requirements.

## Help

- [Bohrium CLI introduction](https://www.bohrium.com/bohr-cli/intro)
- [CLI installation](https://docs.bohrium.com/docs/bohrctl/install/)
- [CLI pricing](https://docs.bohrium.com/docs/bohrctl/pricing)
- [CLI npm package](https://www.npmjs.com/package/@dptech-corp/bohr-cli)

## License

MIT
