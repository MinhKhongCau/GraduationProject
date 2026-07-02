"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Modal, Button } from "@/components/ui";
import type { CreateWithdrawalRequest } from "@/types";

const withdrawSchema = z.object({
  amount: z.coerce.number().positive("Enter an amount greater than 0"),
  bankInfo: z.string().min(3, "Enter your bank account details"),
});

type WithdrawFormInput = z.input<typeof withdrawSchema>;

export interface WithdrawDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateWithdrawalRequest) => void;
  isSubmitting: boolean;
}

export function WithdrawDialog({ open, onOpenChange, onSubmit, isSubmitting }: WithdrawDialogProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<WithdrawFormInput, unknown, CreateWithdrawalRequest>({ resolver: zodResolver(withdrawSchema) });

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Withdraw funds"
      description="Withdrawal requests are processed within 24 working hours."
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Amount (đ)</label>
          <input
            type="number"
            step="1000"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("amount")}
          />
          {errors.amount && <p className="mt-1 text-xs text-danger">{errors.amount.message}</p>}
        </div>
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Bank account details</label>
          <input
            type="text"
            placeholder="Bank name, account number"
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
            {...register("bankInfo")}
          />
          {errors.bankInfo && <p className="mt-1 text-xs text-danger">{errors.bankInfo.message}</p>}
        </div>
        <Button type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? "Submitting..." : "Request withdrawal"}
        </Button>
      </form>
    </Modal>
  );
}
