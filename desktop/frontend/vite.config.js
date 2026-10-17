import { existsSync, readdirSync, rmSync } from 'node:fs';
import { defineConfig } from 'vite';

export default defineConfig({
  // The sidebar logo is the repository's assets/logo.svg, two folders above
  // this one; the dev server must be allowed to serve it.
  server: { fs: { allow: ['../..'] } },
  // dist/.gitkeep is committed so a clean checkout compiles main.go's embed;
  // keep it when the build empties the folder.
  build: { emptyOutDir: false },
  plugins: [
    {
      name: 'clean-dist-keep-placeholder',
      apply: 'build',
      buildStart() {
        const dist = new URL('./dist/', import.meta.url);
        if (!existsSync(dist)) return;
        for (const name of readdirSync(dist)) {
          if (name !== '.gitkeep') rmSync(new URL(name, dist), { recursive: true, force: true });
        }
      },
    },
  ],
});
