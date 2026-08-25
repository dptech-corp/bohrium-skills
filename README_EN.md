# Bohrium Skills

AI Agent skills for the Bohrium scientific computing platform.

> For new integrations, we recommend starting with [Bohrium CLI (`bohr`)](https://www.bohrium.com/bohr-cli/intro). It provides a unified command-line interface and embeds Agent Skills that match the installed CLI version. This repository continues to keep the standalone `SKILL.md` files available for existing integrations.

[中文](README.md)

## Use Bohrium CLI

### Install

```bash
npm install -g @dptech-corp/bohr-cli
```

The package supports x64 and arm64 on macOS, Linux, and Windows. After installation, we recommend signing in to your Bohrium account through the browser:

```bash
bohr auth login    # Sign in to Bohrium
bohr doctor        # Check local configuration and connectivity
bohr --help        # List all commands
```

### Common commands

Start with the root command to discover Bohrium CLI capabilities, then enter a command namespace to inspect its subcommands and options:

```bash
bohr --help
bohr wiki --help
bohr wiki search --help
```

Other capabilities follow the same pattern:

```bash
bohr job --help
bohr sandbox --help
bohr lkm --help
bohr agents mentor --help
```

The `bohr` command corresponding to each standalone Skill is listed below. Always use the installed CLI's `--help` output as the source of truth for subcommands and options.

### Agent Skills

`bohr` embeds command guidance designed for AI Agents, versioned together with the CLI. First check which Skills are available in the installed version; commands not listed there remain discoverable through `bohr <command> --help`:

```bash
bohr skills list                              # List Skills embedded in this version
bohr skills read bohr-sandbox                 # Read a Skill
bohr skills list bohr-sandbox/references      # List a Skill's references
```

Upgrade the CLI to receive new commands and their matching Skills:

```bash
bohr update
```

## Help and documentation

- [Bohrium CLI overview](https://www.bohrium.com/bohr-cli/intro)
- [Bohrium CLI installation guide](https://docs.bohrium.com/docs/bohrctl/install/)
- [Bohrium CLI pricing](https://docs.bohrium.com/docs/bohrctl/pricing)
- [Bohrium CLI on npm](https://www.npmjs.com/package/@dptech-corp/bohr-cli)
- [Bohrium platform](https://www.bohrium.com/)
- In-command help: `bohr --help`, `bohr <command> --help`

## Standalone Skills in this repository

The following `SKILL.md` files remain available for Agents and automation that already depend on this repository. New integrations should prefer `bohr` and its embedded Skills.

### Existing integration configuration

> This section applies only to existing integrations that continue to use the standalone Skills in this repository directly. For new integrations, we still recommend installing `bohr` and running `bohr auth login`.

Standalone Skills that call the platform API directly require a Bohrium AccessKey. Get one from [Bohrium account settings](https://www.bohrium.com/settings/account), then provide it through an environment variable:

![Get AccessKey](docs/images/access-key-settings.png)

```bash
export BOHR_ACCESS_KEY="YOUR_ACCESS_KEY"
```

Never hard-code or commit real credentials.

Existing Claude Code plugin users can continue to use the original installation method:

```text
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Skill list

| Skill | Corresponding `bohr` command | Description |
|-------|------------------------------|-------------|
| [bohrium-job](en/bohrium-job/SKILL.md) | `bohr job` | Compute job management — submit, query, stop, and delete jobs |
| [bohrium-node](en/bohrium-node/SKILL.md) | `bohr node` | Development node management — create, start, stop, and delete containers/VMs |
| [bohrium-dataset](en/bohrium-dataset/SKILL.md) | `bohr dataset` | Dataset management — create, upload, download, and version datasets |
| [bohrium-file](en/bohrium-file/SKILL.md) | `bohr file` | File storage management — list, upload, download, move, copy, and delete files |
| [bohrium-database](en/bohrium-database/SKILL.md) | `bohr database` | Scientific databases — browse domain libraries and tabular data |
| [bohrium-image](en/bohrium-image/SKILL.md) | `bohr image` | Container image management — list, pull, create, and delete images |
| [bohrium-project](en/bohrium-project/SKILL.md) | `bohr project` | Project management — create projects, manage members, and set budgets |
| [bohrium-knowledge-base](en/bohrium-knowledge-base/SKILL.md) | `bohr kb` | Knowledge base management — literature, tags, notes, and recall search |
| [bohrium-paper-search](en/bohrium-paper-search/SKILL.md) | `bohr paper` | Paper and patent search — keyword and semantic retrieval |
| [bohrium-pdf-parser](en/bohrium-pdf-parser/SKILL.md) | `bohr pdf` | PDF parsing — extract text, tables, charts, and formulas |
| [bohrium-scholar-search](en/bohrium-scholar-search/SKILL.md) | `bohr scholar` | Scholar search and profiles — publications, citations, metrics, and research areas |
| [bohrium-sciencepedia](en/bohrium-sciencepedia/SKILL.md) | `bohr wiki` | SciencePedia — search topics, courses, knowledge points, and knowledge graphs |
| [bohrium-tools](en/bohrium-tools/SKILL.md) | `bohr tools` | Scientific tools — browse, search, and inspect tools |
| [bohrium-web-search](en/bohrium-web-search/SKILL.md) | `bohr search` | Web search — search the open internet |
| [bohrium-sandbox](en/bohrium-sandbox/SKILL.md) | `bohr sandbox` | Cloud sandbox — create temporary VMs and run Shell/Python |
| [bohrium-lkm](en/bohrium-lkm/SKILL.md) | `bohr lkm` | Large Knowledge Model — knowledge retrieval, reasoning chains, paper graphs, and async PDF extraction |
| [bohrium-mentor](en/bohrium-mentor/SKILL.md) | `bohr agents mentor` | AI Science Mentor — literature-grounded scientific Q&A |

## Billing

Charged skills bill your account balance per call or per compute-hour. Check your balance and bills on the [Research Assets](https://www.bohrium.com/en/assets). Refer to [Bohrium CLI pricing](https://docs.bohrium.com/docs/bohrctl/pricing) for the latest pricing rules:

| Skill | Charged | Price | Unit | Notes |
|-------|:-------:|-------|------|-------|
| bohrium-job | Yes | See pricing page | ¥/hour | Compute-hour fee after machine starts; see Job pricing page |
| bohrium-node | Yes | See pricing page | ¥/hour | Compute-hour fee after machine starts; see Node pricing page |
| bohrium-sandbox | Yes | See pricing page | ¥/hour | Compute-hour fee after machine starts; see Node pricing page |
| bohrium-paper-search | Yes | From ¥0.05/call | ¥/call | Paper: standard (type 0) ¥0.05, enhanced (type 1) ¥0.1/call; Patent: type 0 ¥0.1, type 1 ¥0.3, type 2 ¥0.5/call |
| bohrium-pdf-parser | Yes | ¥0.05/page | ¥/page | Charged on trigger; fetching results is free |
| bohrium-lkm | Yes | Search ¥0.05/call; Parse ¥1 / ¥0.1 on cache hit | ¥/call | Search endpoints (search, reasoning/search, papers/graph, claims/{id}/reasoning, variables/batch): first 1,000 calls/month free, then ¥0.05/call. Parse: ¥1/call, ¥0.1/call on cache hit. feedback is free |
| bohrium-mentor | Yes | ¥2.0/call | ¥/call | Charged on session creation; ¥2/call or 200 photons/call |
| bohrium-sciencepedia | Yes | First 1,000 calls/month shared with Tools are free; then from ¥0.01/call | ¥/call | article, keyword, knowledge_graph ¥0.02/call each; search/universal, get_wiki_index ¥0.01/call each |
| bohrium-tools | Yes | First 1,000 calls/month shared with SciencePedia are free; then from ¥0.01/call | ¥/call | search/hybrid ¥0.01/call; detail ¥0.02/call |
| bohrium-dataset / bohrium-image / bohrium-project / bohrium-knowledge-base / bohrium-scholar-search / bohrium-web-search | Free | - | - | - |

## Repository structure

```text
bohrium-skills/
├── zh/              # Chinese Skills
├── en/              # English Skills
├── docs/            # Legacy documentation and assets
├── README.md        # Chinese README
└── README_EN.md     # English README (this file)
```

## SKILL.md format

Each `SKILL.md` contains at least:

```yaml
---
name: skill-name
description: "One-line description. Use when: ... NOT for: ..."
---
```

- **Frontmatter** — `name` + `description` with use and exclusion scenarios; `version` is optional.
- **Body** — Feature description, API endpoints, parameter tables, response fields, code examples, and error handling.
- **Code examples** — Use Python `requests` style and never hard-code credentials.

## License

MIT
