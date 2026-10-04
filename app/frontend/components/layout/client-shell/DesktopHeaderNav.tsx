"use client";

import Link from "next/link";
import { BrandLogo } from "@/components/ui";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import { LanguageSwitcher } from "../LanguageSwitcher";
import { WalletBadge } from "./component/WalletBadge";
import { NavAvatarMenu } from "./component/NavAvatarMenu";
import type { NavItem } from "@/constants/nav";

export interface DesktopHeaderNavProps {
  navItems: NavItem[];
  settingsItem: NavItem;
}

export function DesktopHeaderNav({ navItems, settingsItem }: DesktopHeaderNavProps) {
  const pathname = usePathname();
  const { t } = useTranslation();

  return (
    <header className="sticky top-0 z-30 hidden border-b border-border bg-background/95 backdrop-blur lg:block">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-8">
        <div className="flex items-center gap-8">
          <BrandLogo />
          <nav className="flex items-center gap-1">
            {navItems.map((item) => {
              const isActive = pathname === item.href;
              const Icon = item.icon;
              return (
                <Link
                  key={item.id}
                  href={item.href}
                  aria-current={isActive ? "page" : undefined}
                  className={clsx(
                    "flex h-9 items-center gap-1.5 rounded-lg px-3 text-sm font-medium transition-colors",
                    isActive ? "bg-primary-soft text-primary-soft-text" : "text-muted-foreground hover:bg-surface hover:text-foreground"
                  )}
                >
                  <Icon className="h-4 w-4" />
                  {t(item.labelKey, item.label)}
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="flex items-center gap-3">
          <LanguageSwitcher />
          <WalletBadge />
          <NavAvatarMenu settingsItem={settingsItem} />
        </div>
      </div>
    </header>
  );
}
