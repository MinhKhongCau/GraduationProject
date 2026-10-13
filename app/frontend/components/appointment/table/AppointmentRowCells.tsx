"use client";

import type { ComponentProps, ReactNode } from "react";
import clsx from "clsx";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { ChevronDown, FileText } from "lucide-react";
import { Badge, Button } from "@/components/ui";
import {
  APPOINTMENT_STATUS_TEXT as STATUS_TEXT,
  APPOINTMENT_STATUS_TONES as STATUS_TONES,
  formatAppointmentTime,
} from "../status";
import type { Appointment, AppointmentPayment } from "@/types";
import { shortBookingId } from "./downloads";
import type { OptionalColumn } from "./useAppointmentListState";

const PAYMENT_STATUS_TEXT: Record<AppointmentPayment["status"], string> = {
  PENDING: "Chờ thanh toán",
  SUCCESS: "Đã thanh toán",
  FAILED: "Thanh toán lỗi",
  EXPIRED: "Hết hạn thanh toán",
};

const formatVnd = (amount: number) => `${amount.toLocaleString("vi-VN")} đ`;

/** Row checkbox + Booking ID (opens details) with the "has medical record" badge. */
export function AppointmentIdentityCells({
  appointment,
  selected,
  onSelectedChange,
  onOpenDetail,
  hasMedicalRecord,
}: {
  appointment: Appointment;
  selected: boolean;
  onSelectedChange: (selected: boolean) => void;
  onOpenDetail: () => void;
  hasMedicalRecord: boolean;
}) {
  const bookingId = shortBookingId(appointment.appointmentId);
  return (
    <>
      <td className="w-10 px-3 py-3 align-top">
        <input
          type="checkbox"
          checked={selected}
          onChange={(event) => onSelectedChange(event.target.checked)}
          aria-label={`Select booking ${bookingId}`}
          className="mt-0.5 h-4 w-4 accent-primary"
        />
      </td>
      <td className="whitespace-nowrap px-3 py-3 align-top">
        <button
          type="button"
          onClick={onOpenDetail}
          className="font-mono text-sm font-semibold text-primary hover:underline"
        >
          {bookingId}
        </button>
        {hasMedicalRecord && (
          <Badge tone="primary" className="ml-2 align-middle">
            <FileText className="h-3 w-3" aria-hidden="true" />
            Đã có bệnh án
          </Badge>
        )}
      </td>
    </>
  );
}

/** The optional columns (specialization, session, booked at, amount, status), in header order. */
export function AppointmentOptionalCells({
  appointment,
  columns,
}: {
  appointment: Appointment;
  columns: Record<OptionalColumn, boolean>;
}) {
  return (
    <>
      {columns.specialization && (
        <td className="px-3 py-3 align-top text-sm text-muted-foreground">{appointment.specializationName || "—"}</td>
      )}
      {columns.session && (
        <td className="whitespace-nowrap px-3 py-3 align-top text-sm text-foreground">
          {formatAppointmentTime(appointment.startTime, appointment.endTime)}
        </td>
      )}
      {columns.bookedAt && (
        <td className="whitespace-nowrap px-3 py-3 align-top text-sm text-muted-foreground">
          {new Date(appointment.createdAt).toLocaleString("vi-VN")}
        </td>
      )}
      {columns.price && (
        <td className="whitespace-nowrap px-3 py-3 text-right align-top text-sm tabular-nums text-foreground">
          {appointment.payment ? (
            <>
              {formatVnd(appointment.payment.amountVnd)}
              <p
                className={
                  appointment.payment.status === "SUCCESS" ? "text-xs text-success" : "text-xs text-muted-foreground"
                }
              >
                {PAYMENT_STATUS_TEXT[appointment.payment.status] ?? appointment.payment.status}
              </p>
            </>
          ) : appointment.price ? (
            // No payment order yet (or payment-service unreachable): fall back to the slot price.
            <span className="text-muted-foreground">{formatVnd(appointment.price)}</span>
          ) : (
            "—"
          )}
        </td>
      )}
      {columns.status && (
        <td className="px-3 py-3 align-top">
          <Badge tone={STATUS_TONES[appointment.statusLabel] ?? "neutral"}>
            {STATUS_TEXT[appointment.statusLabel] || appointment.statusLabel}
          </Badge>
          {appointment.statusLabel === "CANCELLED" && appointment.cancellationReason && (
            <p className="mt-1 max-w-[220px] text-xs text-muted-foreground">Lý do: {appointment.cancellationReason}</p>
          )}
        </td>
      )}
    </>
  );
}

const MENU_ITEM =
  "flex h-9 cursor-pointer items-center gap-2 rounded-lg px-3 text-sm font-medium text-foreground outline-none data-[highlighted]:bg-surface data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50";

/** "Thao tác ▾" button opening the row's action menu; children are `RowActionItem`s. */
export function RowActionsMenu({ bookingId, children }: { bookingId: string; children: ReactNode }) {
  return (
    <td className="px-3 py-3 text-right align-top">
      {/* Non-modal so the dialogs opened from these items get focus/pointer events back cleanly. */}
      <DropdownMenu.Root modal={false}>
        <DropdownMenu.Trigger asChild>
          <Button
            size="sm"
            variant="outline"
            className="data-[state=open]:bg-surface"
            aria-label={`Thao tác cho lịch hẹn ${shortBookingId(bookingId)}`}
          >
            Thao tác
            <ChevronDown className="h-3.5 w-3.5" aria-hidden="true" />
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content
            align="end"
            sideOffset={6}
            className="z-50 min-w-[210px] rounded-xl border border-border bg-background p-1.5 shadow-elevated"
          >
            {children}
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
    </td>
  );
}

/** A menu item; pass `danger` for destructive actions and `asChild` to render a link. */
export function RowActionItem({
  danger,
  className,
  ...props
}: ComponentProps<typeof DropdownMenu.Item> & { danger?: boolean }) {
  return (
    <DropdownMenu.Item
      className={clsx(MENU_ITEM, danger && "text-danger data-[highlighted]:bg-danger-soft", className)}
      {...props}
    />
  );
}

export function RowActionSeparator() {
  return <DropdownMenu.Separator className="my-1 h-px bg-border" />;
}
