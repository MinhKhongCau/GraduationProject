"use client";

import { CheckCircle2, X, XCircle } from "lucide-react";
import { useErrorContext } from "@/context/ErrorContext";

export function ToastContainer() {
  const { toasts, dismissToast } = useErrorContext();

  if (toasts.length === 0) return null;

  return (
    <div className="fixed inset-x-4 bottom-20 z-[100] flex flex-col gap-2 sm:inset-x-auto sm:bottom-4 sm:right-4 sm:w-full sm:max-w-sm">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          role={toast.tone === "error" ? "alert" : "status"}
          className={`flex items-start gap-3 rounded-xl border border-l-4 bg-background p-3 shadow-elevated ${
            toast.tone === "error" ? "border-border border-l-danger" : "border-border border-l-success"
          }`}
        >
          {toast.tone === "error" ? (
            <XCircle className="mt-0.5 h-5 w-5 shrink-0 text-danger" aria-hidden="true" />
          ) : (
            <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-success" aria-hidden="true" />
          )}
          <p className="flex-1 pt-0.5 text-sm font-medium text-foreground">{toast.message}</p>
          <button
            type="button"
            onClick={() => dismissToast(toast.id)}
            className="-m-1 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground"
            aria-label="Dismiss"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  );
}
