/// <reference types="vitest/config" />
import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig, loadEnv } from "vite"

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "")
  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      // Port doluysa sessizce 5174'e kayma: Keycloak webOrigins 5173'e bağlı,
      // kayma CORS hatası üretir. Dolu ise açıkça hata ver.
      strictPort: true,
      proxy: {
        // The SPA always calls the Go API with relative /api/... URLs; the dev
        // server forwards them so no CORS is needed on the backend. The target
        // follows VITE_API_URL when the API runs on a non-default port (ADDR).
        "/api": env.VITE_API_URL || "http://localhost:8080",
      },
    },
    test: {
      environment: "jsdom",
      globals: true,
      setupFiles: "./src/test/setup.ts",
    },
  }
})
