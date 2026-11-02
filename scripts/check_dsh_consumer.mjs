import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';

const [prefix, home, cwd, repo] = process.argv.slice(2);
const require = createRequire(path.join(prefix, 'package.json'));
const module = await import(pathToFileURL(require.resolve('@deepseek-ai/dsh-skill-filesystem')));
const warnings = [];
const abort = new AbortController();
const context = { get: () => undefined, logger: { warn: (message) => warnings.push(message), debug() {} } };
const provider = new module.FileSystemSkillProvider(context, { invalidate() {}, signal: abort.signal }, {
  watch: false, dshHome: path.join(home, '.dsh'), agentsHome: path.join(home, '.agents'),
});
try {
  const lock = JSON.parse(await fs.readFile(path.join(repo, 'cli-skills.lock.json'), 'utf8'));
  const observation = await provider.list({ cwd, signal: abort.signal });
  const candidates = Array.isArray(observation) ? observation : observation.candidates;
  assert.equal(Array.isArray(observation) || observation.complete, true, 'DSH discovery was incomplete');
  const selected = candidates.filter((item) => item.name.startsWith('bohrium-'));
  assert.equal(selected.length, lock.skills.length, 'DSH did not discover every public Skill exactly once');
  for (const entry of lock.skills) {
    const matches = selected.filter((item) => item.name === entry.name);
    assert.equal(matches.length, 1, `DSH missing/duplicate ${entry.name}`);
    const skill = await provider.get(matches[0], { cwd, signal: abort.signal });
    const root = path.join(home, '.agents', 'skills', entry.name);
    assert.equal(skill.resourceBase.kind, 'directory');
    assert.equal(skill.resourceBase.path, root);
    const original = await fs.readFile(path.join(root, 'SKILL.md'), 'utf8');
    const closing = original.indexOf('\n---\n', 4);
    assert.equal(skill.content, original.slice(closing + 5).trim(), `DSH changed the body of ${entry.name}`);
    for (const file of entry.files.filter((file) => file.startsWith('references/'))) {
      assert.equal(await fs.readFile(path.join(skill.resourceBase.path, file), 'utf8'), await fs.readFile(path.join(repo, 'en', entry.name, file), 'utf8'));
    }
  }
  assert.deepEqual(warnings, [], 'DSH ignored or rejected a published Skill');
  console.log(`DSH real filesystem provider discovered and loaded ${selected.length} complete Skills with correct resource bases.`);
} finally {
  abort.abort();
  await provider.dispose();
}
