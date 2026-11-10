import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    }
  },
  // 👇 只有下面这一小段是 Vite 的跨域代理配置
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080', // 遇到 /api 开头的请求，转发给你的 Go 后端
        changeOrigin: true
      }
    }
  }
})