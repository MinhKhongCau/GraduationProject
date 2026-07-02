"use client";

import Link from "next/link";
import { ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui";
import { useAuthContext } from "@/context/AuthContext";
import { ROUTES } from "@/constants";
import { dashboardForRole } from "@/router";

export default function ForbiddenPage() {
  const { user } = useAuthContext();

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-surface p-6 text-center">
      <ShieldAlert className="h-12 w-12 text-danger" />
      <h1 className="text-2xl font-bold text-foreground">You don&apos;t have access to this page</h1>
      <p className="max-w-md text-sm text-muted-foreground">
        Your account role doesn&apos;t have permission to view this section of MindCare.
      </p>
      <Link href={user ? dashboardForRole(user.role) : ROUTES.HOME}>
        <Button>Go to my dashboard</Button>
      </Link>
    </div>
  );
}
