import { forwardRef, type ButtonHTMLAttributes } from "react";
import clsx from "clsx";

export type ButtonVariant = "primary" | "soft" | "outline" | "ghost" | "danger";
export type ButtonSize = "sm" | "md" | "lg" | "icon";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary: "border border-primary bg-primary text-white hover:border-primary-hover hover:bg-primary-hover shadow-card",
  soft: "border border-transparent bg-primary-soft text-primary-soft-text hover:bg-primary-soft/70",
  outline: "border border-border-strong bg-background text-foreground hover:bg-surface",
  ghost: "border border-transparent text-foreground hover:bg-surface",
  danger: "border border-danger bg-danger text-white hover:opacity-90 shadow-card",
};

const SIZE_CLASSES: Record<ButtonSize, string> = {
  sm: "h-8 px-3 text-xs",
  md: "h-10 px-4 text-sm",
  lg: "h-12 px-6 text-base",
  icon: "h-10 w-10 p-0",
};

/** Shared control-button styles, so links/labels styled as buttons match `<Button>`. */
export function buttonClasses(variant: ButtonVariant = "primary", size: ButtonSize = "md", className?: string) {
  return clsx(
    "inline-flex shrink-0 items-center justify-center gap-2 rounded-lg font-semibold whitespace-nowrap transition-colors duration-150",
    "disabled:cursor-not-allowed disabled:opacity-50",
    VARIANT_CLASSES[variant],
    SIZE_CLASSES[size],
    className
  );
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, variant = "primary", size = "md", disabled, ...props },
  ref
) {
  return (
    <button
      ref={ref}
      disabled={disabled}
      className={buttonClasses(variant, size, className)}
      {...props}
    />
  );
});
