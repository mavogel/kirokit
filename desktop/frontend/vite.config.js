import { defineConfig } from 'vite';

export default defineConfig({
  // The sidebar logo is the repository's assets/logo.svg, two folders above
  // this one; the dev server must be allowed to serve it.
  server: { fs: { allow: ['../..'] } },
});
