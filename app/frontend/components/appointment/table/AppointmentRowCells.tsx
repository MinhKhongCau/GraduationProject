"use client";

import type { ComponentProps, ReactNode } from "react";
import clsx from "clsx";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { CalendarDays, ChevronDown, Clock, FileText, Video } from "lucide-react";
import { Badge, Button } from "@/components/ui";
import {
  APPOINTMENT_STATUS_TEXT as STATUS_TEXT,
  APPOINTMENT_STATUS_TONES as STATUS_TONES,
  MEETING_STATUS_TEXT,
  MEETING_STATUS_TONES,
  appointmentSlot,
  formatSlotDate,
  formatSlotHours,
} from "../status";
import type { Appointment, AppointmentPayment } from "@/types";
import { shortBookingId } from "./bookingId";
import type { OptionalColumn } from "./useAppointmentListState";

const PAYMENT_STATUS_TEXT: Record<AppointmentPayment["status"], string> = {
  PENDING: "Chờ thanh toán",
  SUCCESS: "Đã thanh toán",
  FAILED: "Thanh toán lỗi",
  EXPIRED: "Hết hạn thanh toán",
};

const formatVnd = (amount: number) => `${amount.toLocaleString("vi-VN")} đ`;

/** Booking ID cell; the ID opens the appointment details. */
export function AppointmentIdentityCell({
  appointment,
  onOpenDetail,
}: {
  appointment: Appointment;
  onOpenDetail: () => void;
}) {
  return (
    <td className="whitespace-nowrap px-3 py-3 align-top">
      <button
        type="button"
        onClick={onOpenDetail}
        className="font-mono text-sm font-semibold text-primary hover:underline"
      >
        {shortBookingId(appointment.appointmentId)}
      </button>
    </td>
  );
}

/** Slot day, hours and length, e.g. "Thứ Ba, 13/10/2026" / "11:00 – 12:00 · 60 phút". */
export function AppointmentSlotInfo({ appointment }: { appointment: Appointment }) {
  const slot = appointmentSlot(appointment);
  if (!slot) return <span className="text-muted-foreground">—</span>;
  return (
    <div className="space-y-0.5">
      <p className="flex items-center gap-1.5 font-medium capitalize text-foreground">
        <CalendarDays className="h-3.5 w-3.5 text-muted-foreground" aria-hidden="true" />
        {formatSlotDate(slot.startTime)}
      </p>
      <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Clock className="h-3.5 w-3.5" aria-hidden="true" />
        <span className="tabular-nums">{formatSlotHours(slot.startTime, slot.endTime)}</span>
        {slot.durationMinutes > 0 && <span>· {slot.durationMinutes} phút</span>}
      </p>
    </div>
  );
}

/** Status badge plus "has medical record" and meeting-room window, and the cancel reason. */
export function AppointmentStatusInfo({ appointment }: { appointment: Appointment }) {
  return (
    <div className="flex flex-col items-start gap-1">
      <Badge tone={STATUS_TONES[appointment.statusLabel] ?? "neutral"}>
        {STATUS_TEXT[appointment.statusLabel] || appointment.statusLabel}
      </Badge>
      {appointment.hasMedicalRecord && (
        <Badge tone="primary">
          <FileText className="h-3 w-3" aria-hidden="true" />
          Đã có bệnh án
        </Badge>
      )}
      {appointment.meetingStatus && (
        <Badge tone={MEETING_STATUS_TONES[appointment.meetingStatus] ?? "neutral"}>
          <Video className="h-3 w-3" aria-hidden="true" />
          {MEETING_STATUS_TEXT[appointment.meetingStatus] ?? appointment.meetingStatus}
        </Badge>
      )}
      {appointment.statusLabel === "CANCELLED" && appointment.cancellationReason && (
        <p className="max-w-[220px] text-xs text-muted-foreground">Lý do: {appointment.cancellationReason}</p>
      )}
    </div>
  );
}

/** Paid amount with its payment status, or the slot price when there is no payment order yet. */
export function AppointmentPriceInfo({ appointment }: { appointment: Appointment }) {
  const slotPrice = appointment.slot?.price || appointment.price;
  if (appointment.payment) {
    return (
      <>
        {formatVnd(appointment.payment.amountVnd)}
        <p className={appointment.payment.status === "SUCCESS" ? "text-xs text-success" : "text-xs text-muted-foreground"}>
          {PAYMENT_STATUS_TEXT[appointment.payment.status] ?? appointment.payment.status}
        </p>
      </>
    );
  }
  // No payment order yet (or payment-service unreachable): fall back to the slot price.
  if (slotPrice) return <span className="text-muted-foreground">{formatVnd(slotPrice)}</span>;
  return <>—</>;
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
        <td className="whitespace-nowrap px-3 py-3 align-top text-sm">
          <AppointmentSlotInfo appointment={appointment} />
        </td>
      )}
      {columns.bookedAt && (
        <td className="whitespace-nowrap px-3 py-3 align-top text-sm text-muted-foreground">
          {new Date(appointment.createdAt).toLocaleString("vi-VN")}
        </td>
      )}
      {columns.price && (
        <td className="whitespace-nowrap px-3 py-3 text-right align-top text-sm tabular-nums text-foreground">
          <AppointmentPriceInfo appointment={appointment} />
        </td>
      )}
      {columns.status && (
        <td className="px-3 py-3 align-top">
          <AppointmentStatusInfo appointment={appointment} />
        </td>
      )}
    </>
  );
}

const MENU_ITEM =
  "flex h-9 cursor-pointer items-center gap-2 rounded-lg px-3 text-sm font-medium text-foreground outline-none data-[highlighted]:bg-surface data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50";

/** "Thao tác ▾" button opening an action menu; children are `RowActionItem`s. Usable outside a table. */
export function RowActionsDropdown({
  bookingId,
  triggerClassName,
  children,
}: {
  bookingId: string;
  triggerClassName?: string;
  children: ReactNode;
}) {
  return (
    <>
      {/* Non-modal so the dialogs opened from these items get focus/pointer events back cleanly. */}
      <DropdownMenu.Root modal={false}>
        <DropdownMenu.Trigger asChild>
          <Button
            size="sm"
            variant="outline"
            className={clsx("data-[state=open]:bg-surface", triggerClassName)}
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
    </>
  );
}

/** The action menu in its table cell; children are `RowActionItem`s. */
export function RowActionsMenu({ bookingId, children }: { bookingId: string; children: ReactNode }) {
  return (
    <td className="px-3 py-3 text-right align-top">
      <RowActionsDropdown bookingId={bookingId}>{children}</RowActionsDropdown>
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

/**
 * "Vào phòng tư vấn" for a CONFIRMED appointment. Too early: link hidden, disabled item with the slot time.
 * Too late: disabled "room closed". No link yet: disabled "not ready".
 */
export function MeetingRoomActionItem({ appointment }: { appointment: Appointment }) {
  if (appointment.statusLabel !== "CONFIRMED") return null;
  if (appointment.meetingStatus === "too_early") {
    const slot = appointmentSlot(appointment);
    return (
      <RowActionItem disabled>
        <Video className="h-4 w-4" aria-hidden="true" />
        <span>
          Phòng tư vấn chưa mở
          {slot && (
            <span className="block text-xs font-normal text-muted-foreground">
              Mở gần giờ hẹn: {formatSlotHours(slot.startTime)}, {new Date(slot.startTime).toLocaleDateString("vi-VN")}
            </span>
          )}
        </span>
      </RowActionItem>
    );
  }
  if (appointment.meetingStatus === "too_late") {
    return (
      <RowActionItem disabled>
        <Video className="h-4 w-4" aria-hidden="true" />
        Phòng tư vấn đã đóng
      </RowActionItem>
    );
  }
  if (!appointment.meetingLink) {
    return (
      <RowActionItem disabled>
        <Video className="h-4 w-4" aria-hidden="true" />
        Phòng tư vấn chưa sẵn sàng
      </RowActionItem>
    );
  }
  return (
    <RowActionItem asChild>
      <a href={appointment.meetingLink} target="_blank" rel="noreferrer">
        <Video className="h-4 w-4" aria-hidden="true" />
        Vào phòng tư vấn
      </a>
    </RowActionItem>
  );
}
