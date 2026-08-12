"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import type { ApiErrorResponse } from "@/types";

export interface ToastItem {
  id: string;
  message: string;
  tone: "error" | "success";
}

interface ErrorContextValue {
  toasts: ToastItem[];
  showError: (error: ApiErrorResponse | string) => void;
  showSuccess: (message: string) => void;
  dismissToast: (id: string) => void;
}

const ErrorContext = createContext<ErrorContextValue | null>(null);

const TOAST_DURATION_MS = 5000;

export function ErrorProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const dismissToast = useCallback((id: string) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const pushToast = useCallback(
    (message: string, tone: ToastItem["tone"]) => {
      const id = `toast-${Math.random().toString(36).slice(2)}`;
      setToasts((current) => [...current, { id, message, tone }]);
      if (typeof window !== "undefined") {
        window.setTimeout(() => dismissToast(id), TOAST_DURATION_MS);
      }
    },
    [dismissToast]
  );

  const showError = useCallback(
    (error: ApiErrorResponse | string) => {
      const message = typeof error === "string" ? error : error?.message || "Đã xảy ra lỗi";
      pushToast(message, "error");
    },
    [pushToast]
  );

  const showSuccess = useCallback((message: string) => pushToast(message, "success"), [pushToast]);

  const value = useMemo(
    () => ({ toasts, showError, showSuccess, dismissToast }),
    [toasts, showError, showSuccess, dismissToast]
  );

  return <ErrorContext.Provider value={value}>{children}</ErrorContext.Provider>;
}

export function useErrorContext(): ErrorContextValue {
  const context = useContext(ErrorContext);
  if (!context) {
    throw new Error("useErrorContext must be used within an ErrorProvider");
  }
  return context;
}
