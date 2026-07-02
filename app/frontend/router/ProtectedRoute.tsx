"use client";

import { useEffect, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { Spinner } from "@/components/ui/Spinner";
import { ROUTES } from "@/constants/route";
import type { UserRole } from "@/types";

export interface ProtectedRouteProps {
  allow: UserRole[];
  children: ReactNode;
}

/**
 * The authoritative client-side guard — proxy.ts only does a coarse,
 * cookie-based redirect before render. See DESIGN.md "Auth/session storage".
 */
export function ProtectedRoute({ allow, children }: ProtectedRouteProps) {
  const { user, isLoading, isAuthenticated } = useAuthContext();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (isLoading) return;
    if (!isAuthenticated) {
      router.replace(`${ROUTES.AUTH.LOGIN}?redirect=${encodeURIComponent(pathname)}`);
      return;
    }
    if (user && !allow.includes(user.role)) {
      router.replace(ROUTES.FORBIDDEN);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, isAuthenticated, user?.role]);

  if (isLoading || !isAuthenticated || (user && !allow.includes(user.role))) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner className="h-8 w-8" />
      </div>
    );
  }

  return <>{children}</>;
}
