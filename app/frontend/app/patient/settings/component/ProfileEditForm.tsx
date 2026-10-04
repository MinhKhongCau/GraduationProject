"use client";

import { useEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Card, CardHeader, Button, Input, Label, FieldError } from "@/components/ui";
import { Camera, User } from "lucide-react";
import Image from "next/image";

const profileSchema = z.object({
  fullName: z.string().min(2, "Enter your full name"),
  dateOfBirth: z.string().optional().or(z.literal("")),
  avatarUrl: z.string().optional(),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;

export interface ProfileEditFormProps {
  defaultValues: Partial<ProfileFormValues>;
  onSubmit: (values: ProfileFormValues) => void;
  isSubmitting: boolean;
}

export function ProfileEditForm({ defaultValues, onSubmit, isSubmitting }: ProfileEditFormProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<ProfileFormValues>({ 
    resolver: zodResolver(profileSchema), 
    defaultValues 
  });

  const avatarUrl = watch("avatarUrl");

  useEffect(() => {
    reset(defaultValues);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [defaultValues.fullName, defaultValues.avatarUrl, defaultValues.dateOfBirth]);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        setValue("avatarUrl", reader.result as string, { shouldDirty: true });
      };
      reader.readAsDataURL(file);
    }
  };

  const triggerFileInput = () => {
    fileInputRef.current?.click();
  };

  return (
    <Card>
      <CardHeader title="Profile Settings" />
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 p-5">
        
        {/* Avatar Upload Section */}
        <div className="flex flex-col items-center gap-4 sm:flex-row">
          <div className="group relative h-24 w-24 overflow-hidden rounded-full border-2 border-border bg-surface transition-colors hover:border-primary">
            {avatarUrl ? (
              <Image
                src={avatarUrl}
                alt="Avatar Preview"
                fill
                className="object-cover"
                unoptimized // to support base64 strings correctly in Next.js Image
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                <User size={40} />
              </div>
            )}
            <button
              type="button"
              onClick={triggerFileInput}
              aria-label="Change Photo"
              className="absolute inset-0 flex flex-col items-center justify-center bg-foreground/60 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
            >
              <Camera className="text-white" size={20} />
              <span className="mt-1 text-[10px] font-medium text-white">Change Photo</span>
            </button>
          </div>
          
          <div className="text-center sm:text-left">
            <h3 className="text-sm font-semibold text-foreground">Profile Picture</h3>
            <p className="mt-1 text-xs text-muted-foreground">PNG or JPG. Max 2MB.</p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="mt-2"
              onClick={triggerFileInput}
            >
              Upload Image
            </Button>
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept="image/*"
              className="hidden"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <Label htmlFor="profile-full-name">Full Name</Label>
            <Input id="profile-full-name" invalid={!!errors.fullName} {...register("fullName")} />
            <FieldError>{errors.fullName?.message}</FieldError>
          </div>

          <div>
            <Label htmlFor="profile-dob">Date of Birth</Label>
            <Input id="profile-dob" type="date" invalid={!!errors.dateOfBirth} {...register("dateOfBirth")} />
            <FieldError>{errors.dateOfBirth?.message}</FieldError>
          </div>
        </div>

        <div>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Saving..." : "Save changes"}
          </Button>
        </div>
      </form>
    </Card>
  );
}
