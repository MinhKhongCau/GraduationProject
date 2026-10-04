import type { ReactNode } from "react";
import Image from "next/image";
import Link from "next/link";
import { ROUTES } from "@/constants";
import { BrandMark } from "@/components/ui";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-dvh items-center justify-center bg-surface p-4 pt-[env(safe-area-inset-top)] pb-[env(safe-area-inset-bottom)]">
      <div className="flex min-h-full w-full max-w-[1000px] flex-col overflow-hidden rounded-2xl border border-border bg-background shadow-elevated md:min-h-[700px] md:flex-row">
        <div className="relative hidden md:block md:w-1/2 bg-border">
          <Image src="/images/auth-bg.png" alt="MindCare" fill className="object-cover" priority />
          <Link
            href={ROUTES.HOME}
            className="absolute left-6 top-6 flex items-center gap-2 text-xl font-bold text-white drop-shadow-md"
          >
            <BrandMark />
            MindCare
          </Link>
        </div>

        <div className="flex w-full flex-col justify-center bg-background p-6 sm:p-8 md:w-1/2 lg:p-12">
          {children}
        </div>
      </div>
    </div>
  );
}
