"use client";

import { useEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Card, Button } from "@/components/ui";
import { Camera, User } from "lucide-react";
import Image from "next/image";

const expertProfileSchema = z.object({
  fullName: z.string().min(2, "Enter your full name"),
  avatarUrl: z.string().optional(),
  introductionVideoUrl: z.string().url("Enter a valid URL").or(z.literal("")),
  bio: z.string().optional(),
});

export type ExpertProfileFormValues = z.infer<typeof expertProfileSchema>;

export interface ExpertProfileEditFormProps {
  defaultValues: Partial<ExpertProfileFormValues>;
  onSubmit: (values: ExpertProfileFormValues) => void;
  isSubmitting: boolean;
}

export function ExpertProfileEditForm({ defaultValues, onSubmit, isSubmitting }: ExpertProfileEditFormProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<ExpertProfileFormValues>({ 
    resolver: zodResolver(expertProfileSchema), 
    defaultValues 
  });

  const avatarUrl = watch("avatarUrl");

  useEffect(() => {
    reset(defaultValues);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [defaultValues.fullName, defaultValues.avatarUrl, defaultValues.introductionVideoUrl, defaultValues.bio]);

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
    <Card className="p-6">
      <h3 className="mb-6 text-lg font-bold text-foreground">Profile Settings</h3>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        
        {/* Avatar Upload Section */}
        <div className="flex flex-col items-center gap-4 sm:flex-row">
          <div className="group relative h-24 w-24 overflow-hidden rounded-full border-2 border-border bg-muted transition-all hover:border-primary">
            {avatarUrl ? (
              <Image
                src={avatarUrl}
                alt="Avatar Preview"
                fill
                className="object-cover"
                unoptimized
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                <User size={40} />
              </div>
            )}
            <button
              type="button"
              onClick={triggerFileInput}
              className="absolute inset-0 flex flex-col items-center justify-center bg-black/60 opacity-0 transition-opacity group-hover:opacity-100"
            >
              <Camera className="text-white" size={20} />
              <span className="mt-1 text-[10px] font-medium text-white">Change Photo</span>
            </button>
          </div>
          
          <div className="text-center sm:text-left">
            <h4 className="text-sm font-semibold text-foreground">Profile Picture</h4>
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

        <div className="grid grid-cols-1 gap-4">
          <div>
            <label className="mb-1 block text-xs font-bold text-foreground">Full Name</label>
            <input
              className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
              {...register("fullName")}
            />
            {errors.fullName && <p className="mt-1 text-xs text-danger">{errors.fullName.message}</p>}
          </div>

          <div>
            <label className="mb-1 block text-xs font-bold text-foreground">Introduction URL</label>
            <input
              placeholder="e.g. https://www.youtube.com/watch?v=..."
              className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
              {...register("introductionVideoUrl")}
            />
            {errors.introductionVideoUrl && <p className="mt-1 text-xs text-danger">{errors.introductionVideoUrl.message}</p>}
          </div>

          <div>
            <label className="mb-1 block text-xs font-bold text-foreground">Bio</label>
            <textarea
              rows={4}
              placeholder="Tell patients about your background, experience and philosophy..."
              className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
              {...register("bio")}
            />
            {errors.bio && <p className="mt-1 text-xs text-danger">{errors.bio.message}</p>}
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
