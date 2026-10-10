import { ENV } from "../../../config/env.js";

// Caches the internal (M2M) JWT issued by auth-service's POST /internal/auth/token
// and refreshes it one minute before it expires — same contract as the Go
// services' TokenManager.
export function createTokenManager({
  authUrl = ENV.AUTH_SERVICE_INTERNAL_URL,
  clientId = ENV.INTERNAL_CLIENT_ID,
  clientSecret = ENV.INTERNAL_CLIENT_SECRET,
  fetchImpl = fetch,
  now = Date.now,
} = {}) {
  let token = "";
  let expiresAt = 0;
  let inflight = null;

  async function fetchToken() {
    const res = await fetchImpl(`${authUrl}/internal/auth/token`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ clientId, clientSecret }),
    });
    if (!res.ok) {
      throw new Error(`internal_auth: auth-service returned ${res.status}`);
    }
    const body = await res.json();
    token = body.access_token;
    expiresAt = now() + Number(body.expires_in || 0) * 1000;
    return token;
  }

  return {
    async getToken() {
      if (token && now() + 60_000 < expiresAt) return token;
      if (!inflight) {
        inflight = fetchToken().finally(() => {
          inflight = null;
        });
      }
      return inflight;
    },
    invalidate() {
      token = "";
      expiresAt = 0;
    },
  };
}
