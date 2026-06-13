import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import { visualizer } from 'rollup-plugin-visualizer'
import viteCompression from 'vite-plugin-compression'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    preact(),
    tailwindcss(),
    visualizer({
      open: false,
      filename: 'stats.html'
    }),
    viteCompression({
      deleteOriginFile: true
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      react: 'preact/compat',
      'react-dom/test-utils': 'preact/test-utils',
      'react-dom/client': 'preact/compat/client',
      'react-dom': 'preact/compat',
      'react/jsx-runtime': 'preact/jsx-runtime',
      'react/jsx-dev-runtime': 'preact/jsx-dev-runtime',
    },
  },
  build: {
    target: 'esnext',
    outDir: '../app/infra/static/web-dist',
    minify: 'oxc',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/preact')) {
            return 'vendor-preact'
          }
          if (id.includes('node_modules/@cloudflare/kumo') || id.includes('node_modules/@base-ui')) {
            return 'vendor-kumo'
          }
          if (id.includes('node_modules/luckysheet') || id.includes('node_modules/luckyexcel')) {
            return 'vendor-luckysheet'
          }
          if (id.includes('node_modules/xlsx') || id.includes('node_modules/docx-preview')) {
            return 'vendor-office'
          }
          if (id.includes('node_modules/lexical') || id.includes('node_modules/@lexical')) {
            return 'vendor-lexical'
          }
          if (id.includes('node_modules/prosekit') || id.includes('node_modules/@xyflow')) {
            return 'vendor-editors'
          }
          if (id.includes('node_modules/@tanstack/react-query')) {
            return 'vendor-query'
          }
          if (id.includes('node_modules/@phosphor-icons')) {
            return 'vendor-icons'
          }
          if (id.includes('node_modules')) {
            return 'vendor-common'
          }
        }
      }
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:18080'
      }
    }
  }
})
