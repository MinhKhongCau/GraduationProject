import type { UserRole } from "@/types";

export const PUBLIC_PREFIXES = ["/", "/auth", "/forbidden", "/not-found"] as const;

export const ROLE_PREFIXES: Record<string, UserRole> = {
  "/patient": "PATIENT",
  "/expert": "EXPERT",
  "/admin": "ADMIN",
};

/** Returns the role required for a pathname, or null if it's public. */
export function resolveRequiredRole(pathname: string): UserRole | null {
  for (const [prefix, role] of Object.entries(ROLE_PREFIXES)) {
    if (pathname === prefix || pathname.startsWith(`${prefix}/`)) {
      return role;
    }
  }
  return null;
}

export function isAuthRoute(pathname: string): boolean {
  return pathname === "/auth" || pathname.startsWith("/auth/");
}

export function dashboardForRole(role: UserRole): string {
  switch (role) {
    case "PATIENT":
      return "/patient/dashboard";
    case "EXPERT":
      return "/expert/dashboard";
    case "ADMIN":
      return "/admin/dashboard";
  }
}
