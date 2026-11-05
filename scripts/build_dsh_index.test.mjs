import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import test from 'node:test';
import { buildIndex } from './build_dsh_index.mjs';

const commit = '0123456789abcdef0123456789abcdef01234567';
const publishTimestamp = 1793865600;

test('index pins the independent bundle to a full Git commit and package directory', async () => {
  const index = buildIndex({ commit, publishTimestamp });
  assert.equal(index.format, 1);
  const plugin = index.plugins[0];
  assert.equal(plugin.name, '@dptech-corp/bohrium-dsh');
  assert.deepEqual(plugin.versions[0].source, { type: 'git', url: 'https://github.com/dptech-corp/bohrium-skills', commit, path: 'skills' });
  assert.equal(plugin.versions[0].publishTimestamp, publishTimestamp);
  const pkg = JSON.parse(await fs.readFile(new URL('../skills/package.json', import.meta.url), 'utf8'));
  assert.equal(plugin.versions[0].version, pkg.version);
  assert.equal(plugin.versions[0].engines, undefined);
  const require = createRequire(path.join(process.env.DSH_TEST_PREFIX, 'package.json'));
  const validator = await import(pathToFileURL(require.resolve('@deepseek-ai/dsh-plugin-market-index')));
  const parsed = validator.parseMarketIndex(index, 'https://dptech-corp.github.io/bohrium-skills/index.json');
  assert.equal(parsed.plugins[0].name, plugin.name);
});

test('index refuses moving refs and invalid timestamps', () => {
  for (const value of [undefined, 'main', 'v0.2.0', commit.slice(1), `${commit}0`, `https://example.com/${commit}`]) {
    assert.throws(() => buildIndex({ commit: value, publishTimestamp }), /commit/i);
  }
  for (const value of [undefined, 0, -1, 1.5, '1793865600', NaN, Infinity]) {
    assert.throws(() => buildIndex({ commit, publishTimestamp: value }), /timestamp/i);
  }
});

test('index CLI emits a static site with a colocated icon and fails on incomplete options', async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), 'bohrium-dsh-index-'));
  const cli = fileURLToPath(new URL('./build_dsh_index.mjs', import.meta.url));
  const output = path.join(directory, 'index.json');
  try {
    execFileSync(process.execPath, [cli, '--commit', commit, '--timestamp', String(publishTimestamp), '--output', output]);
    assert.deepEqual(JSON.parse(await fs.readFile(output, 'utf8')), buildIndex({ commit, publishTimestamp }));
    assert.deepEqual(await fs.readFile(path.join(directory, 'icon.svg')), await fs.readFile(new URL('../skills/icon.svg', import.meta.url)));
    assert.throws(() => execFileSync(process.execPath, [cli, '--commit', commit], { stdio: 'pipe' }));
  } finally {
    await fs.rm(directory, { recursive: true, force: true });
  }
});

test('external bundle and its public index impose no DSH runtime version gate', async () => {
  const pkg = JSON.parse(await fs.readFile(new URL('../skills/package.json', import.meta.url), 'utf8'));
  assert.equal(Object.hasOwn(pkg.engines ?? {}, 'dsh'), false, 'package must not restrict DSH versions');
  assert.deepEqual(Object.keys(pkg.peerDependencies ?? {}).filter(name => name === '@deepseek-ai/dsh' || name.startsWith('@deepseek-ai/dsh-')), [], 'DSH peers must not create an implicit version gate');
  const index = buildIndex({ commit, publishTimestamp });
  assert.equal(Object.hasOwn(index.plugins[0].versions[0], 'engines'), false, 'index must not restrict DSH versions');
});
