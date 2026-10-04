"use client";

import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Modal, Button, Input, Textarea, Label, FieldError } from "@/components/ui";
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
          <Label htmlFor="history-condition">Condition name</Label>
          <Input id="history-condition" type="text" invalid={!!errors.conditionName} {...register("conditionName")} />
          <FieldError>{errors.conditionName?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="history-description">Description</Label>
          <Textarea id="history-description" rows={3} {...register("description")} />
        </div>
        <div>
          <Label htmlFor="history-diagnosed-at">Diagnosed on</Label>
          <Input id="history-diagnosed-at" type="date" invalid={!!errors.diagnosedAt} {...register("diagnosedAt")} />
          <FieldError>{errors.diagnosedAt?.message}</FieldError>
        </div>
        <Controller
          control={control}
          name="isChronic"
          render={({ field }) => (
            <label className="flex cursor-pointer items-center gap-2 text-sm font-medium text-foreground">
              <input type="checkbox" className="h-4 w-4 rounded accent-primary" checked={field.value} onChange={(event) => field.onChange(event.target.checked)} />
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
