"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as Dialog from "@radix-ui/react-dialog";
import { Menu, X } from "lucide-react";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import { ROUTES } from "@/constants";
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
                className="rounded-lg p-2 text-foreground hover:bg-surface"
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
                    <span className="rounded-lg bg-primary p-1 text-white">🧠</span>
                    MindCare
                  </Dialog.Title>
                  <Dialog.Close aria-label="Close" className="rounded-full p-1 text-muted-foreground hover:bg-surface">
                    <X className="h-4 w-4" />
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
                        className={clsx(
                          "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                          isActive ? "bg-primary-soft text-primary-soft-text" : "text-muted-foreground hover:bg-surface"
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
                    className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-surface"
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

          <Link href={ROUTES.HOME} className="flex items-center gap-2 text-base font-bold text-foreground">
            <span className="rounded-lg bg-primary p-1 text-white">🧠</span>
            MindCare
          </Link>
        </div>

        <NavAvatarMenu settingsItem={settingsItem} />
      </div>
    </header>
  );
}
