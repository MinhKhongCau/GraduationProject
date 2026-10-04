import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";
import { fieldClasses, FieldError } from "@/components/ui";

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
      <label htmlFor={props.id ?? props.name} className="mb-1.5 block text-sm font-medium text-foreground">{label}</label>
      <div className="relative">
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted-foreground">
          {icon}
        </div>
        <input
          ref={ref}
          id={props.id ?? props.name}
          aria-invalid={!!error || undefined}
          className={fieldClasses(!!error, `h-11 pl-10 ${className ?? ""}`)}
          {...props}
        />
      </div>
      <FieldError>{error}</FieldError>
    </div>
  );
});
