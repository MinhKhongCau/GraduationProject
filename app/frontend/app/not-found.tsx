import Link from "next/link";
import { Compass } from "lucide-react";
import { buttonClasses } from "@/components/ui";
import { ROUTES } from "@/constants";

export default function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-surface p-6 text-center">
      <Compass className="h-12 w-12 text-primary" />
      <h1 className="text-2xl font-bold text-foreground">Page not found</h1>
      <p className="max-w-md text-sm text-muted-foreground">
        The page you&apos;re looking for doesn&apos;t exist or has moved.
      </p>
      <Link href={ROUTES.HOME} className={buttonClasses("primary", "md")}>
        Back to home
      </Link>
    </div>
  );
}
