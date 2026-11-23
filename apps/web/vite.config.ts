import { fileURLToPath } from "node:url"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vitest/config"

// En desarrollo, /api y /.well-known van al backend local (puerto 5161), igual que hace
// el spa-server en la imagen: el navegador ve un solo origen y la cookie funciona sin CORS.
const backendLocal = "http://127.0.0.1:5161"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  server: {
    port: 5160,
    strictPort: true,
    proxy: {
      "/api": backendLocal,
      "/.well-known": backendLocal,
    },
  },
  preview: { port: 5160, strictPort: true },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/pruebas/configuracion.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
    coverage: {
      provider: "v8",
      reporter: ["text-summary", "lcov"],
      reportsDirectory: "coverage",
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.{ts,tsx}",
        "src/components/ui/**",
        "src/api/esquema.d.ts",
        "src/mocks/**",
        "src/pruebas/**",
        "src/main.tsx",
        "src/app-desarrollo.tsx",
      ],
    },
  },
})
