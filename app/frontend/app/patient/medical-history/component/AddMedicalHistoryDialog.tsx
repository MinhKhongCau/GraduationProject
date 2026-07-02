"use client";

import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Modal, Button } from "@/components/ui";
import type { CreateMedicalHistoryRequest } from "@/types";

const schema = z.object({
  conditionName: z.string().min(2, "Enter a condition name"),
  description: z.string().optional(),
  diagnosedAt: z.string().min(1, "Enter a date"),
  isChronic: z.boolean(),
});

export interface AddMedicalHistoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateMedicalHistoryRequest) => void;
  isSubmitting: boolean;
}

export function AddMedicalHistoryDialog({ open, onOpenChange, onSubmit, isSubmitting }: AddMedicalHistoryDialogProps) {
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<CreateMedicalHistoryRequest>({
    resolver: zodResolver(schema),
    defaultValues: { isChronic: false },
  });

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Add a medical history entry">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Condition name</label>
          <input
            type="text"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("conditionName")}
          />
          {errors.conditionName && <p className="mt-1 text-xs text-danger">{errors.conditionName.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Description</label>
          <textarea
            rows={3}
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("description")}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Diagnosed on</label>
          <input
            type="date"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("diagnosedAt")}
          />
          {errors.diagnosedAt && <p className="mt-1 text-xs text-danger">{errors.diagnosedAt.message}</p>}
        </div>
        <Controller
          control={control}
          name="isChronic"
          render={({ field }) => (
            <label className="flex items-center gap-2 text-sm font-medium text-foreground">
              <input type="checkbox" checked={field.value} onChange={(event) => field.onChange(event.target.checked)} />
              This is a chronic condition
            </label>
          )}
        />
        <Button type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? "Saving..." : "Add entry"}
        </Button>
      </form>
    </Modal>
  );
}
