# SciencePedia command reference

## Command surface

| Command | Cost and purpose |
| --- | --- |
| `bohr wiki search <query>` | 0.01 CNY; find articles, keywords, and related courses |
| `bohr wiki article <entry-id>` | 0.02 CNY; read one complete article |
| `bohr wiki keyword list` | Free; browse keyword IDs and language variants |
| `bohr wiki keyword get <keyword-id>` | 0.02 CNY; read one keyword document |
| `bohr wiki levels` | Free; list major disciplines and levels |
| `bohr wiki disciplines` | Free compatibility equivalent of `levels` |
| `bohr wiki course <field-id>` | 0.01 CNY; read a course outline and entries |
| `bohr wiki graph <entry-or-keyword-id>` | 0.02 CNY; expand a concept knowledge graph |

## Billing and automation

```bash
bohr billing pricing --resource wiki -o json
```

Wiki and Tool fixed-price operations share a personal monthly 1,000-call quota.
Within the Wiki domain, search and course cost **0.01 CNY** per call;
article, keyword get, and graph cost **0.02 CNY**. The CLI still treats these as
billable while quota remains. A non-interactive caller adds `--yes` only after
the user authorizes spending. `--dry-run` makes no request and bypasses billing
confirmation.

`levels`, `disciplines`, and `keyword list` are free discovery commands. Prefer
`-o json` when a downstream command needs an ID or when the full recursive
result matters.

## Search

```bash
bohr wiki search <query> \
  [--lang en-US] \
  [--style Feynman] \
  --yes -o json
```

The query accepts one or more positional words and is joined with spaces.
`--lang` defaults to `en-US`; `--style` defaults to `Feynman`.

Structured output has two result groups:

- `data.articles[]`: `id`, `article_name`, `type`, and
  `matched_elements[].content`. `type=article` identifies an article ID;
  `type=keyword` identifies a keyword ID.
- `data.fields[]`: related courses with `field_id` and `node_name`.

Search snippets may contain `<em>` tags around matches. Remove the tags when
answering, but preserve the underlying text. For users, both article and
keyword hits can be presented as encyclopedia entries; retain the type only for
the next command and reading-link template.

## Article and keyword documents

```bash
bohr wiki article <article_id> --lang en-US --style Feynman --yes -o json
bohr wiki keyword get <keyword_id> --lang en-US --style Feynman --yes -o json
```

Both documents commonly return `data.document.article_name` and
`data.document.main_content`; the broader service can also return definition,
key points, applications, and a short description in structured output.

When the selected language/style has no generated document, the service can
return content-not-ready code `250002`. Use the matching search snippet as a
bounded fallback, or retry another supported language/style and disclose the
change. Do not fabricate the missing full article.

### Keyword discovery

```bash
bohr wiki keyword list [--page-size 20] [--page-num 1] -o json
```

Page size and page number must be positive. Results include `items[]` with
`keyword_id`, `language`, and optional content. The same keyword can have
multiple language variants. Use the returned keyword ID with `keyword get`;
do not send it to `article`.

## Discipline hierarchy and courses

```bash
bohr wiki levels [--lang en-US] -o json
bohr wiki disciplines [--language en-US] -o json
bohr wiki course <field_id> --lang en-US --style Feynman --yes -o json
```

`levels` is the canonical hierarchy command. `disciplines` is a compatibility
equivalent and supports `--language` plus `--lang` as an alias. Both return
`data.majors[]`, each with `node_id`, `name`, and nested `levels[]`.

The current CLI obtains a course field ID from `wiki search` related fields;
major and level IDs are navigation-only and are not course field IDs. If search
returns no `field_id`, stop the course workflow and report that the current CLI
cannot enumerate courses from a level. `course` returns:

- total, foundational, core, and advanced entry counts;
- `wiki_indices[]` roots;
- recursive nodes with `node_id`, `node_name`, `node_type`, `snapshot`, `tag`,
  `children`, and an `entry_id` on article leaves.

Present the result as field → category → chapter → entry. Use snapshots for
short descriptions and preserve the foundational/core/advanced tag when it is
available.

## Knowledge graph

```bash
bohr wiki graph <entry_or_keyword_id> --yes -o json
```

Use an article or keyword ID returned by search or a course entry leaf. The
current command has no command-specific flags beyond its one positional ID.

Structured fields:

- `data.domains[]`: `major_node_id`, `major_name`.
- `data.nodes[]`: `node_id`, preferred `display_name`, fallback names,
  `node_type`, and `description`.
- `data.relationships[]`: `src_node_id`, `desc_node_id`, `relation_type`,
  `relationship`, and `description`.

Keep relationship direction. Use `relation_type` as the compact label when
present, retain the natural-language relationship and description as context,
and avoid turning a relationship into a causal or evidentiary claim unless its
text actually says so.

## User-facing reading links

Use the public site, never an API host. URL-encode dynamic IDs before inserting
them into a path. Lowercase the style segment.

| Kind | Template |
| --- | --- |
| Article | `{site}/sciencepedia/{style}/{entry_id}` |
| Keyword | `{site}/sciencepedia/{style}/keyword/{keyword_id}` |
| Course | `{site}/sciencepedia/field/{style}/{field_id}` |

`{site}` is `https://www.bohrium.com/en` for `en-US` and
`https://www.bohrium.com` for `zh-CN`.

Examples with already URL-safe placeholder IDs:

```text
https://www.bohrium.com/en/sciencepedia/feynman/<entry_id>
https://www.bohrium.com/sciencepedia/feynman/keyword/<keyword_id>
https://www.bohrium.com/en/sciencepedia/field/hardcore/<field_id>
```

For a search response, attach title, a short cleaned snippet, and the link. For
an explanation, lead with definition and intuition, then key points and the
link. For a course, preserve chapter hierarchy and link the leaf entries. For a
graph, explain the center before the strongest or most informative relations.

## CLI capability boundary

The current command tree does not expose course enumeration below an arbitrary
level, separate graph node/relationship detail requests, graph depth/domain
filters, or arbitrary service fields. Use `bohr wiki <command> --help` as the
flag contract.

## Version contract

This Skill is embedded in the `bohr` binary and changes with the CLI version:

```bash
bohr guide read bohr-wiki
bohr guide list bohr-wiki/references
bohr guide read bohr-wiki/references/commands.md
```

`bohr guide` does not fetch the standalone Bohrium Skills repository. Upgrade
`bohr` to get guidance for a newer SciencePedia command surface.

## Payment-source option

The optional global `--payment-source` flag accepts `photon` or `balance`.
Omitting it sends no preference and preserves photon-first/balance fallback.
Follow [Choose the payment source](../SKILL.md#choose-the-payment-source) for
funding authorization and insufficient-funds recovery. After balance is authorized:

```bash
bohr wiki search "graphene" --payment-source balance --yes -o json
```

The same option is available on `bohr api`. Free commands remain free; selecting
a source does not create a charge on an otherwise free operation.
