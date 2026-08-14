import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    // 将 /api 代理到 skill-hub 后端（默认 :8080）
    proxy: {
      '/api': {
        target: process.env.VITE_API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        // 允许 WebSocket 长连接（执行进度/日志实时推送）
        ws: true,
      },
    },
  },
})
