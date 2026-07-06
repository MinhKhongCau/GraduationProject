"use client";

import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Mail, LogIn } from "lucide-react";
import { AuthCard } from "../component/AuthCard";
import { FormField } from "../component/FormField";
import { PasswordField } from "../component/PasswordField";
import { SocialAuthButtons } from "../component/SocialAuthButtons";
import { Button } from "@/components/ui";
import { useAuth, useApiMutation } from "@/hooks";
import { authApi } from "@/api";
import { EMAIL_REGEX, ROUTES } from "@/constants";
import { dashboardForRole } from "@/router";
import { decodeGoogleEmail } from "../decodeGoogleCredential";
import type { LoginResponse } from "@/types";

const loginSchema = z.object({
  email: z.string().regex(EMAIL_REGEX, "Enter a valid email address"),
  password: z.string().min(1, "Password is required"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, loginMutation, applySession } = useAuth();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({ resolver: zodResolver(loginSchema) });

  const googleLoginMutation = useApiMutation<LoginResponse, string>({
    mutationFn: (idToken) => authApi.loginWithGoogle({ idToken }),
  });

  function redirectAfterLogin(role: LoginResponse["data"]["role"]) {
    const redirect = searchParams.get("redirect");
    
    if (redirect) {
      const isAllowed = 
        (role === "ADMIN" && redirect.startsWith("/admin")) ||
        (role === "EXPERT" && redirect.startsWith("/expert")) ||
        (role === "PATIENT" && redirect.startsWith("/patient"));
      
      if (isAllowed) {
        router.push(redirect);
        return;
      }
    }

    router.push(dashboardForRole(role));
  }

  function onSubmit(values: LoginFormValues) {
    login(values, (response) => redirectAfterLogin(response.data.role));
  }

  return (
    <AuthCard title="Welcome back to MindCare" subtitle="Connect with experts and manage your well-being.">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          label="Email Address"
          icon={<Mail className="h-4 w-4" />}
          type="email"
          placeholder="name@example.com"
          error={errors.email?.message}
          {...register("email")}
        />
        <PasswordField label="Password" placeholder="••••••••" error={errors.password?.message} {...register("password")} />

        <div className="text-right">
          <Link href={ROUTES.AUTH.FORGOT_PASSWORD} className="text-xs font-semibold text-primary hover:underline">
            Forgot password?
          </Link>
        </div>

        <Button type="submit" className="w-full" disabled={loginMutation.isPending}>
          <LogIn className="h-5 w-5" />
          {loginMutation.isPending ? "Signing in..." : "Sign In"}
        </Button>
      </form>

      <SocialAuthButtons
        onGoogleCredential={(idToken) =>
          googleLoginMutation.mutate(idToken, {
            onSuccess: (response) => {
              applySession(response, decodeGoogleEmail(idToken));
              redirectAfterLogin(response.data.role);
            },
          })
        }
      />

      <p className="text-center text-sm text-muted-foreground">
        Don&apos;t have an account?{" "}
        <Link href={ROUTES.AUTH.REGISTER} className="font-semibold text-primary hover:underline">
          Sign up
        </Link>
      </p>
    </AuthCard>
  );
}
