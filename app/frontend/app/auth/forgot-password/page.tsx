"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Mail, Send } from "lucide-react";
import { AuthCard } from "../component/AuthCard";
import { FormField } from "../component/FormField";
import { Button } from "@/components/ui";
import { useApiMutation } from "@/hooks";
import { authApi } from "@/api";
import { EMAIL_REGEX, ROUTES } from "@/constants";

const forgotPasswordSchema = z.object({
  email: z.string().regex(EMAIL_REGEX, "Enter a valid email address"),
});

type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export default function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormValues>({ resolver: zodResolver(forgotPasswordSchema) });

  // Documented in API-document.md; auth-service has no handler yet, so this
  // is expected to fail — we still show the generic "check your email"
  // message either way to avoid leaking whether an account exists.
  const forgotPasswordMutation = useApiMutation({
    mutationFn: (payload: ForgotPasswordFormValues) => authApi.forgotPassword(payload),
  });

  function onSubmit(values: ForgotPasswordFormValues) {
    forgotPasswordMutation.mutate(values, {
      onSuccess: () => setSent(true),
      onError: () => setSent(true),
    });
  }

  return (
    <AuthCard title="Forgot your password?" subtitle="Enter your email and we'll send you a reset link.">
      {sent ? (
        <p role="status" className="rounded-xl border border-success/20 bg-success-soft p-4 text-sm font-medium text-success">
          If an account exists for that email, a reset link is on its way.
        </p>
      ) : (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <FormField
            label="Email Address"
            icon={<Mail className="h-4 w-4" />}
            type="email"
            placeholder="name@example.com"
            error={errors.email?.message}
            {...register("email")}
          />
          <Button type="submit" size="lg" className="w-full" disabled={forgotPasswordMutation.isPending}>
            <Send className="h-4 w-4" />
            {forgotPasswordMutation.isPending ? "Sending..." : "Send reset link"}
          </Button>
        </form>
      )}

      <p className="text-center text-sm text-muted-foreground">
        <Link href={ROUTES.AUTH.LOGIN} className="font-semibold text-primary hover:underline">
          Back to sign in
        </Link>
      </p>
    </AuthCard>
  );
}
