"use client";

import { CheckCircle2, X, XCircle } from "lucide-react";
import { useErrorContext } from "@/context/ErrorContext";

export function ToastContainer() {
  const { toasts, dismissToast } = useErrorContext();

  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex w-full max-w-sm flex-col gap-2">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          role="status"
          className={`flex items-start gap-2 rounded-lg border p-3 shadow-elevated ${
            toast.tone === "error"
              ? "border-danger/20 bg-danger-soft text-danger"
              : "border-success/20 bg-success-soft text-success"
          }`}
        >
          {toast.tone === "error" ? (
            <XCircle className="mt-0.5 h-4 w-4 shrink-0" />
          ) : (
            <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0" />
          )}
          <p className="flex-1 text-sm font-medium">{toast.message}</p>
          <button
            type="button"
            onClick={() => dismissToast(toast.id)}
            className="shrink-0 opacity-60 hover:opacity-100"
            aria-label="Dismiss"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  );
}
