"use client";

import { Badge, type BadgeTone } from "@/components/ui";
import { useTranslation } from "@/hooks";
import type { CompensationStatus, FulfillmentStatus, OrderStatus, PaymentOrderType } from "@/types";

const ORDER_STATUS: Record<OrderStatus, { tone: BadgeTone; label: string }> = {
  PENDING: { tone: "warning", label: "Pending" },
  SUCCESS: { tone: "success", label: "Paid" },
  FAILED: { tone: "danger", label: "Failed" },
  EXPIRED: { tone: "neutral", label: "Expired" },
};

const FULFILLMENT: Record<FulfillmentStatus, { tone: BadgeTone; label: string }> = {
  PENDING: { tone: "neutral", label: "Processing" },
  BOOKING_CONFIRMED: { tone: "success", label: "Booking confirmed" },
  BOOKING_FAILED: { tone: "danger", label: "Booking failed" },
  MANUAL_REVIEW: { tone: "warning", label: "Manual review" },
  REFUND_REQUIRED: { tone: "danger", label: "Refund required" },
};

const COMPENSATION: Record<CompensationStatus, { tone: BadgeTone; label: string }> = {
  MANUAL_REVIEW: { tone: "warning", label: "Manual review" },
  REFUND_REQUIRED: { tone: "danger", label: "Refund required" },
  RESOLVED: { tone: "success", label: "Resolved" },
};

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  const { t } = useTranslation();
  const style = ORDER_STATUS[status] ?? { tone: "neutral" as BadgeTone, label: status };
  return <Badge tone={style.tone}>{t(`payment.status.${status}`, style.label)}</Badge>;
}

export function FulfillmentBadge({ status }: { status: FulfillmentStatus }) {
  const { t } = useTranslation();
  const style = FULFILLMENT[status] ?? { tone: "neutral" as BadgeTone, label: status };
  return <Badge tone={style.tone}>{t(`payment.fulfillment.${status}`, style.label)}</Badge>;
}

export function CompensationStatusBadge({ status }: { status: CompensationStatus }) {
  const { t } = useTranslation();
  const style = COMPENSATION[status] ?? { tone: "neutral" as BadgeTone, label: status };
  return <Badge tone={style.tone}>{t(`payment.compensation.${status}`, style.label)}</Badge>;
}

export function OrderTypeLabel({ type }: { type: PaymentOrderType }) {
  const { t } = useTranslation();
  return <>{type === "TOP_UP" ? t("payment.type.TOP_UP", "Wallet top-up") : t("payment.type.APPOINTMENT", "Appointment")}</>;
}
