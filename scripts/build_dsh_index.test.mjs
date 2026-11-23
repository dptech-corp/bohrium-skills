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
const bundles = [['skills', '@dptech-corp/bohrium-dsh'], ['dsh/bohr-cli', '@dptech-corp/bohrium-dsh-bohr-cli']];

test('index pins both bundles to a full Git commit and their own package directory', async () => {
  const index = buildIndex({ commit, publishTimestamp });
  assert.equal(index.format, 1);
  const require = createRequire(path.join(process.env.DSH_TEST_PREFIX, 'package.json'));
  const validator = await import(pathToFileURL(require.resolve('@deepseek-ai/dsh-plugin-market-index')));
  const parsed = validator.parseMarketIndex(index, 'https://dptech-corp.github.io/bohrium-skills/index.json');
  assert.deepEqual(parsed.plugins.map((plugin) => plugin.name), bundles.map(([, name]) => name));
  for (const [directory, name] of bundles) {
    const plugin = index.plugins.find((item) => item.name === name);
    assert.deepEqual(plugin.versions[0].source, { type: 'git', url: 'https://github.com/dptech-corp/bohrium-skills', commit, path: directory });
    assert.equal(plugin.versions[0].publishTimestamp, publishTimestamp);
    assert.equal(plugin.repository.directory, directory);
    const pkg = JSON.parse(await fs.readFile(new URL(`../${directory}/package.json`, import.meta.url), 'utf8'));
    assert.equal(plugin.versions[0].version, pkg.version);
    assert.equal(plugin.versions[0].engines, undefined);
  }
  // Only the skills bundle ships an icon; the CLI bundle has none to copy.
  assert.deepEqual(index.plugins.map((plugin) => plugin.icon), ['icon.svg', undefined]);
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

test('external bundles and their public index impose no DSH runtime version gate', async () => {
  for (const [directory] of bundles) {
    const pkg = JSON.parse(await fs.readFile(new URL(`../${directory}/package.json`, import.meta.url), 'utf8'));
    assert.equal(Object.hasOwn(pkg.engines ?? {}, 'dsh'), false, `${directory} package must not restrict DSH versions`);
    assert.deepEqual(Object.keys(pkg.peerDependencies ?? {}).filter(name => name === '@deepseek-ai/dsh' || name.startsWith('@deepseek-ai/dsh-')), [], 'DSH peers must not create an implicit version gate');
  }
  const index = buildIndex({ commit, publishTimestamp });
  for (const plugin of index.plugins) {
    assert.equal(Object.hasOwn(plugin.versions[0], 'engines'), false, 'index must not restrict DSH versions');
  }
});
