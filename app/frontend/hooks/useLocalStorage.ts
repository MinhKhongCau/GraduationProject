"use client";

import { useEffect, useState } from "react";

export function useLocalStorage<T>(key: string, initialValue: T) {
  const [value, setValue] = useState<T>(initialValue);

  useEffect(() => {
    const stored = window.localStorage.getItem(key);
    if (stored !== null) {
      try {
        setValue(JSON.parse(stored) as T);
      } catch {
        setValue(initialValue);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key]);

  function set(next: T) {
    setValue(next);
    window.localStorage.setItem(key, JSON.stringify(next));
  }

  function remove() {
    setValue(initialValue);
    window.localStorage.removeItem(key);
  }

  return [value, set, remove] as const;
}
