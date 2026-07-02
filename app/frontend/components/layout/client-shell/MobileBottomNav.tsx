"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import type { NavItem } from "@/constants/nav";

export interface MobileBottomNavProps {
  navItems: NavItem[];
}

export function MobileBottomNav({ navItems }: MobileBottomNavProps) {
  const pathname = usePathname();
  const { t } = useTranslation();

  return (
    <nav className="fixed inset-x-0 bottom-0 z-30 flex border-t border-border bg-background sm:hidden">
      {navItems.map((item) => {
        const isActive = pathname === item.href;
        const Icon = item.icon;
        return (
          <Link
            key={item.id}
            href={item.href}
            className={clsx(
              "flex flex-1 flex-col items-center gap-0.5 py-2.5 text-[11px] font-medium",
              isActive ? "text-primary" : "text-muted-foreground"
            )}
          >
            <Icon className="h-5 w-5" />
            {t(item.labelKey, item.label)}
          </Link>
        );
      })}
    </nav>
  );
}
