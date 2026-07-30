import { defineConfig, devices } from "@playwright/test";

const MOCK_BACKEND_URL = "http://127.0.0.1:3101";
const FRONTEND_URL = "http://localhost:3210";

export default defineConfig({
  testDir: "./e2e",
  use: {
    baseURL: FRONTEND_URL,
    trace: "on-first-retry",
  },
  webServer: [
    {
      name: "mock-backend",
      command: "node e2e/support/mock-backend.mjs",
      url: `${MOCK_BACKEND_URL}/health`,
      reuseExistingServer: false,
    },
    {
      name: "next",
      command: "npm run dev -- --port 3210",
      env: { API_BASE_URL: MOCK_BACKEND_URL },
      url: `${FRONTEND_URL}/login`,
      reuseExistingServer: false,
    },
  ],
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
