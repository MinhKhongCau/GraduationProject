"use client";

import type { ReactNode } from "react";
import { ClientShell } from "@/components/layout/client-shell";
import { ProtectedRoute } from "@/router";
import { PATIENT_HEADER_NAV_ITEMS, PATIENT_BOTTOM_NAV_ITEMS, SETTINGS_NAV_ITEM } from "@/constants";

export default function PatientLayout({ children }: { children: ReactNode }) {
  return (
    <ProtectedRoute allow={["PATIENT"]}>
      <ClientShell
        navItems={PATIENT_HEADER_NAV_ITEMS}
        bottomNavItems={PATIENT_BOTTOM_NAV_ITEMS}
        settingsItem={SETTINGS_NAV_ITEM.patient}
      >
        {children}
      </ClientShell>
    </ProtectedRoute>
  );
}
