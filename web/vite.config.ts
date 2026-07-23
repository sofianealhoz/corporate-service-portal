import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// En dev le front tourne sur :5173 et l'API sur :8080.
// On proxifie /api vers l'API pour rester en same-origin (pas de CORS),
// comme en prod ou le back sert le front.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
