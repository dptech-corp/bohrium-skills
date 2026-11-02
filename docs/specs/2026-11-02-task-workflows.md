# LKM-style task workflows

## Goal and scope

The user asks for the other public Skills to guide a task as concretely as the native LKM Skill. Preserve the 17 public identities and complete source packages; enrich the other 16 main documents with intent routing, ordered execution, identifier/result handoffs, completion evidence, and recovery. LKM remains the reference example.

## Decisions

- Verify command contracts against the published `bohr v2.9.19` implementation and existing local references. Public JSON fields come from the actual output path, including raw upstream envelopes, rather than unused internal result types.
- Add a short task-specific route table and one complete recipe to each main document. Reuse the detailed native sections for cost, flags and exceptional branches. Simple retrieval domains need evidence selection and result interpretation, not invented asynchronous tasks or IDs.
- Show where each required ID comes from and what consumes it. For unnormalized upstream results, explicitly inspect the returned record instead of asserting an undocumented fixed JSON path.
- Completion means the requested deliverable is verified: terminal compute and downloaded outputs, matching dataset version, complete parser result, selected source evidence, or validated retrieval/synthesis. Request acceptance alone is insufficient.
- Recovery resumes the same resource/token when possible; ambiguous create outcomes are reconciled before new writes, and failed paid synthesis is not retried automatically.
- Record curated English additions as deterministic source adaptations in `cli-skills.lock.json`. Translate their prose in Chinese and keep fenced examples identical. Keep the existing 64-file inventory per language and source-reference content.

## Acceptance and verification

1. All 16 enriched main documents contain domain-specific routes, a concrete ordered recipe, handoff rules and completion/recovery criteria; the LKM exemplar remains intact.
2. Job/node/image/dataset/project and research identifiers are verified against the published output implementation. No guessed flags or JSON paths.
3. Existing source parity, bilingual examples, links, full installer and real Codex/Claude/DSH consumer checks still pass in CI.
4. Dry-run/help checks cover representative new command chains without authentication or paid requests. Text checks alone are not business acceptance; no paid or production operation is performed.
5. No private CLI source change, new public command, main merge or release.
