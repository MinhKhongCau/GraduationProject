"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Capacitor, type PluginListenerHandle } from "@capacitor/core";
import { Browser } from "@capacitor/browser";

export interface UsePaymentRedirectOptions {
  /** Called after the native payment browser is dismissed, successful or not. */
  onBrowserFinished?: () => void;
}

/**
 * Keeps payment providers out of the native WebView while preserving the
 * browser redirect behavior used by the web application.
 */
export function usePaymentRedirect({ onBrowserFinished }: UsePaymentRedirectOptions = {}) {
  const [isPaymentBrowserOpen, setIsPaymentBrowserOpen] = useState(false);
  const listenerRef = useRef<PluginListenerHandle | null>(null);
  const onBrowserFinishedRef = useRef(onBrowserFinished);

  useEffect(() => {
    onBrowserFinishedRef.current = onBrowserFinished;
  }, [onBrowserFinished]);

  const removeBrowserFinishedListener = useCallback(async () => {
    const listener = listenerRef.current;
    listenerRef.current = null;
    await listener?.remove();
  }, []);

  useEffect(() => {
    return () => {
      void removeBrowserFinishedListener();
    };
  }, [removeBrowserFinishedListener]);

  const openPayment = useCallback(
    async (paymentUrl: string): Promise<boolean> => {
      if (!Capacitor.isNativePlatform()) {
        window.location.href = paymentUrl;
        return true;
      }

      try {
        await removeBrowserFinishedListener();
        listenerRef.current = await Browser.addListener("browserFinished", () => {
          listenerRef.current = null;
          setIsPaymentBrowserOpen(false);
          onBrowserFinishedRef.current?.();
        });

        setIsPaymentBrowserOpen(true);
        await Browser.open({ url: paymentUrl });
        return true;
      } catch {
        setIsPaymentBrowserOpen(false);
        await removeBrowserFinishedListener();
        return false;
      }
    },
    [removeBrowserFinishedListener]
  );

  return { isPaymentBrowserOpen, openPayment };
}
