// Public adapter (dsh/bohr-cli) contract for the platform-package variant.
// Runs in the repo's CI beside scripts/dsh_plugin.test.mjs, from the repository
// root:
//   npm install --prefix dsh/bohr-cli --ignore-scripts --package-lock=false --no-audit --no-fund
//   node --test scripts/dsh_bohr_cli.test.mjs
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs/promises';
import { createRequire } from 'node:module';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath, pathToFileURL } from 'node:url';

const repo = fileURLToPath(new URL('../', import.meta.url));
const bundle = path.join(repo, 'dsh/bohr-cli');
const PLATFORMS = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
const ARCHITECTURES = { x64: 'amd64', arm64: 'arm64' };
const PLATFORM_PACKAGES = [
  '@dptech-corp/bohr-cli-darwin-amd64',
  '@dptech-corp/bohr-cli-darwin-arm64',
  '@dptech-corp/bohr-cli-linux-amd64',
  '@dptech-corp/bohr-cli-linux-arm64',
  '@dptech-corp/bohr-cli-windows-amd64',
  '@dptech-corp/bohr-cli-windows-arm64',
];

test('public adapter ships its own files, points its patch at itself, and runs the pinned native CLI', async () => {
  const manifest = JSON.parse(await fs.readFile(path.join(bundle, 'package.json'), 'utf8'));
  assert.equal(manifest.name, '@dptech-corp/bohrium-dsh-bohr-cli');
  assert.equal(manifest.scripts, undefined, 'a bundle must not run install scripts');
  assert.deepEqual(manifest.repository, {
    type: 'git',
    url: 'git+https://github.com/dptech-corp/bohrium-skills.git',
    directory: 'dsh/bohr-cli',
  });
  assert.equal(manifest.dsh.bundle.patch, './cordis.patch.yml');
  // One exact version across all six platform packages: the release train
  // publishes them together, and a mixed pin would ship two CLI versions.
  const pins = new Set(PLATFORM_PACKAGES.map((name) => manifest.optionalDependencies?.[name]));
  assert.equal(pins.size, 1, `platform packages must share one pin, got ${[...pins].join(', ')}`);
  assert.match([...pins][0], /^\d+\.\d+\.\d+$/, 'the pin must be an exact version');

  const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'bohr-cli-dsh-test-'));
  try {
    const pack = JSON.parse(execFileSync('npm', ['pack', '--ignore-scripts', '--json', '--pack-destination', temporary], {
      cwd: bundle, encoding: 'utf8',
    }))[0];
    const packedFiles = new Set(pack.files.map((file) => file.path));
    for (const file of ['index.mjs', 'cordis.patch.yml', 'package.json']) assert.ok(packedFiles.has(file), `missing ${file}`);
    assert.ok(![...packedFiles].some((file) => file.startsWith('node_modules/') || file.includes('postinstall')));

    const require = createRequire(path.join(bundle, 'package.json'));
    const { default: parseYaml } = await import(pathToFileURL(require.resolve('yaml')));
    const patch = parseYaml.parse(await fs.readFile(path.join(bundle, manifest.dsh.bundle.patch), 'utf8'));
    assert.equal(patch[0].insert[0].name, manifest.name, 'the patch row must name this package');

    // Install the packed adapter the way a profile does: exactly one platform
    // package arrives, and its binary is what the entry hands the harness.
    execFileSync('npm', ['install', '--prefix', temporary, '--ignore-scripts', '--no-audit', '--no-fund',
      path.join(temporary, pack.filename)], { encoding: 'utf8' });
    const hostPackage = `@dptech-corp/bohr-cli-${PLATFORMS[process.platform]}-${ARCHITECTURES[process.arch]}`;
    const installed = path.join(temporary, 'node_modules', manifest.name);
    const plugin = await import(pathToFileURL(path.join(installed, 'index.mjs')));
    const registered = { shellEnv: [], skills: [], warnings: [] };
    plugin.apply({
      logger: { warn: (message) => registered.warnings.push(message) },
      shellEnv: { register: (contributor) => registered.shellEnv.push(contributor) },
      skills: { register: (skill) => registered.skills.push(skill) },
    });
    assert.deepEqual(plugin.inject, ['shellEnv', 'skills']);
    assert.deepEqual(registered.warnings, []);
    assert.equal(registered.shellEnv.length, 1);
    assert.equal(registered.skills.length, 1);
    const values = registered.shellEnv[0].resolve({});
    const expectedLauncher = path.join(temporary, 'node_modules', hostPackage, 'bin',
      process.platform === 'win32' ? 'bohr.exe' : 'bohr');
    // macOS reaches the same directory as /var and /private/var; compare real paths.
    assert.equal(await fs.realpath(values.DSH_BOHR_CLI), await fs.realpath(expectedLauncher));
    const [skill] = registered.skills;
    assert.equal(skill.name, plugin.name);
    assert.equal(skill.provider, plugin.name);
    assert.match(skill.content, /guide list/);
    assert.match(skill.content, /guide read/);

    // The binary the entry points at is the CLI: version and guides are offline.
    const version = JSON.parse(execFileSync(values.DSH_BOHR_CLI, ['version', '-o', 'json'], { encoding: 'utf8' }));
    assert.equal(version.ok, true);
    assert.equal(version.data.version, [...pins][0]);
    const guides = JSON.parse(execFileSync(values.DSH_BOHR_CLI, ['guide', 'list', '-o', 'json'], { encoding: 'utf8' }));
    assert.equal(guides.ok, true);
    assert.ok(guides.data.items.length > 0);
  } finally {
    await fs.rm(temporary, { recursive: true, force: true });
  }
});

test('the CLI bundle and the skills bundle coexist in one profile and unmount on their own', async () => {
  const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'bohr-cli-coexist-'));
  let ctx;
  try {
    const tarballs = [bundle, path.join(repo, 'skills')].map((directory) => {
      const [pack] = JSON.parse(execFileSync('npm', ['pack', '--ignore-scripts', '--json', '--pack-destination', temporary], {
        cwd: directory, encoding: 'utf8',
      }));
      return path.join(temporary, pack.filename);
    });
    execFileSync('npm', ['install', '--prefix', temporary, '--ignore-scripts', '--no-audit', '--no-fund', ...tarballs], { encoding: 'utf8' });
    const skills = path.join(temporary, 'node_modules', '@dptech-corp/bohrium-dsh');
    // The skills bundle resolves its DSH filesystem provider from this checkout's
    // own install, the way the job that installs it does.
    await fs.symlink(path.join(repo, 'skills', 'node_modules'), path.join(skills, 'node_modules'), 'dir');

    const require = createRequire(path.join(process.env.DSH_TEST_PREFIX, 'package.json'));
    const { Context, Service } = await import(pathToFileURL(require.resolve('@deepseek-ai/cordis')));
    const { default: SkillRegistry } = await import(pathToFileURL(require.resolve('@deepseek-ai/dsh-skill')));
    // The host's own shell-env registry is not a standalone test dependency, so
    // this stands in for it under the same service name and register contract;
    // the real registry is covered by the host probes.
    class ShellEnvStub extends Service {
      constructor(parent) {
        super(parent, 'shellEnv');
        this.contributors = [];
      }
      register(contributor) {
        this.contributors.push(contributor);
        return () => {};
      }
    }
    ctx = new Context();
    await ctx.plugin(SkillRegistry);
    await ctx.plugin(ShellEnvStub);
    const adapter = ctx.plugin(await import(pathToFileURL(path.join(temporary, 'node_modules', '@dptech-corp', 'bohrium-dsh-bohr-cli', 'index.mjs'))));
    const skillsBundle = ctx.plugin(await import(pathToFileURL(path.join(skills, 'index.mjs'))));
    await Promise.all([adapter, skillsBundle]);

    const options = { cwd: temporary };
    assert.equal((await ctx.skills.get('bohr-cli', options)).provider, 'bohr-cli');
    assert.equal((await ctx.skills.get('bohrium-setup', options)).provider, 'bohrium');
    assert.deepEqual(ctx.shellEnv.contributors.map((item) => item.name), ['bohr-cli']);
    const snapshot = await ctx.skills.snapshot(options);
    assert.equal(snapshot.complete, true);
    assert.deepEqual(snapshot.skills.filter((item) => item.provider === 'bohr-cli').map((item) => item.name), ['bohr-cli']);
    assert.ok(snapshot.skills.some((item) => item.provider === 'bohrium' && item.name === 'bohrium-job'));

    await skillsBundle.dispose();
    assert.equal(await ctx.skills.get('bohrium-setup', options), undefined);
    assert.equal((await ctx.skills.get('bohr-cli', options)).provider, 'bohr-cli');
    await adapter.dispose();
    assert.equal(await ctx.skills.get('bohr-cli', options), undefined);
  } finally {
    await ctx?.fiber.dispose();
    await fs.rm(temporary, { recursive: true, force: true });
  }
});
