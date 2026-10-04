"use client";

import { forwardRef, useState } from "react";
import { Lock, Eye, EyeOff } from "lucide-react";
import { FormField, type FormFieldProps } from "./FormField";

export const PasswordField = forwardRef<HTMLInputElement, Omit<FormFieldProps, "icon" | "type">>(
  function PasswordField(props, ref) {
    const [visible, setVisible] = useState(false);

    return (
      <div className="relative">
        <FormField ref={ref} icon={<Lock className="h-4 w-4" />} type={visible ? "text" : "password"} {...props} className={`pr-11 ${props.className ?? ""}`} />
        <button
          type="button"
          onClick={() => setVisible((current) => !current)}
          className="absolute right-1 top-[30px] inline-flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground"
          aria-label={visible ? "Hide password" : "Show password"}
        >
          {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
        </button>
      </div>
    );
  }
);
