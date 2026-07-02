"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { KeyRound } from "lucide-react";
import { AuthCard } from "../component/AuthCard";
import { PasswordField } from "../component/PasswordField";
import { Button } from "@/components/ui";
import { useApiMutation } from "@/hooks";
import { authApi } from "@/api";
import { PASSWORD_MIN_LENGTH, ROUTES } from "@/constants";

const resetPasswordSchema = z
  .object({
    newPassword: z.string().min(PASSWORD_MIN_LENGTH, `At least ${PASSWORD_MIN_LENGTH} characters`),
    confirmPassword: z.string(),
  })
  .refine((values) => values.newPassword === values.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;

export default function ResetPasswordPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordFormValues>({ resolver: zodResolver(resetPasswordSchema) });

  // Documented in API-document.md; auth-service has no handler yet.
  const resetPasswordMutation = useApiMutation({
    mutationFn: (payload: ResetPasswordFormValues) =>
      authApi.resetPassword({ token, newPassword: payload.newPassword }),
  });

  function onSubmit(values: ResetPasswordFormValues) {
    resetPasswordMutation.mutate(values, {
      onSuccess: () => router.push(ROUTES.AUTH.LOGIN),
    });
  }

  return (
    <AuthCard title="Set a new password" subtitle="Choose a strong password you haven't used before.">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <PasswordField
          label="New Password"
          placeholder="••••••••"
          error={errors.newPassword?.message}
          {...register("newPassword")}
        />
        <PasswordField
          label="Confirm New Password"
          placeholder="••••••••"
          error={errors.confirmPassword?.message}
          {...register("confirmPassword")}
        />
        <Button type="submit" className="w-full" disabled={!token || resetPasswordMutation.isPending}>
          <KeyRound className="h-4 w-4" />
          {resetPasswordMutation.isPending ? "Resetting..." : "Reset password"}
        </Button>
        {!token && (
          <p className="text-center text-xs text-danger">
            This link is missing a reset token — please use the link from your email.
          </p>
        )}
      </form>
    </AuthCard>
  );
}
