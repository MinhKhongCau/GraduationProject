import type { CapacitorConfig } from "@capacitor/cli";
import { KeyboardResize, KeyboardStyle } from "@capacitor/keyboard";

/**
 * Remote Server Shell configuration.
 *
 * Set CAPACITOR_ENV=production for a release sync/build. CI can also provide
 * CAPACITOR_PRODUCTION_URL without changing this file. Development retains a
 * localhost default, with CAPACITOR_DEV_SERVER_URL available for a LAN IP or
 * emulator-specific host address.
 */
const isProduction = process.env.CAPACITOR_ENV === "production";
const productionUrl = process.env.CAPACITOR_PRODUCTION_URL ?? "https://myapp.example.com";
const developmentUrl = process.env.CAPACITOR_DEV_SERVER_URL ?? "http://localhost:3000";
const serverUrl = isProduction ? productionUrl : developmentUrl;
const navigationHost = new URL(serverUrl).hostname;

const config: CapacitorConfig = {
  appId: "com.mindcare.app",
  appName: "MindCare",

  // Required by Capacitor tooling even though this remote-shell app loads its
  // runtime UI from `server.url`, not from a bundled static export.
  webDir: "public",

  // Keep pinch/double-tap zoom disabled on both native WebViews.
  zoomEnabled: false,

  server: {
    // Release builds must load the HTTPS production site. The localhost value
    // is intentionally development-only; override it with a reachable LAN URL
    // when testing on physical devices.
    url: serverUrl,
    cleartext: !isProduction,

    // Android must use a standards-based scheme for history routing.
    androidScheme: "https",

    // `iosScheme` is only used for bundled local assets. It must not be http
    // or https because WKWebView reserves those schemes; `capacitor` is the
    // safe default and does not change the remote HTTPS server URL above.
    iosScheme: "capacitor",

    // Only the app's own host may remain in the in-app WebView. Other links
    // retain Capacitor's default behavior and open outside the WebView.
    allowNavigation: [navigationHost],
  },

  ios: {
    // Keep long-press destination previews available and request the mobile
    // rendering mode from WKWebView.
    allowsLinkPreview: true,
    preferredContentMode: "mobile",
    zoomEnabled: false,
  },

  android: {
    // HTTP/mixed content is allowed only during local development. Never set
    // this to true for the HTTPS production shell.
    allowMixedContent: !isProduction,
    zoomEnabled: false,

    // Hardware acceleration is enabled by Android's Capacitor activity by
    // default. If a future native customization changes it, set
    // android:hardwareAccelerated="true" in AndroidManifest.xml; Capacitor
    // exposes no capacitor.config.ts field for that Android manifest setting.
  },

  plugins: {
    // Keep the WebView viewport stable while the chat layout subtracts the
    // keyboard height reported by @capacitor/keyboard.
    Keyboard: {
      resize: KeyboardResize.None,
      resizeOnFullScreen: true,
      style: KeyboardStyle.Light,
    },

    // Requires `@capacitor/splash-screen` after native platforms are added.
    SplashScreen: {
      launchShowDuration: 2000,
      launchAutoHide: true,
      backgroundColor: "#ffffff",
      showSpinner: false,
    },

    // Requires `@capacitor/status-bar`. "DARK" gives dark status-bar content
    // on the application's light background.
    StatusBar: {
      overlaysWebView: false,
      style: "DARK",
      backgroundColor: "#ffffff",
    },
  },
};

export default config;
