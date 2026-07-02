/**
 * Google's One Tap/GoogleLogin credential is a JWT; we only need the email
 * claim to complete our own LoginResponse (which doesn't include it — see
 * types/auth.ts). Not worth a jwt-decode dependency for one field.
 */
export function decodeGoogleEmail(idToken: string): string {
  const payload = idToken.split(".")[1];
  const decoded = JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/")));
  return decoded.email ?? "";
}
