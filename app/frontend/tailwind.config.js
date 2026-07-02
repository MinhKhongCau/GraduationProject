/**
 * Tailwind v4 is CSS-first: the real design tokens live in app/index.css's
 * `@theme inline` block, which is what actually generates `bg-primary`,
 * `text-danger`, etc. This file exists because it was explicitly requested
 * and to hold `safelist` entries for class names built dynamically at
 * runtime (Tailwind's static analyzer can't see those). Do not add colors
 * here expecting them to take effect — edit app/index.css instead.
 */

/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "media",
  safelist: [
    // Status badges built from a variable (see wallet/booking status pills).
    { pattern: /^(bg|text|border)-(primary|danger|success|warning)(-soft)?$/ },
  ],
  theme: {
    extend: {
      colors: {
        primary: "var(--color-primary)",
        "primary-hover": "var(--color-primary-hover)",
        "primary-soft": "var(--color-primary-soft)",
        "primary-soft-text": "var(--color-primary-soft-text)",
        danger: "var(--color-danger)",
        "danger-soft": "var(--color-danger-soft)",
        success: "var(--color-success)",
        "success-soft": "var(--color-success-soft)",
        warning: "var(--color-warning)",
        "warning-soft": "var(--color-warning-soft)",
        background: "var(--color-background)",
        surface: "var(--color-surface)",
        border: "var(--color-border)",
        foreground: "var(--color-foreground)",
        "muted-foreground": "var(--color-muted-foreground)",
      },
      borderRadius: {
        sm: "var(--radius-sm)",
        md: "var(--radius-md)",
        lg: "var(--radius-lg)",
        xl: "var(--radius-xl)",
      },
      boxShadow: {
        card: "var(--shadow-card)",
        elevated: "var(--shadow-elevated)",
      },
    },
  },
  plugins: [],
};
