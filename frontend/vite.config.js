import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import history from 'connect-history-api-fallback'

export default defineConfig({
  plugins: [svelte()],
  server: {
    fs: {
      allow: ['.'],
    },
    middlewareMode: false,
    setupMiddlewares: (middlewares) => {
      middlewares.use(history())
      return middlewares
    }
  }
})