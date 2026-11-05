import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { createRequire } from 'node:module';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import test from 'node:test';

const repo = fileURLToPath(new URL('../', import.meta.url));
const bundle = path.join(repo, 'skills');

test('packed DSH bundle loads all complete Skills and removes only its own contributions', async () => {
  const manifest = JSON.parse(await fs.readFile(path.join(bundle, 'package.json'), 'utf8'));
  assert.equal(manifest.name, '@dptech-corp/bohrium-dsh');
  assert.equal(manifest.scripts, undefined, 'bundle must not run the legacy installer');
  assert.equal(manifest.version, JSON.parse(await fs.readFile(path.join(repo, 'package.json'), 'utf8')).version);
  const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'bohrium-dsh-test-'));
  let ctx;
  try {
    const pack = JSON.parse(execFileSync('npm', ['pack', '--ignore-scripts', '--json', '--pack-destination', temporary], { cwd: bundle, encoding: 'utf8' }))[0];
    const packedFiles = new Set(pack.files.map((file) => file.path));
    for (const file of ['index.mjs', 'cordis.patch.yml', 'locale/en.json', 'locale/zh.json', 'icon.svg']) assert.ok(packedFiles.has(file), `missing ${file}`);
    assert.ok(![...packedFiles].some((file) => file.startsWith('node_modules/') || file.includes('postinstall')));
    execFileSync('tar', ['-xzf', path.join(temporary, pack.filename), '-C', temporary]);
    const installed = path.join(temporary, 'package');
    await fs.symlink(path.join(bundle, 'node_modules'), path.join(installed, 'node_modules'), 'dir');
    const require = createRequire(path.join(process.env.DSH_TEST_PREFIX, 'package.json'));
    const { Context } = await import(pathToFileURL(require.resolve('@deepseek-ai/cordis')));
    const { default: SkillRegistry } = await import(pathToFileURL(require.resolve('@deepseek-ai/dsh-skill')));
    const bundleRequire = createRequire(path.join(bundle, 'package.json'));
    const { default: parseYaml } = await import(pathToFileURL(bundleRequire.resolve('yaml')));
    const filesystem = await import(pathToFileURL(bundleRequire.resolve('@deepseek-ai/dsh-skill-filesystem')));
    const plugin = await import(pathToFileURL(path.join(installed, 'index.mjs')));
    const patch = parseYaml.parse(await fs.readFile(path.join(installed, manifest.dsh.bundle.patch), 'utf8'));
    assert.equal(patch[0].insert[0].name, manifest.name);
    ctx = new Context();
    await ctx.plugin(SkillRegistry);
    const userHome = path.join(temporary, 'user-dsh');
    const personalDirectory = path.join(userHome, 'skills/personal-example');
    await fs.mkdir(personalDirectory, { recursive: true });
    await fs.writeFile(path.join(personalDirectory, 'SKILL.md'), '---\nname: personal-example\ndescription: Existing user skill\n---\nKeep this user instruction.\n');
    const userConfig = { providerName: 'user-skills', dshHome: userHome, agentsHome: path.join(temporary, 'agents'), watch: false };
    let user = ctx.plugin(filesystem, userConfig);
    await user;
    const mounted = ctx.plugin(plugin);
    await mounted;
    const lock = JSON.parse(await fs.readFile(path.join(repo, 'cli-skills.lock.json'), 'utf8'));
    const expected = [...lock.skills.map((entry) => entry.name), 'bohrium-setup'].sort();
    const snapshot = await ctx.skills.snapshot({ cwd: temporary });
    assert.equal(snapshot.complete, true);
    assert.deepEqual(snapshot.skills.filter((item) => item.provider === 'bohrium').map((item) => item.name).sort(), expected);
    for (const name of expected) {
      const loaded = await ctx.skills.get(name, { cwd: temporary });
      assert.equal(loaded.source, 'bundled');
      assert.equal(loaded.invocation.modelInvocable, true);
      assert.equal(loaded.resourceBase.kind, 'directory');
      const source = path.join(bundle, name);
      const raw = await fs.readFile(path.join(source, 'SKILL.md'), 'utf8');
      const frontmatter = /^---\r?\n[\s\S]*?\r?\n---(?:\r?\n|$)/u.exec(raw);
      assert.equal(loaded.content, raw.slice(frontmatter[0].length).trim(), name);
      for (const file of await fs.readdir(source, { recursive: true, withFileTypes: true })) {
        if (!file.isFile()) continue;
        const relative = path.relative(source, path.join(file.parentPath, file.name));
        assert.ok(packedFiles.has(`${name}/${relative.split(path.sep).join('/')}`), `packed resource missing: ${name}/${relative}`);
        assert.deepEqual(await fs.readFile(path.join(loaded.resourceBase.path, relative)), await fs.readFile(path.join(source, relative)));
      }
    }
    const overrideDirectory = path.join(userHome, 'skills/bohrium-job');
    await fs.mkdir(overrideDirectory);
    await fs.writeFile(path.join(overrideDirectory, 'SKILL.md'), '---\nname: bohrium-job\ndescription: User override\n---\nMy existing workflow.\n');
    await user.dispose();
    user = ctx.plugin(filesystem, userConfig);
    await user;
    assert.equal((await ctx.skills.get('bohrium-job')).content, 'My existing workflow.');
    const project = path.join(temporary, 'project');
    const projectSkill = path.join(project, '.dsh/skills/bohrium-job');
    await fs.mkdir(projectSkill, { recursive: true });
    await fs.mkdir(path.join(project, '.git'));
    await fs.writeFile(path.join(projectSkill, 'SKILL.md'), '---\nname: bohrium-job\ndescription: Project override\n---\nProject workflow wins.\n');
    assert.equal((await ctx.skills.get('bohrium-job', { cwd: project })).content, 'Project workflow wins.');
    await mounted.dispose();
    assert.equal(await ctx.skills.get('bohrium-paper-search'), undefined);
    assert.equal((await ctx.skills.get('bohrium-job', { cwd: project })).content, 'Project workflow wins.');
    assert.equal((await ctx.skills.get('bohrium-job')).content, 'My existing workflow.');
    assert.equal((await ctx.skills.get('personal-example')).content, 'Keep this user instruction.');
    await fs.rm(overrideDirectory, { recursive: true });
    await user.dispose();
    user = ctx.plugin(filesystem, userConfig);
    await user;
    const remounted = ctx.plugin(plugin);
    await remounted;
    assert.equal((await ctx.skills.get('bohrium-job')).provider, 'bohrium');
    await remounted.dispose();
    await user.dispose();
  } finally {
    await ctx?.fiber.dispose();
    await fs.rm(temporary, { recursive: true, force: true });
  }
});
