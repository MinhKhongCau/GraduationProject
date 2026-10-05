import {
  forwardRef,
  type InputHTMLAttributes,
  type LabelHTMLAttributes,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from "react";
import clsx from "clsx";

/**
 * Shared form-control styling. Matches Button's `md` height/radius so inputs
 * and buttons line up in the same row.
 */
export function fieldClasses(invalid?: boolean, className?: string) {
  return clsx(
    "block w-full rounded-lg border bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground/70",
    "transition-colors duration-150 outline-none focus-visible:outline-none",
    "focus:border-primary focus:ring-2 focus:ring-primary/20",
    "disabled:cursor-not-allowed disabled:bg-surface disabled:opacity-70",
    invalid ? "border-danger focus:border-danger focus:ring-danger/20" : "border-border-strong",
    className
  );
}

export function Label({ className, ...props }: LabelHTMLAttributes<HTMLLabelElement>) {
  return <label className={clsx("mb-1.5 block text-sm font-medium text-foreground", className)} {...props} />;
}

export function FieldError({ children }: { children?: string }) {
  if (!children) return null;
  return <p className="mt-1.5 text-xs font-medium text-danger">{children}</p>;
}

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input({ className, invalid, ...props }, ref) {
  return <input ref={ref} aria-invalid={invalid || undefined} className={fieldClasses(invalid, clsx("h-10", className))} {...props} />;
});

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  invalid?: boolean;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(function Textarea(
  { className, invalid, ...props },
  ref
) {
  return <textarea ref={ref} aria-invalid={invalid || undefined} className={fieldClasses(invalid, clsx("py-2", className))} {...props} />;
});

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  invalid?: boolean;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(function Select({ className, invalid, ...props }, ref) {
  return <select ref={ref} aria-invalid={invalid || undefined} className={fieldClasses(invalid, clsx("h-10 pr-8", className))} {...props} />;
});
