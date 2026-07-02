"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Card, Button } from "@/components/ui";
import { PHONE_REGEX } from "@/constants";

const profileSchema = z.object({
  fullName: z.string().min(2, "Enter your full name"),
  dateOfBirth: z.string().min(1, "Enter your date of birth"),
  phoneNumber: z.string().regex(PHONE_REGEX, "Enter a valid phone number").or(z.literal("")),
  gender: z.string().optional(),
  address: z.string().optional(),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;

export interface ProfileEditFormProps {
  defaultValues: Partial<ProfileFormValues>;
  onSubmit: (values: ProfileFormValues) => void;
  isSubmitting: boolean;
}

export function ProfileEditForm({ defaultValues, onSubmit, isSubmitting }: ProfileEditFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ProfileFormValues>({ resolver: zodResolver(profileSchema), defaultValues });

  useEffect(() => {
    reset(defaultValues);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [defaultValues.fullName]);

  return (
    <Card className="p-6">
      <h3 className="mb-4 text-sm font-bold text-foreground">Account Settings</h3>
      <form onSubmit={handleSubmit(onSubmit)} className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Full Name</label>
          <input
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("fullName")}
          />
          {errors.fullName && <p className="mt-1 text-xs text-danger">{errors.fullName.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Date of Birth</label>
          <input
            type="date"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("dateOfBirth")}
          />
          {errors.dateOfBirth && <p className="mt-1 text-xs text-danger">{errors.dateOfBirth.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Phone Number</label>
          <input
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("phoneNumber")}
          />
          {errors.phoneNumber && <p className="mt-1 text-xs text-danger">{errors.phoneNumber.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Gender</label>
          <select
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("gender")}
          >
            <option value="">Prefer not to say</option>
            <option value="FEMALE">Female</option>
            <option value="MALE">Male</option>
            <option value="OTHER">Other</option>
          </select>
        </div>
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-bold text-foreground">Address</label>
          <input
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("address")}
          />
        </div>
        <div className="sm:col-span-2">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Saving..." : "Save changes"}
          </Button>
        </div>
      </form>
    </Card>
  );
}
