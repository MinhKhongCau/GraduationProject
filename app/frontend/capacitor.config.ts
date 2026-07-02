import type { CapacitorConfig } from "@capacitor/cli";

/**
 * No native ios/android projects yet (`npx cap add ios|android` deliberately
 * not run this pass). The mobile shell loads the deployed web app by URL
 * rather than a static export, so it stays compatible with next.config.ts's
 * `output: "standalone"` Docker deploy.
 */
const config: CapacitorConfig = {
  appId: "com.mindcare.app",
  appName: "MindCare",
  webDir: "public",
  server: {
    url: process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000",
    cleartext: true,
  },
};

export default config;
