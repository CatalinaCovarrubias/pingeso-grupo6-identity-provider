import { defineConfig, devices } from "@playwright/test"

// Por defecto la prueba de humo corre contra la app de desarrollo con MSW (pnpm dev).
// Con E2E_URL apunta a otra instancia, por ejemplo la del Compose:
//   E2E_URL=http://127.0.0.1:5160 pnpm test:e2e
const urlExterna = process.env.E2E_URL
const enCI = Boolean(process.env.CI)

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: enCI,
  retries: enCI ? 1 : 0,
  reporter: enCI ? [["list"], ["html", { open: "never" }]] : "list",
  use: {
    baseURL: urlExterna ?? "http://localhost:5160",
    locale: "es-CL",
    timezoneId: "America/Santiago",
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: urlExterna
    ? undefined
    : {
        command: "pnpm dev",
        url: "http://localhost:5160",
        reuseExistingServer: !enCI,
        timeout: 60_000,
      },
})
