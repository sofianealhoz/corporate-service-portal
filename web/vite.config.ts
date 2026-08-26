import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// En dev le front tourne sur :5173 et l'API sur :8080.
// On proxifie /api vers l'API pour rester en same-origin (pas de CORS),
// comme en prod ou le back sert le front.
// API_PROXY_TARGET permet de suivre l'API si PORT a ete change.
const target = process.env.API_PROXY_TARGET ?? 'http://localhost:8080'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': target,
    },
  },
})
