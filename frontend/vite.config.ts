import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'service-worker.js',
      injectRegister: false,
    }),
  ],
  server: {
    port: 3000,
    // Proxy /api/* requests to the local Go backend so the frontend behaves
    // exactly like production (same-origin API calls, no CORS).
    proxy: {
      // Matches PORT in .env (default 8090 locally to avoid the common 8080 clash).
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
});
