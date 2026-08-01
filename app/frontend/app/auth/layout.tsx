import type { ReactNode } from "react";
import Image from "next/image";
import Link from "next/link";
import { ROUTES } from "@/constants";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-surface p-4">
      <div className="flex min-h-[700px] w-full max-w-[1000px] flex-col overflow-hidden rounded-3xl bg-background shadow-elevated md:flex-row">
        <div className="relative hidden md:block md:w-1/2 bg-border">
          <Image src="/images/auth-bg.png" alt="MindCare" fill className="object-cover" priority />
          <Link
            href={ROUTES.HOME}
            className="absolute left-6 top-6 flex items-center gap-2 text-xl font-bold text-white drop-shadow-md"
          >
            <span className="rounded-lg bg-primary p-1.5">🧠</span>
            MindCare
          </Link>
        </div>

        <div className="flex w-full flex-col justify-center bg-background p-8 md:w-1/2 lg:p-12">
          {children}
        </div>
      </div>
    </div>
  );
}
