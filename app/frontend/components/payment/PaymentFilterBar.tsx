"use client";

import type { ReactNode } from "react";
import { Input, Label, Select } from "@/components/ui";
import { useTranslation } from "@/hooks";
import type { OrderStatus, PaymentOrderType } from "@/types";

export interface PaymentFilterValue {
  status?: OrderStatus;
  type?: PaymentOrderType;
  from?: string;
  to?: string;
}

export interface PaymentFilterBarProps<T extends PaymentFilterValue> {
  value: T;
  /** Receives the full next value; callers should reset paging. */
  onChange: (next: T) => void;
  /** Extra controls (e.g. expert picker) rendered before the date range. */
  children?: ReactNode;
}

/** Status / type / date-range filters shared by patient, expert and admin transaction lists. */
export function PaymentFilterBar<T extends PaymentFilterValue>({ value, onChange, children }: PaymentFilterBarProps<T>) {
  const { t } = useTranslation();
  const set = <K extends keyof PaymentFilterValue>(key: K, next: PaymentFilterValue[K]) =>
    onChange({ ...value, [key]: next || undefined });

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5">
      <div>
        <Label htmlFor="payment-filter-status">{t("payment.filter.status", "Status")}</Label>
        <Select
          id="payment-filter-status"
          value={value.status ?? ""}
          onChange={(e) => set("status", e.target.value as OrderStatus)}
        >
          <option value="">{t("payment.filter.all", "All")}</option>
          <option value="SUCCESS">{t("payment.status.SUCCESS", "Paid")}</option>
          <option value="PENDING">{t("payment.status.PENDING", "Pending")}</option>
          <option value="FAILED">{t("payment.status.FAILED", "Failed")}</option>
          <option value="EXPIRED">{t("payment.status.EXPIRED", "Expired")}</option>
        </Select>
      </div>
      <div>
        <Label htmlFor="payment-filter-type">{t("payment.filter.type", "Type")}</Label>
        <Select
          id="payment-filter-type"
          value={value.type ?? ""}
          onChange={(e) => set("type", e.target.value as PaymentOrderType)}
        >
          <option value="">{t("payment.filter.all", "All")}</option>
          <option value="APPOINTMENT">{t("payment.type.APPOINTMENT", "Appointment")}</option>
          <option value="TOP_UP">{t("payment.type.TOP_UP", "Wallet top-up")}</option>
        </Select>
      </div>
      {children}
      <div>
        <Label htmlFor="payment-filter-from">{t("payment.filter.from", "From")}</Label>
        <Input
          id="payment-filter-from"
          type="date"
          value={value.from ?? ""}
          max={value.to}
          onChange={(e) => set("from", e.target.value)}
        />
      </div>
      <div>
        <Label htmlFor="payment-filter-to">{t("payment.filter.to", "To")}</Label>
        <Input
          id="payment-filter-to"
          type="date"
          value={value.to ?? ""}
          min={value.from}
          onChange={(e) => set("to", e.target.value)}
        />
      </div>
    </div>
  );
}
