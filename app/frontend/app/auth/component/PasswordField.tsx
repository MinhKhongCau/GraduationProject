"use client";

import { forwardRef, useState } from "react";
import { Lock, Eye, EyeOff } from "lucide-react";
import { FormField, type FormFieldProps } from "./FormField";

export const PasswordField = forwardRef<HTMLInputElement, Omit<FormFieldProps, "icon" | "type">>(
  function PasswordField(props, ref) {
    const [visible, setVisible] = useState(false);

    return (
      <div className="relative">
        <FormField ref={ref} icon={<Lock className="h-4 w-4" />} type={visible ? "text" : "password"} {...props} />
        <button
          type="button"
          onClick={() => setVisible((current) => !current)}
          className="absolute right-3 top-8 text-muted-foreground hover:text-foreground"
          aria-label={visible ? "Hide password" : "Show password"}
        >
          {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
        </button>
      </div>
    );
  }
);
