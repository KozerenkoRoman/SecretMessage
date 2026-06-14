import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import pkg from './package.json'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  define: {
    '__APP_VERSION__': JSON.stringify(pkg.version)
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/ws': {
        target: 'http://localhost:3000',
        ws: true, // ЦЕЙ РЯДОК ДОЗВОЛЯЄ ПРОКСЮВАТИ ВЕБСОКЕТИ
      },
    },
  }
})