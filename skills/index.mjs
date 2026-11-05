import { fileURLToPath } from 'node:url';
import { FileSystemSkillProvider } from '@deepseek-ai/dsh-skill-filesystem';

export const name = 'bohrium-skills';
export const inject = ['skills'];

/** Contribute the existing complete Skills without scanning or writing user directories. */
export function apply(ctx) {
  let provider;
  ctx.skills.registerProvider((control) => {
    provider = new FileSystemSkillProvider(ctx, control, {
      providerName: 'bohrium',
      includeDefaultRoots: false,
      bundledSkillDir: fileURLToPath(new URL('./', import.meta.url)),
      watch: false,
    });
    return provider;
  });
  ctx.effect(() => () => provider.dispose());
}
