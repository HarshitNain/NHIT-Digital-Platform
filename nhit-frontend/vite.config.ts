import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBase = env.VITE_API_GATEWAY_URL || (mode === 'production' ? 'https://nhit-api-gateway.onrender.com' : '')

  return {
    plugins: [react()],
    // Expose VITE_API_GATEWAY_URL to the built bundle at compile time
    define: {
      '__API_BASE__': JSON.stringify(apiBase),
    },
    server: {
      port: 5173,
      proxy: {
        // In dev: proxy /api/* to local backend (no env var needed in dev)
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        }
      }
    }
  }
})
