import type { ReactNode } from "react";
import { AdminSidebar } from "@/components/layout/admin-shell/AdminSidebar";
import { AdminTopbar } from "@/components/layout/admin-shell/AdminTopbar";
import { ProtectedRoute } from "@/router";

export default function AdminLayout({ children }: { children: ReactNode }) {
  return (
    <ProtectedRoute allow={["ADMIN"]}>
      <div className="flex min-h-screen bg-surface">
        <AdminSidebar />
        <div className="flex min-w-0 flex-1 flex-col">
          <AdminTopbar />
          <main className="mx-auto w-full max-w-7xl flex-1 p-4 sm:p-6 lg:p-8">{children}</main>
        </div>
      </div>
    </ProtectedRoute>
  );
}
