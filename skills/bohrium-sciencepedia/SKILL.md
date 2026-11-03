---
name: bohrium-sciencepedia
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Search and explain scientific terms, read Bohrium SciencePedia articles and
  keywords, browse disciplines and course outlines, and explore concept graphs
  with the `bohr wiki` CLI. Use for SciencePedia, 科学百科, scientific concept
  explanations, course knowledge trees, or encyclopedia knowledge graphs. Use
  `bohr paper` or `bohr kb` for paper search or personal knowledge bases.
---

# Bohrium SciencePedia

## Workflow: explain a concept with its prerequisites

| User's deliverable | Ordered route |
|---|---|
| A concept explanation | Search → route article/keyword ID → read → explain with a reading link |
| A learning outline | Search → returned `field_id` → course tree → selected entry articles |
| Related concepts | Confirm entry/keyword ID → graph → preserve edge direction |

1. Set the requested language/style and authorize the bounded search/read/graph calls using the prices below.
2. Search, then choose a result using the ID routing section. The following is an article branch; a `type=keyword` result uses `keyword get` instead:

```bash
bohr wiki search "quantum entanglement" --lang en-US --style Feynman --yes -o json
bohr wiki article <article_id> --lang en-US --style Feynman --yes -o json
bohr wiki graph <article_id> --yes -o json
```

3. Replace `<article_id>` with `data.articles[].id` only from the selected `type=article` hit. Graph expansion is optional: run it only if relationships are requested and its cost is authorized. A course uses `data.fields[].field_id`, not a major/level ID.
4. Complete with the supported explanation, source reading link and relevant prerequisites/relationships. Preserve graph direction; a relation label is not proof of causality. For a course, retain chapter/entry hierarchy rather than flattening it into search hits.
5. If content is ungenerated/blank, apply the language/style fallback below and disclose it. If no course `field_id` is returned, stop that branch rather than substituting another ID. Repeated paid search or graph expansion requires room in the authorized budget.

Operate SciencePedia through the version-matched `bohr wiki` commands. Choose
the route from the user's intent:

- Search a term and get snippets plus IDs with `wiki search`.
- Read a complete topic with `wiki article` or a smaller concept card with
  `wiki keyword get`.
- Browse the free major/level hierarchy with `wiki levels`.
- Read a course's chapter and knowledge-entry tree with `wiki course`.
- Expand an article or keyword into related concepts with `wiki graph`.
- Find scientific software in SciencePedia's tool library with `bohr wiki tools`;
  read `bohr guide read bohr-tools` for its routes and prices.

Use `bohr paper search` for papers and `bohr kb` for a personal knowledge base.
SciencePedia is an encyclopedia, not either of those systems.

## Authentication, output, and billing

Use an existing Bohrium login or AccessKey and prefer structured output:

```bash
bohr auth whoami -o json
bohr billing pricing --resource wiki -o json
```

Fixed prices in the current CLI are:

- **0.01 CNY**: `wiki search` and `wiki course`.
- **0.02 CNY**: `wiki article`, `wiki keyword get`, and `wiki graph`.

These calls use the shared monthly 1,000-call Wiki/Tool quota. The CLI
confirmation gate applies even when quota may make a call free. In
non-interactive use, add `--yes` only after spending is authorized; otherwise
report `CONFIRMATION_REQUIRED`.

`wiki levels`, its `wiki disciplines` compatibility equivalent, and
`wiki keyword list` are free discovery commands.

## Search and route IDs

```bash
bohr wiki search "quantum entanglement" \
  --lang en-US \
  --style Feynman \
  --yes \
  -o json
```

Search mixes concepts and courses. Route by the structured result, not by a
display name:

- `data.articles[]` with `type=article`: its `id` is an article ID for
  `bohr wiki article <id>`.
- `data.articles[]` with `type=keyword`: its `id` is a keyword ID for
  `bohr wiki keyword get <id>`.
- `data.fields[]`: its `field_id` is a course field ID for
  `bohr wiki course <field_id>`.
- An entry or keyword ID can be the center for `bohr wiki graph <id>`.

For a user-facing search answer, present title, a cleaned one-sentence snippet,
and the reading link. Remove `<em>` highlight tags from structured snippets;
the result type matters for command and link routing but need not burden the
user.

## Read concepts and courses

```bash
bohr wiki article <article_id> --lang en-US --style Feynman --yes -o json
bohr wiki keyword get <keyword_id> --lang en-US --style Feynman --yes -o json
bohr wiki levels --lang en-US -o json
bohr wiki course <field_id> --lang en-US --style Feynman --yes -o json
```

`levels` returns majors and their level IDs. `disciplines` is a compatibility
equivalent that calls the same supported hierarchy; prefer `levels` in new
workflows. The current CLI does not list every course beneath a selected level,
so obtain a course field ID from `wiki search`. Major and level IDs are
navigation-only in the current CLI; they are not course field IDs.
If search returns no `field_id`, stop the course workflow and report the
limitation.

`course` returns a recursive field → category → chapter → entry tree plus
foundational, core, and advanced counts. Preserve that hierarchy when
presenting an outline. Entry leaves carry `entry_id` values that can be opened
as SciencePedia articles.

If article or keyword content is not generated for the chosen language/style,
fall back to the search snippet or retry another supported language/style and
state the fallback. Do not report a blank document as a complete explanation.

## Explore a concept graph

```bash
bohr wiki graph <entry_or_keyword_id> --yes -o json
```

The graph returns domains, concept nodes, and relationships. Use node
`display_name` with `node_name`, `name`, and `node_id` as fallbacks. Explain the
center first, then the most relevant relationships; preserve source/destination
direction and distinguish a relationship label from its longer description.

The current graph command has no language, style, depth, node-detail, or
cross-domain filter flags. Use only flags shown by `bohr wiki graph --help`.

## Reading links

API access remains inside `bohr`, but user-facing reading links use
`https://www.bohrium.com`. URL-encode every dynamic ID.

- Article: `{site}/sciencepedia/{style-lowercase}/{entry_id}`.
- Keyword: `{site}/sciencepedia/{style-lowercase}/keyword/{keyword_id}`.
- Course: `{site}/sciencepedia/field/{style-lowercase}/{field_id}`.

For `en-US`, `{site}` is `https://www.bohrium.com/en`; for `zh-CN`, it is
`https://www.bohrium.com`. For example, an English Feynman article begins with
`/en/sciencepedia/feynman/`, a Chinese Feynman keyword contains
`/sciencepedia/feynman/keyword/`, and every course contains
`/sciencepedia/field/`.

## Capability boundary

The current CLI intentionally omits several broader service operations,
including course listing by selected level, separate graph-node or graph-edge
detail calls, and graph depth/domain filters. Inspect `bohr wiki <command>
--help` rather than converting service request fields into flags.

## Choose the payment source

Without `--payment-source`, the CLI sends no channel preference: the gateway
uses photons first and falls back to balance only on definite photon
insufficiency. Tell the user this policy when obtaining spending authorization.
To restrict spending to one source, use `--payment-source photon` or
`--payment-source balance`. Explicit photon stops when photons are insufficient;
explicit balance charges balance directly. Existing free quotas and exemptions
still apply.

When an explicitly selected channel returns `BILLING_INSUFFICIENT` with
`retryable: false`, report the refusal and wait
for the user to recharge the selected source or authorize the other source for
the original command. Explicit balance insufficiency also stops without charging
photons; use `--payment-source photon` only after that source is authorized.
`--yes` confirms execution; it does not authorize changing the payment source.
Use `--payment-source balance` only when that source is authorized. Keep the
selected source on subsequent billable calls in a multi-step workflow; options
from a previous CLI invocation are not saved. `bohr api` follows the same rule.

## Detailed reference

Read [`references/commands.md`](references/commands.md) for the complete command,
flag, output, billing, fallback, graph, reading-link, and version reference.
