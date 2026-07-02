import type { ReactNode } from "react";
import { ClientShell } from "@/components/layout/client-shell";
import { ProtectedRoute } from "@/router";
import { EXPERT_NAV_ITEMS, EXPERT_BOTTOM_NAV_ITEMS, SETTINGS_NAV_ITEM } from "@/constants";

export default function ExpertLayout({ children }: { children: ReactNode }) {
  return (
    <ProtectedRoute allow={["EXPERT"]}>
      <ClientShell
        navItems={EXPERT_NAV_ITEMS}
        bottomNavItems={EXPERT_BOTTOM_NAV_ITEMS}
        settingsItem={SETTINGS_NAV_ITEM.expert}
      >
        {children}
      </ClientShell>
    </ProtectedRoute>
  );
}
