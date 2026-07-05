"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Mail, User, Calendar, UserPlus } from "lucide-react";
import { AuthCard } from "../component/AuthCard";
import { FormField } from "../component/FormField";
import { PasswordField } from "../component/PasswordField";
import { RoleToggle } from "../component/RoleToggle";
import { SocialAuthButtons } from "../component/SocialAuthButtons";
import { decodeGoogleEmail } from "../decodeGoogleCredential";
import { Button } from "@/components/ui";
import { useAuth, useApiMutation } from "@/hooks";
import { authApi } from "@/api";
import { EMAIL_REGEX, PASSWORD_MIN_LENGTH, ROUTES } from "@/constants";
import { dashboardForRole } from "@/router";
import type { LoginResponse } from "@/types";

const registerSchema = z
  .object({
    fullName: z.string().min(2, "Enter your full name"),
    email: z.string().regex(EMAIL_REGEX, "Enter a valid email address"),
    dateOfBirth: z.string().min(1, "Date of birth is required"),
    password: z.string().min(PASSWORD_MIN_LENGTH, `At least ${PASSWORD_MIN_LENGTH} characters`),
    confirmPassword: z.string(),
    role: z.enum(["PATIENT", "EXPERT"]),
  })
  .refine((values) => values.password === values.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type RegisterFormValues = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const router = useRouter();
  const { registerMutation, applySession } = useAuth();
  const [submitted, setSubmitted] = useState(false);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { role: "PATIENT" },
  });

  const googleLoginMutation = useApiMutation<LoginResponse, string>({
    mutationFn: (idToken) => authApi.loginWithGoogle({ idToken }),
  });

  function onSubmit(values: RegisterFormValues) {
    registerMutation.mutate(values, {
      onSuccess: () => {
        setSubmitted(true);
        router.push(ROUTES.AUTH.LOGIN);
      },
    });
  }

  return (
    <AuthCard title="Create your account" subtitle="Join our community to start your mental health journey.">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <Controller
          control={control}
          name="role"
          render={({ field }) => <RoleToggle value={field.value} onChange={field.onChange} />}
        />

        <FormField
          label="Full Name"
          icon={<User className="h-4 w-4" />}
          placeholder="John Doe"
          error={errors.fullName?.message}
          {...register("fullName")}
        />
        <FormField
          label="Email Address"
          icon={<Mail className="h-4 w-4" />}
          type="email"
          placeholder="name@example.com"
          error={errors.email?.message}
          {...register("email")}
        />
        <FormField
          label="Date of Birth"
          icon={<Calendar className="h-4 w-4" />}
          type="date"
          error={errors.dateOfBirth?.message}
          {...register("dateOfBirth")}
        />
        <PasswordField label="Password" placeholder="••••••••" error={errors.password?.message} {...register("password")} />
        <PasswordField
          label="Confirm Password"
          placeholder="••••••••"
          error={errors.confirmPassword?.message}
          {...register("confirmPassword")}
        />

        <Button type="submit" className="w-full" disabled={registerMutation.isPending}>
          <UserPlus className="h-5 w-5" />
          {registerMutation.isPending ? "Creating account..." : "Create Account"}
        </Button>
        {submitted && (
          <p className="text-center text-xs text-success">Account created — please sign in.</p>
        )}
      </form>

      <SocialAuthButtons
        onGoogleCredential={(idToken) =>
          googleLoginMutation.mutate(idToken, {
            onSuccess: (response) => {
              applySession(response, decodeGoogleEmail(idToken));
              router.push(dashboardForRole(response.data.role));
            },
          })
        }
      />

      <p className="text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link href={ROUTES.AUTH.LOGIN} className="font-semibold text-primary hover:underline">
          Sign in
        </Link>
      </p>
    </AuthCard>
  );
}
