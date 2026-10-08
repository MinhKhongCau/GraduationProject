"use client";

import type { ReactNode } from "react";
import { Video } from "lucide-react";
import { Badge, Modal, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { APPOINTMENT_STATUS_TEXT, APPOINTMENT_STATUS_TONES, formatAppointmentTime } from "./status";
import { BookingAccountPanel, ExpertPanel } from "./ParticipantPanels";

export interface AppointmentDetailModalProps {
  /** null closes the modal. */
  appointmentId: string | null;
  onClose: () => void;
  /** Extra actions for the viewer's role (pay, cancel, medical record...). */
  actions?: ReactNode;
}

const RELATIONSHIP_TEXT: Record<string, string> = {
  SELF: "Bản thân",
  PARENT: "Cha/Mẹ",
  CHILD: "Con",
  SPOUSE: "Vợ/Chồng",
  SIBLING: "Anh/Chị/Em",
  RELATIVE: "Người thân",
  OTHER: "Khác",
};

const CANCELLED_BY_TEXT: Record<string, string> = {
  SYSTEM: "Hệ thống",
  PATIENT: "Bệnh nhân",
  EXPERT: "Chuyên gia",
};

/** `list` sections hold label/value rows (a <dl>); others hold arbitrary content such as panels. */
function Section({ title, children, list = true }: { title: string; children: ReactNode; list?: boolean }) {
  const className = "divide-y divide-border rounded-xl border border-border";
  return (
    <section>
      <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{title}</h3>
      {list ? <dl className={className}>{children}</dl> : <div className={className}>{children}</div>}
    </section>
  );
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="grid grid-cols-3 gap-3 px-4 py-2.5 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="col-span-2 break-words text-foreground">{children || "—"}</dd>
    </div>
  );
}

/**
 * Appointment detail shared by patient, expert and admin screens. booking-service attaches the
 * expert and booking-account profiles (profile-service gRPC) and enforces who may read it.
 */
export function AppointmentDetailModal({ appointmentId, onClose, actions }: AppointmentDetailModalProps) {
  const { data: appointment, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.appointmentDetail(appointmentId ?? ""),
    queryFn: () => bookingApi.getAppointmentDetail(appointmentId!),
    enabled: !!appointmentId,
  });

  return (
    <Modal
      open={appointmentId !== null}
      onOpenChange={(open) => !open && onClose()}
      title="Chi tiết lịch hẹn"
      description={appointmentId ? `Mã lịch hẹn: ${appointmentId}` : undefined}
      size="2xl"
    >
      {isLoading || !appointment ? (
        <div className="flex justify-center py-10">
          <Spinner className="h-6 w-6" />
        </div>
      ) : (
        <div className="space-y-5">
          <Section title="Lịch hẹn">
            <Row label="Trạng thái">
              <Badge tone={APPOINTMENT_STATUS_TONES[appointment.statusLabel] ?? "neutral"}>
                {APPOINTMENT_STATUS_TEXT[appointment.statusLabel] ?? appointment.statusLabel}
              </Badge>
            </Row>
            <Row label="Thời gian khám">{formatAppointmentTime(appointment.startTime, appointment.endTime)}</Row>
            <Row label="Chuyên khoa">{appointment.specializationName}</Row>
            <Row label="Phí tư vấn">
              {appointment.price ? `${appointment.price.toLocaleString("vi-VN")} đ` : "—"}
            </Row>
            {appointment.statusLabel === "CONFIRMED" && appointment.meetingLink && (
              <Row label="Phòng tư vấn">
                <a
                  href={appointment.meetingLink}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1.5 font-semibold text-primary hover:underline"
                >
                  <Video className="h-3.5 w-3.5" /> Vào phòng tư vấn
                </a>
              </Row>
            )}
            <Row label="Đặt lúc">{new Date(appointment.createdAt).toLocaleString("vi-VN")}</Row>
            {appointment.confirmedAt && (
              <Row label="Xác nhận lúc">{new Date(appointment.confirmedAt).toLocaleString("vi-VN")}</Row>
            )}
            {appointment.statusLabel === "CANCELLED" && (
              <Row label="Hủy bởi">
                {appointment.cancelledBy ? CANCELLED_BY_TEXT[appointment.cancelledBy] ?? appointment.cancelledBy : "—"}
                {appointment.cancellationReason && (
                  <span className="block text-xs text-muted-foreground">Lý do: {appointment.cancellationReason}</span>
                )}
              </Row>
            )}
          </Section>

          <Section title="Chuyên gia" list={false}>
            <ExpertPanel accountId={appointment.expertId} party={appointment.expert} />
          </Section>

          <Section title="Người khám">
            <Row label="Họ và tên">{appointment.patient?.fullName}</Row>
            <Row label="Ngày sinh">{appointment.patient?.dateOfBirth}</Row>
            <Row label="Giới tính">{appointment.patient?.gender}</Row>
            <Row label="Liên hệ">
              {[appointment.patient?.phoneNumber, appointment.patient?.email].filter(Boolean).join(" · ")}
            </Row>
            <Row label="Quan hệ với người đặt">
              {appointment.patient?.relationship
                ? RELATIONSHIP_TEXT[appointment.patient.relationship] ?? appointment.patient.relationship
                : ""}
            </Row>
          </Section>

          <Section title="Tài khoản đặt lịch" list={false}>
            <BookingAccountPanel accountId={appointment.patientId} party={appointment.patientAccount} />
          </Section>

          {actions && <div className="flex flex-wrap justify-end gap-2">{actions}</div>}
        </div>
      )}
    </Modal>
  );
}
