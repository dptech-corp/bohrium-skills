import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseArgs } from 'node:util';
import { checkedReleaseVersion } from './check_release_version.mjs';

const manifest = JSON.parse(fs.readFileSync(new URL('../skills/package.json', import.meta.url), 'utf8'));
const root = JSON.parse(fs.readFileSync(new URL('../package.json', import.meta.url), 'utf8'));

/** Build a format-1 index with an immutable Git source for the standalone bundle. */
export function buildIndex({ commit, publishTimestamp } = {}) {
  if (typeof commit !== 'string' || !/^[0-9a-f]{40}$/.test(commit)) throw new Error('commit must be a full 40-character lowercase Git SHA');
  if (!Number.isSafeInteger(publishTimestamp) || publishTimestamp <= 0 || publishTimestamp > 100_000_000_000) throw new Error('timestamp must be positive Unix seconds');
  checkedReleaseVersion(manifest.version, root.version, manifest.version);
  return {
    format: 1,
    title: { en: 'Bohrium plugins', zh: 'Bohrium 插件' },
    description: { en: 'Scientific workflows using the Bohrium CLI.', zh: '通过 Bohrium CLI 使用科研工作流。' },
    plugins: [{
      name: manifest.name,
      title: { en: 'Bohrium Skills', zh: 'Bohrium 科研技能' },
      description: { en: manifest.description, zh: '完整的 Bohrium 科研技能与 setup 引导；需要单独安装并登录 bohr。' },
      icon: 'icon.svg',
      author: manifest.author,
      homepage: manifest.homepage,
      repository: { url: 'https://github.com/dptech-corp/bohrium-skills', directory: 'skills' },
      versions: [{
        version: manifest.version,
        publishTimestamp,
        source: { type: 'git', url: 'https://github.com/dptech-corp/bohrium-skills', commit, path: 'skills' },
      }],
    }],
  };
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  try {
    const { values } = parseArgs({ options: { commit: { type: 'string' }, timestamp: { type: 'string' }, output: { type: 'string' } } });
    if (!values.output) throw new Error('--output is required');
    const index = buildIndex({ commit: values.commit, publishTimestamp: Number(values.timestamp) });
    fs.mkdirSync(path.dirname(values.output), { recursive: true });
    fs.writeFileSync(values.output, `${JSON.stringify(index, null, 2)}\n`);
    fs.copyFileSync(new URL('../skills/icon.svg', import.meta.url), path.join(path.dirname(values.output), 'icon.svg'));
    console.log(values.output);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
