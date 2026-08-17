# Bohrium Skill Hub

A collection of AI skills for the Bohrium platform, providing structured SKILL.md files for [OpenClaw](https://github.com/openclaw) and [Claude Code](https://claude.com/claude-code). Each skill describes a standalone capability (API calls, Agent workflows, etc.) that AI coding assistants can load on demand during conversations.

[中文](README.md)

---

## Authentication

All API Skills require a Bohrium AccessKey for authentication.

### Get your AccessKey

1. Register on [Bohrium](https://www.bohrium.com/) (enterprise users should contact DP Technology sales at bd@dp.tech)
2. Log in and go to [Settings → Account](https://www.bohrium.com/settings/account), copy your AccessKey

![Get AccessKey](docs/images/access-key-settings.png)

### Configuration

Choose one based on your runtime environment:

**Environment variable** (Claude Code / general):

```bash
export BOHR_ACCESS_KEY="your_access_key_here"
```

**OpenClaw config** `~/.openclaw/openclaw.json`:

```json
{
  "skills": {
    "<skill-name>": {
      "enabled": true,
      "apiKey": "YOUR_BOHR_ACCESS_KEY",
      "env": {
        "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY"
      }
    }
  }
}
```

---

## Skill List

### Platform API Skills

Operate Bohrium platform resources via `bohr` CLI or `open.bohrium.com` HTTP API.

| Skill | Description |
|-------|------------|
| [bohrium-job](en/bohrium-job/SKILL.md) | Compute job management — submit, list, kill, delete jobs |
| [bohrium-node](en/bohrium-node/SKILL.md) | Dev node management — create, start, stop, delete containers/VMs |
| [bohrium-dataset](en/bohrium-dataset/SKILL.md) | Dataset management — create, upload, download, version control |
| [bohrium-file](en/bohrium-file/SKILL.md) | File disk management — list, upload, download, move, copy, and delete personal/share disk files |
| [bohrium-database](en/bohrium-database/SKILL.md) | Private database — inspect schemas, query/filter/export rows, insert/update/delete records, create and alter tables |
| [bohrium-image](en/bohrium-image/SKILL.md) | Container image management — list, pull, create, delete images |
| [bohrium-project](en/bohrium-project/SKILL.md) | Project management — create projects, manage members, set budgets |
| [bohrium-knowledge-base](en/bohrium-knowledge-base/SKILL.md) | Knowledge base management — literature, tags, notes, recall search |
| [bohrium-paper-search](en/bohrium-paper-search/SKILL.md) | Paper & patent search — RAG engine keyword + semantic retrieval |
| [bohrium-pdf-parser](en/bohrium-pdf-parser/SKILL.md) | PDF parsing — extract text, tables, charts, formulas |
| [bohrium-scholar-search](en/bohrium-scholar-search/SKILL.md) | Scholar search & profile — find scholars by name/affiliation, view papers/citations/h-index/research directions |
| [bohrium-sciencepedia](en/bohrium-sciencepedia/SKILL.md) | SciencePedia — search topics/keywords for summaries & links, browse fields & courses, get a course's chapters & knowledge points, explore a topic's knowledge graph |
| [bohrium-tools](en/bohrium-tools/SKILL.md) | Scientific Tools library — browse by domain/subdomain, hybrid-search tools, view tool details & taxonomy |
| [bohrium-web-search](en/bohrium-web-search/SKILL.md) | Web search — proxy to searchapi.io for open internet search |
| [bohrium-sandbox](en/bohrium-sandbox/SKILL.md) | Cloud sandbox — on-demand temp VM for running shell/Python |
| [bohrium-lkm](en/bohrium-lkm/SKILL.md) | Large Knowledge Model — node search, reasoning-chain retrieval, paper knowledge graphs, claim provenance tracing, batch node hydration, async PDF extraction, feedback submission |
| [bohrium-mentor](en/bohrium-mentor/SKILL.md) | AI Science Mentor — deep-reasoning scientific Q&A with automatic literature retrieval, structured Markdown answers |

---

## Billing

Charged skills bill your account balance per call or per compute-hour. Check your balance and bills on the [Research Assets](https://www.bohrium.com/en/assets):

| Skill | Charged | Price | Unit | Notes |
|-------|:-------:|-------|------|-------|
| bohrium-job | Yes | See pricing page | ¥/hour | Compute-hour fee after machine starts; see Job pricing page |
| bohrium-node | Yes | See pricing page | ¥/hour | Compute-hour fee after machine starts; see Node pricing page |
| bohrium-sandbox | Yes | See pricing page | ¥/hour | Compute-hour fee after machine starts; see Node pricing page |
| bohrium-paper-search | Yes | From ¥0.05/call | ¥/call | Paper: standard (type 0) ¥0.05, enhanced (type 1) ¥0.1/call; Patent: type 0 ¥0.1, type 1 ¥0.3, type 2 ¥0.5/call |
| bohrium-pdf-parser | Yes | ¥0.05/page | ¥/page | Charged on trigger; fetching results is free |
| bohrium-lkm | Yes | First 1,000 calls/month free; then ¥0.05/call | ¥/call | ¥0.05 per billable call (search, reasoning/search, papers/graph, claims/{id}/reasoning, variables/batch); feedback is free |
| bohrium-mentor | Yes | ¥2.0/call | ¥/call | Charged on session creation; ¥2/call or 200 photons/call |
| bohrium-sciencepedia | Yes | First 1,000 calls/month shared with Tools are free; then from ¥0.01/call | ¥/call | article, keyword, knowledge_graph ¥0.02/call each; search/universal, get_wiki_index ¥0.01/call each |
| bohrium-tools | Yes | First 1,000 calls/month shared with SciencePedia are free; then from ¥0.01/call | ¥/call | search/hybrid ¥0.01/call; detail ¥0.02/call |
| bohrium-dataset / bohrium-image / bohrium-project / bohrium-knowledge-base / bohrium-scholar-search / bohrium-web-search | Free | - | - | - |

---

## Installation

### Option 1: Claude Code Plugin

```
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Option 2: bohrium-skills-cli (recommended)

```bash
npm install -g bohrium-skills-cli
```

After install, skills are automatically synced to `~/.agents/skills`, `~/.claude/skills`, and `~/.codex/skills`. To update:

```bash
bohrium-skills-cli update
```

> For more options (manual binary install, env vars, all commands) see [CLI Guide](docs/cli.md).

---

## Directory Structure

```
bohrium-skill-hub/
├── zh/                          # Chinese version
│   ├── bohrium-job/SKILL.md
│   ├── bohrium-node/SKILL.md
│   └── ...
├── en/                          # English version
│   ├── bohrium-job/SKILL.md
│   └── ...
├── docs/images/                 # Documentation images
├── README.md                    # Chinese README
└── README_EN.md                 # English README (this file)
```

## SKILL.md Format

Each SKILL.md contains at least:

```yaml
---
name: skill-name
description: "One-line description. Use when: ... NOT for: ..."
---
```

- **Frontmatter** — `name` + `description` (with use/exclusion scenarios); optionally `version` and `metadata.openclaw.primaryEnv`
- **Body** — Feature description, API endpoints, parameter tables, response fields, code examples, error handling
- **Code examples** — Python `requests` style, preferring `os.environ.get("BOHR_ACCESS_KEY")`, never hardcoded

---

## License

MIT
