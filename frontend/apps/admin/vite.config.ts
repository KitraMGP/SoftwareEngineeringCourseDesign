import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [vue()],
    resolve: {
      dedupe: ['vue']
    },
    optimizeDeps: {
      exclude: ['@private-kb/shared']
    },
    server: {
      host: '0.0.0.0',
      port: Number(env.VITE_PORT || 5174),
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080',
          changeOrigin: true
        }
      }
    },
    build: {
      outDir: 'dist'
    }
  };
});
