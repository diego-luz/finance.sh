import { defineConfig } from 'vitest/config';
import path from 'node:path';

// Unit tests (Vitest + happy-dom). Kept apart from vite.config.ts so the PWA
// plugin does not run for tests.
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    environment: 'happy-dom',
    include: ['src/**/*.test.ts'],
  },
});
