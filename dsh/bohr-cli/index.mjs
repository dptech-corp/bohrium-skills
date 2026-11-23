/**
 * Public DSH adapter for the Bohrium CLI, carrying no install scripts.
 *
 * It depends on the six published platform packages — the same binaries the
 * `@dptech-corp/bohr-cli` launcher runs — so a profile install pulls exactly
 * one native `bohr` binary for the current host and nothing that pnpm's build
 * policy has to approve. Guides stay inside the binary and are read through
 * `bohr guide`; no Bohrium documentation is copied here.
 *
 * Registers no tool: every command keeps running through the host's own shell
 * tool and its approval flow.
 */
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';

export const name = 'bohr-cli';
export const inject = ['shellEnv', 'skills'];

const PLATFORMS = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
const ARCHITECTURES = { x64: 'amd64', arm64: 'arm64' };
const invoke = '"$DSH_BOHR_CLI"';
const content = [
  `The Bohrium CLI installed in this profile is \`bohr\`. Run it through the shell tool as \`${invoke} <command>\`.`,
  `Find its embedded guides with \`${invoke} guide list -o json\`, read one with \`${invoke} guide read <name> -o json\`, and a file it references with \`${invoke} guide read <name> <relative-path> -o json\`.`,
  'This bundle installs the CLI binary and its guides only. Companion programs are optional and stay with the CLI: install one with `bohr extension install <tool>` (macOS Lobster Web: `bohr agents local install`) when a workflow asks for it.',
].join('\n\n');

/** Absolute path of the native launcher installed for this host. */
function launcher() {
  const platform = PLATFORMS[process.platform];
  const arch = ARCHITECTURES[process.arch];
  if (platform === undefined || arch === undefined) {
    throw new Error(`no Bohrium CLI binary for ${process.platform}-${process.arch}`);
  }
  const manifest = createRequire(import.meta.url)
    .resolve(`@dptech-corp/bohr-cli-${platform}-${arch}/package.json`);
  return join(dirname(manifest), 'bin', process.platform === 'win32' ? 'bohr.exe' : 'bohr');
}

export function apply(ctx) {
  let cli;
  try {
    cli = launcher();
  } catch (error) {
    // A host without a matching platform package must not take the whole
    // profile down with it.
    ctx.logger?.warn?.(`[bohr-cli] ${String(error)}`);
    return;
  }
  ctx.shellEnv.register({
    name,
    variables: {
      DSH_BOHR_CLI: { description: 'Absolute path of the native bohr binary this bundle installed for the current host.' },
    },
    resolve: () => ({ DSH_BOHR_CLI: cli }),
  });
  ctx.skills.register({
    name,
    description: 'Run the Bohrium CLI installed in this profile; the guides it carries hold the platform workflows.',
    content,
    source: 'bundled',
    provider: name,
  });
}
