import { defineConfig } from 'vite'
import topLevelAwait from 'vite-plugin-top-level-await'; // 正确导入方式
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue(), topLevelAwait()],
  build: {
    target: 'esnext', // 强制覆盖
    minify: false, // 禁用压缩
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  }
})
