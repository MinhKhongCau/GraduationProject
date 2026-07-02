import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";
import clsx from "clsx";

export interface FormFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  icon: ReactNode;
  error?: string;
}

export const FormField = forwardRef<HTMLInputElement, FormFieldProps>(function FormField(
  { label, icon, error, className, ...props },
  ref
) {
  return (
    <div>
      <label className="mb-1 ml-1 block text-xs font-bold text-foreground">{label}</label>
      <div className="relative">
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted-foreground">
          {icon}
        </div>
        <input
          ref={ref}
          className={clsx(
            "block w-full rounded-xl border border-border py-2 pl-10 pr-3 text-sm outline-none transition-all",
            "focus:border-primary focus:ring-2 focus:ring-primary-soft",
            error && "border-danger focus:border-danger focus:ring-danger-soft",
            className
          )}
          {...props}
        />
      </div>
      {error && <p className="mt-1 ml-1 text-xs text-danger">{error}</p>}
    </div>
  );
});
