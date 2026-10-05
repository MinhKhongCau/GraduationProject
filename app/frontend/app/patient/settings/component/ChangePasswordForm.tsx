"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Card, CardHeader, Button, Input, Label, FieldError } from "@/components/ui";
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
    <Card>
      <CardHeader title="Change Password" />
      <form
        onSubmit={handleSubmit((values) => {
          onSubmit(values);
          reset();
        })}
        className="max-w-md space-y-4 p-5"
      >
        <div>
          <Label htmlFor="old-password">Current Password</Label>
          <Input id="old-password" type="password" invalid={!!errors.oldPassword} {...register("oldPassword")} />
          <FieldError>{errors.oldPassword?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="new-password">New Password</Label>
          <Input id="new-password" type="password" invalid={!!errors.newPassword} {...register("newPassword")} />
          <FieldError>{errors.newPassword?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="confirm-new-password">Confirm New Password</Label>
          <Input id="confirm-new-password" type="password" invalid={!!errors.confirmNewPassword} {...register("confirmNewPassword")} />
          <FieldError>{errors.confirmNewPassword?.message}</FieldError>
        </div>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Updating..." : "Update password"}
        </Button>
      </form>
    </Card>
  );
}
