"use client";

import type { ReactNode } from "react";
import { BrandLogo } from "@/components/ui";
import { DesktopHeaderNav } from "./DesktopHeaderNav";
import { TabletDrawerNav } from "./TabletDrawerNav";
import { MobileBottomNav } from "./MobileBottomNav";
import { NavAvatarMenu } from "./component/NavAvatarMenu";
import type { NavItem } from "@/constants/nav";

export interface ClientShellProps {
  navItems: NavItem[];
  bottomNavItems: NavItem[];
  settingsItem: NavItem;
  children: ReactNode;
}

/**
 * The shared "client" responsive nav shell used by both the patient and
 * expert portals — desktop header, tablet left drawer, mobile bottom tabs.
 * Switching is CSS-first (Tailwind sm/lg breakpoints) to avoid a
 * hydration flash. See DESIGN.md "Responsive client shell".
 */
export function ClientShell({ navItems, bottomNavItems, settingsItem, children }: ClientShellProps) {
  return (
    <div className="flex min-h-dvh flex-col bg-surface">
      <DesktopHeaderNav navItems={navItems} settingsItem={settingsItem} />
      <TabletDrawerNav navItems={navItems} settingsItem={settingsItem} />

      <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border bg-background/95 px-4 pt-[env(safe-area-inset-top)] backdrop-blur sm:hidden">
<BrandLogo className="text-base" />
        <NavAvatarMenu settingsItem={settingsItem} />
      </header>

      <main className="mx-auto w-full max-w-7xl flex-1 px-4 pt-5 pb-24 sm:px-6 sm:py-6 lg:px-8 lg:py-8">{children}</main>

      <MobileBottomNav navItems={bottomNavItems} />
    </div>
  );
}
