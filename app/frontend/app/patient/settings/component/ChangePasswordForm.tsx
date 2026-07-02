"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Card, Button } from "@/components/ui";
import { PASSWORD_MIN_LENGTH } from "@/constants";
import type { ChangePasswordRequest } from "@/types";

const schema = z
  .object({
    oldPassword: z.string().min(1, "Enter your current password"),
    newPassword: z.string().min(PASSWORD_MIN_LENGTH, `At least ${PASSWORD_MIN_LENGTH} characters`),
    confirmNewPassword: z.string(),
  })
  .refine((values) => values.newPassword === values.confirmNewPassword, {
    message: "Passwords do not match",
    path: ["confirmNewPassword"],
  });

export interface ChangePasswordFormProps {
  onSubmit: (values: ChangePasswordRequest) => void;
  isSubmitting: boolean;
}

export function ChangePasswordForm({ onSubmit, isSubmitting }: ChangePasswordFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ChangePasswordRequest>({ resolver: zodResolver(schema) });

  return (
    <Card className="p-6">
      <h3 className="mb-4 text-sm font-bold text-foreground">Change Password</h3>
      <form
        onSubmit={handleSubmit((values) => {
          onSubmit(values);
          reset();
        })}
        className="max-w-md space-y-4"
      >
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Current Password</label>
          <input
            type="password"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("oldPassword")}
          />
          {errors.oldPassword && <p className="mt-1 text-xs text-danger">{errors.oldPassword.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">New Password</label>
          <input
            type="password"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("newPassword")}
          />
          {errors.newPassword && <p className="mt-1 text-xs text-danger">{errors.newPassword.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Confirm New Password</label>
          <input
            type="password"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("confirmNewPassword")}
          />
          {errors.confirmNewPassword && <p className="mt-1 text-xs text-danger">{errors.confirmNewPassword.message}</p>}
        </div>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Updating..." : "Update password"}
        </Button>
      </form>
    </Card>
  );
}
