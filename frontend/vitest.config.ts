import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';
import vuetify from 'vite-plugin-vuetify';

export default defineConfig({
  plugins: [
    vue(),
    vuetify(),
  ],
  test: {
    globals: true,
    environment: 'happy-dom',
    include: ['**/*.spec.ts'],
    setupFiles: ['./src/test/setup.ts'],
    deps: {
      inline: ['vuetify'],
    },
    css: true,
  },
  resolve: {
    alias: {
      '@': '/src',
    },
  },
});