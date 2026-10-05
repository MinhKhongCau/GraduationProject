"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Modal, Button, Input, Label, FieldError } from "@/components/ui";

const topUpSchema = z.object({
  amount: z.coerce.number().positive("Enter an amount greater than 0"),
});

type TopUpFormInput = z.input<typeof topUpSchema>;
type TopUpFormOutput = z.output<typeof topUpSchema>;

export interface TopUpDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (amount: number) => void;
  isSubmitting: boolean;
}

export function TopUpDialog({ open, onOpenChange, onSubmit, isSubmitting }: TopUpDialogProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TopUpFormInput, unknown, TopUpFormOutput>({ resolver: zodResolver(topUpSchema) });

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Top up your wallet" description="Add funds to your MindCare wallet.">
      <form onSubmit={handleSubmit((values) => onSubmit(values.amount))} className="space-y-4">
        <div>
          <Label htmlFor="topup-amount">Amount (đ)</Label>
          <Input id="topup-amount" type="number" step="1000" invalid={!!errors.amount} {...register("amount")} />
          <FieldError>{errors.amount?.message}</FieldError>
        </div>
        <Button type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? "Processing..." : "Confirm top up"}
        </Button>
      </form>
    </Modal>
  );
}
