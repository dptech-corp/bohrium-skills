import fs from 'node:fs';
import { fileURLToPath } from 'node:url';

export function checkedReleaseVersion(requested, packageVersion, dshVersion = packageVersion) {
  const version = String(requested ?? '').replace(/^v/, '');
  if (!version || version !== packageVersion) {
    throw new Error(`Release version ${version || '(missing)'} does not match package.json ${packageVersion}; no assets should be published.`);
  }
  for (const bundleVersion of [].concat(dshVersion)) {
    if (bundleVersion !== packageVersion) {
      throw new Error(`DSH bundle version ${bundleVersion} does not match package.json ${packageVersion}; no assets should be published.`);
    }
  }
  return version;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const pkg = JSON.parse(fs.readFileSync(new URL('../package.json', import.meta.url), 'utf8'));
  const dsh = JSON.parse(fs.readFileSync(new URL('../skills/package.json', import.meta.url), 'utf8'));
  const cli = JSON.parse(fs.readFileSync(new URL('../dsh/bohr-cli/package.json', import.meta.url), 'utf8'));
  console.log(checkedReleaseVersion(process.argv[2], pkg.version, [dsh.version, cli.version]));
}
