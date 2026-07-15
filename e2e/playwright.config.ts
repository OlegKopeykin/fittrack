import { defineConfig, devices } from '@playwright/test'

// E2E гоняются против собранного бинаря (make build) со встроенной SPA —
// то, что реально едет в прод. Порт отдельный, чтобы не мешать dev-серверу.
const PORT = 8931

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: `http://127.0.0.1:${PORT}`,
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'iphone-15-pro', use: { ...devices['iPhone 15 Pro'] } },
    { name: 'desktop-chrome', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'rm -f e2e.db e2e.db-wal e2e.db-shm && ../bin/fittrack',
    url: `http://127.0.0.1:${PORT}/healthz`,
    env: {
      FITTRACK_ADDR: `127.0.0.1:${PORT}`,
      FITTRACK_DB: 'e2e.db',
      FITTRACK_E2E_SEED: '1',
    },
    reuseExistingServer: false,
    timeout: 15_000,
  },
})
