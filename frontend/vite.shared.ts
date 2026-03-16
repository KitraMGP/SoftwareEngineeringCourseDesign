import { loadEnv, type UserConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export function createAppViteConfig(mode: string, defaultPort: number): UserConfig {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [vue()],
    resolve: {
      dedupe: ['vue']
    },
    optimizeDeps: {
      exclude: ['@private-kb/shared']
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler'
        }
      }
    },
    server: {
      host: '0.0.0.0',
      port: Number(env.VITE_PORT || defaultPort),
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
}
