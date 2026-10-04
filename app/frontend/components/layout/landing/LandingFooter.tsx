import Link from "next/link";
import { ROUTES } from "@/constants";
import { BrandMark } from "@/components/ui";

export function LandingFooter() {
  return (
    <footer className="border-t border-border bg-surface px-6 py-10">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 text-sm text-muted-foreground sm:flex-row">
        <div className="flex items-center gap-2 font-bold text-foreground">
          <BrandMark className="h-7 w-7" />
          MindCare
        </div>
        <div className="flex gap-6">
          <Link href={ROUTES.AUTH.LOGIN} className="font-medium transition-colors hover:text-primary">
            Log in
          </Link>
          <Link href={ROUTES.AUTH.REGISTER} className="font-medium transition-colors hover:text-primary">
            Sign up
          </Link>
        </div>
        <p>© 2026 MindCare. All rights reserved.</p>
      </div>
    </footer>
  );
}
