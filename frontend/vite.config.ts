import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';
import path from 'node:path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      // The icons live in public/ and are copied verbatim into dist/.
      includeAssets: ['favicon.svg', 'pwa-icon.svg', 'mask-icon.svg'],
      manifest: {
        name: 'finance.sh',
        short_name: 'finance.sh',
        description: 'Controle financeiro self-hosted',
        lang: 'pt-BR',
        theme_color: '#10b981',
        background_color: '#0f1115',
        display: 'standalone',
        start_url: '/',
        scope: '/',
        orientation: 'portrait',
        icons: [
          {
            src: '/pwa-icon.svg',
            sizes: '192x192',
            type: 'image/svg+xml',
            purpose: 'any',
          },
          {
            src: '/pwa-icon.svg',
            sizes: '512x512',
            type: 'image/svg+xml',
            purpose: 'any',
          },
          {
            // Maskable variant: SVG with safe-zone padding so Android can
            // clip it to any device-specific shape without cropping the mark.
            src: '/pwa-icon-maskable.svg',
            sizes: '512x512',
            type: 'image/svg+xml',
            purpose: 'maskable',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,woff2}'],
        // SPA shell: any uncached navigation falls back to index.html...
        navigateFallback: '/index.html',
        // ...EXCEPT API calls — they must never be answered with the app shell.
        navigateFallbackDenylist: [/^\/api\//],
        cleanupOutdatedCaches: true,
        clientsClaim: true,
        runtimeCaching: [
          {
            // Financial data must always be fresh: never cache /api responses.
            urlPattern: ({ url }) => url.pathname.startsWith('/api/'),
            handler: 'NetworkOnly',
            method: 'GET',
            options: { cacheName: 'api-no-cache' },
          },
          {
            // Google Fonts stylesheets + font files: safe to cache long-term.
            urlPattern: ({ url }) =>
              url.origin === 'https://fonts.googleapis.com' ||
              url.origin === 'https://fonts.gstatic.com',
            handler: 'CacheFirst',
            options: {
              cacheName: 'google-fonts',
              expiration: { maxEntries: 20, maxAgeSeconds: 60 * 60 * 24 * 365 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
        ],
      },
      devOptions: { enabled: false },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        // Split large vendor libraries into dedicated chunks so the initial
        // bundle stays small and chunks are cached independently.
        manualChunks: {
          react: ['react', 'react-dom', 'react-router-dom'],
          recharts: ['recharts'],
          'data-vendor': ['@tanstack/react-query', 'axios', 'zustand'],
          'form-vendor': ['react-hook-form', '@hookform/resolvers', 'zod'],
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        // Backend (make backend-dev) listens on APP_PORT=8090 in dev.
        target: 'http://localhost:8090',
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
