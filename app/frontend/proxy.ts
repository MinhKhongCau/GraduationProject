import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { resolveRequiredRole, isAuthRoute, dashboardForRole } from "@/router/routes.config";
import { SESSION_COOKIE, ROLE_COOKIE } from "@/constants/api";
import { ROUTES } from "@/constants/route";
import type { UserRole } from "@/types";

/**
 * Coarse, cookie-based guard only (see DESIGN.md "Auth/session storage") —
 * the real access token lives in localStorage, invisible here. The
 * authoritative check is router/ProtectedRoute.tsx on the client.
 */
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasSession = request.cookies.get(SESSION_COOKIE)?.value === "1";
  const role = request.cookies.get(ROLE_COOKIE)?.value as UserRole | undefined;

  const requiredRole = resolveRequiredRole(pathname);

  if (requiredRole) {
    if (!hasSession) {
      const loginUrl = new URL(ROUTES.AUTH.LOGIN, request.url);
      loginUrl.searchParams.set("redirect", pathname);
      return NextResponse.redirect(loginUrl);
    }
    if (role && role !== requiredRole) {
      return NextResponse.redirect(new URL(ROUTES.FORBIDDEN, request.url));
    }
  }

  if (hasSession && role && isAuthRoute(pathname)) {
    return NextResponse.redirect(new URL(dashboardForRole(role), request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/patient/:path*", "/expert/:path*", "/admin/:path*", "/auth/:path*"],
};
