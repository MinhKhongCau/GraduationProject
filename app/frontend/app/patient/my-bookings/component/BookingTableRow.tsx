"use client";

import { useState } from "react";
import Link from "next/link";
import clsx from "clsx";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { ChevronDown, CreditCard, Eye, FileText, UserRound, Video, XCircle } from "lucide-react";
import { Badge, Button } from "@/components/ui";
import { useAppointmentMedicalRecord } from "@/hooks";
import { ROUTES } from "@/constants";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import {
  APPOINTMENT_STATUS_TEXT as STATUS_TEXT,
  APPOINTMENT_STATUS_TONES as STATUS_TONES,
  AppointmentDetailModal,
  formatAppointmentTime,
} from "@/components/appointment";
import type { AppointmentWithExpert } from "@/hooks";
import type { AppointmentPayment } from "@/types";
import { shortBookingId } from "./downloads";

export type OptionalColumn = "specialization" | "session" | "bookedAt" | "price" | "status";

export interface BookingTableRowProps {
  appointment: AppointmentWithExpert;
  columns: Record<OptionalColumn, boolean>;
  selected: boolean;
  onSelectedChange: (selected: boolean) => void;
  onCancel: (appointment: AppointmentWithExpert) => void;
  onPayNow: (appointment: AppointmentWithExpert) => void;
  isPaying: boolean;
}

const PAYMENT_STATUS_TEXT: Record<AppointmentPayment["status"], string> = {
  PENDING: "Chờ thanh toán",
  SUCCESS: "Đã thanh toán",
  FAILED: "Thanh toán lỗi",
  EXPIRED: "Hết hạn thanh toán",
};

const MENU_ITEM =
  "flex h-9 cursor-pointer items-center gap-2 rounded-lg px-3 text-sm font-medium text-foreground outline-none data-[highlighted]:bg-surface data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50";

const formatVnd = (amount: number) => `${amount.toLocaleString("vi-VN")} đ`;

export function BookingTableRow({
  appointment,
  columns,
  selected,
  onSelectedChange,
  onCancel,
  onPayNow,
  isPaying,
}: BookingTableRowProps) {
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);
  const { data: medicalRecord } = useAppointmentMedicalRecord(appointment.appointmentId);

  const canCancel = appointment.statusLabel === "CONFIRMED" || appointment.statusLabel === "PENDING_PAYMENT";
  const canPay = appointment.statusLabel === "PENDING_PAYMENT";
  const isConfirmed = appointment.statusLabel === "CONFIRMED";

  return (
    <tr className={selected ? "bg-primary-soft/40" : "hover:bg-surface/60"}>
      <td className="w-10 px-3 py-3 align-top">
        <input
          type="checkbox"
          checked={selected}
          onChange={(event) => onSelectedChange(event.target.checked)}
          aria-label={`Select booking ${shortBookingId(appointment.appointmentId)}`}
          className="mt-0.5 h-4 w-4 accent-primary"
        />
      </td>

      <td className="whitespace-nowrap px-3 py-3 align-top">
        <button
          type="button"
          onClick={() => setAppointmentDetailOpen(true)}
          className="font-mono text-sm font-semibold text-primary hover:underline"
        >
          {shortBookingId(appointment.appointmentId)}
        </button>
        {medicalRecord && (
          <Badge tone="primary" className="ml-2 align-middle">
            <FileText className="h-3 w-3" aria-hidden="true" />
            Đã có bệnh án
          </Badge>
        )}
      </td>

      <td className="px-3 py-3 align-top text-sm font-medium text-foreground">
        {appointment.expertName ?? "Chuyên gia"}
      </td>

      {columns.specialization && (
        <td className="px-3 py-3 align-top text-sm text-muted-foreground">{appointment.specializationName ?? "—"}</td>
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

      <td className="px-3 py-3 text-right align-top">
        {/* Non-modal so the dialogs opened from these items get focus/pointer events back cleanly. */}
        <DropdownMenu.Root modal={false}>
          <DropdownMenu.Trigger asChild>
            <Button
              size="sm"
              variant="outline"
              className="data-[state=open]:bg-surface"
              aria-label={`Thao tác cho lịch hẹn ${shortBookingId(appointment.appointmentId)}`}
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
              <DropdownMenu.Item className={MENU_ITEM} onSelect={() => setAppointmentDetailOpen(true)}>
                <Eye className="h-4 w-4" aria-hidden="true" />
                Chi tiết
              </DropdownMenu.Item>
              <DropdownMenu.Item asChild className={MENU_ITEM}>
                <Link href={ROUTES.PATIENT.EXPERT_DETAIL(appointment.expertId)}>
                  <UserRound className="h-4 w-4" aria-hidden="true" />
                  Hồ sơ chuyên gia
                </Link>
              </DropdownMenu.Item>
              {isConfirmed &&
                (appointment.meetingLink ? (
                  <DropdownMenu.Item asChild className={MENU_ITEM}>
                    <a href={appointment.meetingLink} target="_blank" rel="noreferrer">
                      <Video className="h-4 w-4" aria-hidden="true" />
                      Vào phòng tư vấn
                    </a>
                  </DropdownMenu.Item>
                ) : (
                  <DropdownMenu.Item disabled className={MENU_ITEM}>
                    <Video className="h-4 w-4" aria-hidden="true" />
                    Phòng tư vấn chưa sẵn sàng
                  </DropdownMenu.Item>
                ))}
              {medicalRecord && (
                <DropdownMenu.Item className={MENU_ITEM} onSelect={() => setDetailModalOpen(true)}>
                  <FileText className="h-4 w-4" aria-hidden="true" />
                  Xem bệnh án
                </DropdownMenu.Item>
              )}
              {canPay && (
                <DropdownMenu.Item className={MENU_ITEM} disabled={isPaying} onSelect={() => onPayNow(appointment)}>
                  <CreditCard className="h-4 w-4" aria-hidden="true" />
                  {isPaying ? "Đang chuyển hướng..." : "Thanh toán ngay"}
                </DropdownMenu.Item>
              )}
              {canCancel && (
                <>
                  <DropdownMenu.Separator className="my-1 h-px bg-border" />
                  <DropdownMenu.Item
                    className={clsx(MENU_ITEM, "text-danger data-[highlighted]:bg-danger-soft")}
                    onSelect={() => onCancel(appointment)}
                  >
                    <XCircle className="h-4 w-4" aria-hidden="true" />
                    Hủy lịch
                  </DropdownMenu.Item>
                </>
              )}
            </DropdownMenu.Content>
          </DropdownMenu.Portal>
        </DropdownMenu.Root>
      </td>

      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />
      <MedicalRecordDetailModal
        open={detailModalOpen}
        onOpenChange={setDetailModalOpen}
        record={medicalRecord || null}
        isExpert={false}
      />
    </tr>
  );
}
