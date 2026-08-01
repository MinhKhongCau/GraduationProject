"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { ROUTES } from "@/constants";
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
    <div className="flex min-h-screen flex-col bg-surface/30">
      <DesktopHeaderNav navItems={navItems} settingsItem={settingsItem} />
      <TabletDrawerNav navItems={navItems} settingsItem={settingsItem} />

      <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border bg-background/95 px-4 backdrop-blur sm:hidden">
        <Link href={ROUTES.HOME} className="flex items-center gap-2 text-base font-bold text-foreground">
          <span className="rounded-lg bg-primary p-1 text-white">🧠</span>
          MindCare
        </Link>
        <NavAvatarMenu settingsItem={settingsItem} />
      </header>

      <main className="flex-1 p-4 pb-20 sm:p-6 sm:pb-6 lg:p-8">{children}</main>

      <MobileBottomNav navItems={bottomNavItems} />
    </div>
  );
}
