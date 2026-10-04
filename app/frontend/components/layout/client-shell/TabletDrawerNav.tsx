"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as Dialog from "@radix-ui/react-dialog";
import { Menu, X } from "lucide-react";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import { BrandLogo, BrandMark } from "@/components/ui";
import { LanguageSwitcher } from "../LanguageSwitcher";
import { NavAvatarMenu } from "./component/NavAvatarMenu";
import type { NavItem } from "@/constants/nav";

export interface TabletDrawerNavProps {
  navItems: NavItem[];
  settingsItem: NavItem;
}

export function TabletDrawerNav({ navItems, settingsItem }: TabletDrawerNavProps) {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  const { t } = useTranslation();

  return (
    <header className="sticky top-0 z-30 hidden border-b border-border bg-background/95 backdrop-blur sm:flex lg:hidden">
      <div className="flex h-16 w-full items-center justify-between px-4">
        <div className="flex items-center gap-3">
          <Dialog.Root open={open} onOpenChange={setOpen}>
            <Dialog.Trigger asChild>
              <button
                type="button"
                className="inline-flex h-10 w-10 items-center justify-center rounded-lg text-foreground transition-colors hover:bg-surface"
                aria-label="Open navigation menu"
              >
                <Menu className="h-5 w-5" />
              </button>
            </Dialog.Trigger>
            <Dialog.Portal>
              <Dialog.Overlay className="fixed inset-0 z-40 bg-foreground/40" />
              <Dialog.Content className="fixed inset-y-0 left-0 z-50 flex h-full w-72 flex-col bg-background p-4 shadow-elevated focus:outline-none">
                <div className="mb-6 flex items-center justify-between">
                  <Dialog.Title className="flex items-center gap-2 text-lg font-bold text-foreground">
                    <BrandMark />
                    MindCare
                  </Dialog.Title>
                  <Dialog.Close aria-label="Close" className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground">
                    <X className="h-5 w-5" />
                  </Dialog.Close>
                </div>

                <nav className="flex-1 space-y-1">
                  {navItems.map((item) => {
                    const isActive = pathname === item.href;
                    const Icon = item.icon;
                    return (
                      <Link
                        key={item.id}
                        href={item.href}
                        onClick={() => setOpen(false)}
                        aria-current={isActive ? "page" : undefined}
                        className={clsx(
                          "flex h-11 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors",
                          isActive ? "bg-primary-soft text-primary-soft-text font-semibold" : "text-muted-foreground hover:bg-surface hover:text-foreground"
                        )}
                      >
                        <Icon className="h-4 w-4" />
                        {t(item.labelKey, item.label)}
                      </Link>
                    );
                  })}
                  <Link
                    href={settingsItem.href}
                    onClick={() => setOpen(false)}
                    className="flex h-11 items-center gap-3 rounded-lg px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-surface hover:text-foreground"
                  >
                    <settingsItem.icon className="h-4 w-4" />
                    {t(settingsItem.labelKey, settingsItem.label)}
                  </Link>
                </nav>

                <div className="border-t border-border pt-4">
                  <LanguageSwitcher />
                </div>
              </Dialog.Content>
            </Dialog.Portal>
          </Dialog.Root>

          <BrandLogo className="text-base" />
        </div>

        <NavAvatarMenu settingsItem={settingsItem} />
      </div>
    </header>
  );
}
