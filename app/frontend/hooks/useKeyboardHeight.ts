"use client";

import { useEffect, useState } from "react";
import { Capacitor, type PluginListenerHandle } from "@capacitor/core";
import { Keyboard } from "@capacitor/keyboard";

/** Reports the native virtual keyboard height without changing desktop layout. */
export function useKeyboardHeight(): number {
  const [keyboardHeight, setKeyboardHeight] = useState(0);

  useEffect(() => {
    if (!Capacitor.isNativePlatform()) return;

    let disposed = false;
    let listeners: PluginListenerHandle[] = [];

    void Promise.all([
      Keyboard.addListener("keyboardWillShow", ({ keyboardHeight: height }) => {
        if (!disposed) setKeyboardHeight(height);
      }),
      Keyboard.addListener("keyboardWillHide", () => {
        if (!disposed) setKeyboardHeight(0);
      }),
    ]).then((registeredListeners) => {
      if (disposed) {
        void Promise.all(registeredListeners.map((listener) => listener.remove()));
        return;
      }
      listeners = registeredListeners;
    });

    return () => {
      disposed = true;
      void Promise.all(listeners.map((listener) => listener.remove()));
    };
  }, []);

  return keyboardHeight;
}
