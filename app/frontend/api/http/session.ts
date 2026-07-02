import { STORAGE_KEYS, SESSION_COOKIE, ROLE_COOKIE } from "@/constants/api";
import type { UserRole } from "@/types";

/**
 * Plain (non-React) session storage shared by the axios interceptors and
 * AuthContext, so there is one source of truth for where tokens live.
 * See DESIGN.md "Auth/session storage" for the localStorage + non-httpOnly
 * cookie tradeoff this implements.
 */

function isBrowser() {
  return typeof window !== "undefined";
}

export function getAccessToken(): string | null {
  if (!isBrowser()) return null;
  return window.localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN);
}

export function getRefreshToken(): string | null {
  if (!isBrowser()) return null;
  return window.localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);
}

export function getStoredRole(): UserRole | null {
  if (!isBrowser()) return null;
  return window.localStorage.getItem(STORAGE_KEYS.ROLE) as UserRole | null;
}

export interface SessionPayload {
  accessToken: string;
  refreshToken: string;
  accountId: string;
  fullName: string;
  email: string;
  role: UserRole;
}

export function getStoredUser(): { id: string; fullName: string; email: string; role: UserRole } | null {
  if (!isBrowser()) return null;
  const id = window.localStorage.getItem(STORAGE_KEYS.ACCOUNT_ID);
  const fullName = window.localStorage.getItem(STORAGE_KEYS.FULL_NAME);
  const email = window.localStorage.getItem(STORAGE_KEYS.EMAIL);
  const role = getStoredRole();
  if (!id || !fullName || !email || !role) return null;
  return { id, fullName, email, role };
}

export function setSession(payload: SessionPayload) {
  if (!isBrowser()) return;
  window.localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, payload.accessToken);
  window.localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, payload.refreshToken);
  window.localStorage.setItem(STORAGE_KEYS.ROLE, payload.role);
  window.localStorage.setItem(STORAGE_KEYS.ACCOUNT_ID, payload.accountId);
  window.localStorage.setItem(STORAGE_KEYS.FULL_NAME, payload.fullName);
  window.localStorage.setItem(STORAGE_KEYS.EMAIL, payload.email);
  // Presence + role only, no token value — routing convenience for proxy.ts,
  // not a security boundary. See DESIGN.md.
  document.cookie = `${SESSION_COOKIE}=1; path=/; max-age=${60 * 60 * 24 * 7}; samesite=lax`;
  document.cookie = `${ROLE_COOKIE}=${payload.role}; path=/; max-age=${60 * 60 * 24 * 7}; samesite=lax`;
}

export function setAccessToken(accessToken: string) {
  if (!isBrowser()) return;
  window.localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, accessToken);
}

export function clearSession() {
  if (!isBrowser()) return;
  Object.values(STORAGE_KEYS).forEach((key) => window.localStorage.removeItem(key));
  document.cookie = `${SESSION_COOKIE}=; path=/; max-age=0`;
  document.cookie = `${ROLE_COOKIE}=; path=/; max-age=0`;
}
