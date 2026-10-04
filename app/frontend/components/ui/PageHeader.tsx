import type { ReactNode } from "react";
import clsx from "clsx";

export interface PageHeaderProps {
  title: ReactNode;
  description?: ReactNode;
  /** Right-aligned actions (primary CTA last). Wraps below the title on small screens. */
  actions?: ReactNode;
  className?: string;
}

/** Consistent page title block used at the top of every portal page. */
export function PageHeader({ title, description, actions, className }: PageHeaderProps) {
  return (
    <div className={clsx("mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between", className)}>
      <div className="min-w-0">
        <h1 className="text-2xl font-bold tracking-tight text-foreground">{title}</h1>
        {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}
