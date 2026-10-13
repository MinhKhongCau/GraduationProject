"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import { isNavItemActive, type NavItem } from "@/constants/nav";

export interface SideRailNavProps {
  navItems: NavItem[];
}

/** Narrow icon rail beside the page content on desktop (tablet/mobile use the drawer and bottom tabs). */
export function SideRailNav({ navItems }: SideRailNavProps) {
  const pathname = usePathname();
  const { t } = useTranslation();

  return (
    <nav
      aria-label="Sections"
      className="sticky top-24 hidden w-14 shrink-0 flex-col items-center gap-1 self-start rounded-xl border border-border bg-background py-3 lg:flex"
    >
      {navItems.map((item) => {
        const isActive = isNavItemActive(item, pathname);
        const Icon = item.icon;
        const label = t(item.labelKey, item.label);
        return (
          <Link
            key={item.id}
            href={item.href}
            title={label}
            aria-label={label}
            aria-current={isActive ? "page" : undefined}
            className={clsx(
              "flex h-10 w-10 items-center justify-center rounded-lg transition-colors",
              isActive
                ? "bg-primary-soft text-primary-soft-text"
                : "text-muted-foreground hover:bg-surface hover:text-foreground"
            )}
          >
            <Icon className="h-5 w-5" aria-hidden="true" />
          </Link>
        );
      })}
    </nav>
  );
}
