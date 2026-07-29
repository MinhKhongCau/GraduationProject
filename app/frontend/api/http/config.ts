/**
 * Public because Axios runs in the browser/WebView. A remote Capacitor shell
 * therefore uses the same HTTPS API origin as the deployed web application.
 */
function readApiBaseUrl(): string {
  const value = process.env.NEXT_PUBLIC_API_URL;
  if (!value) {
    throw new Error(
      "Missing NEXT_PUBLIC_API_URL. Set it to the Kong gateway URL, for example https://api.myapp.example.com/api/v1."
    );
  }

  let url: URL;
  try {
    url = new URL(value);
  } catch {
    throw new Error("NEXT_PUBLIC_API_URL must be an absolute HTTP(S) URL.");
  }

  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error("NEXT_PUBLIC_API_URL must use the http or https protocol.");
  }

  if (process.env.NODE_ENV === "production" && url.protocol !== "https:") {
    throw new Error("NEXT_PUBLIC_API_URL must use HTTPS in production, including Capacitor release builds.");
  }

  return value.replace(/\/+$/, "");
}

export const API_BASE_URL = readApiBaseUrl();
