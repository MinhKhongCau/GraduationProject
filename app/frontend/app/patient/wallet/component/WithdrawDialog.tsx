"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Modal, Button, Input, Label, FieldError } from "@/components/ui";
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
          <Label htmlFor="withdraw-amount">Amount (đ)</Label>
          <Input id="withdraw-amount" type="number" step="1000" invalid={!!errors.amount} {...register("amount")} />
          <FieldError>{errors.amount?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="withdraw-bank">Bank account details</Label>
          <Input
            id="withdraw-bank"
            type="text"
            placeholder="Bank name, account number"
            invalid={!!errors.bankInfo}
            {...register("bankInfo")}
          />
          <FieldError>{errors.bankInfo?.message}</FieldError>
        </div>
        <Button type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? "Submitting..." : "Request withdrawal"}
        </Button>
      </form>
    </Modal>
  );
}
