import { fileURLToPath, URL } from 'node:url';

import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      vue: fileURLToPath(
        new URL('./apps/user/node_modules/vue/dist/vue.runtime.esm-bundler.js', import.meta.url)
      )
    }
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['packages/shared/src/**/*.test.ts', 'apps/**/src/**/*.test.ts']
  }
});
