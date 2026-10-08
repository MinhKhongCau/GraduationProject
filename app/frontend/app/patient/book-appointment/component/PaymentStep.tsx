"use client";

import { useEffect } from "react";
import Link from "next/link";
import { Loader2, AlertTriangle } from "lucide-react";
import { Button, Card, buttonClasses } from "@/components/ui";
import { formatAppointmentTime } from "@/components/appointment";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";

export interface PaymentStepProps {
  /** The PENDING_PAYMENT appointment just created; its detail is shown while paying. */
  appointmentId?: string;
  onPay: () => void;
  isPending: boolean;
  isError: boolean;
}

/** What the patient is paying for: expert, session time, specialization and fee. */
function BookingSummary({ appointmentId }: { appointmentId: string }) {
  const { data: appointment } = useApiQuery({
    queryKey: QUERY_KEYS.appointmentDetail(appointmentId),
    queryFn: () => bookingApi.getAppointmentDetail(appointmentId),
  });
  if (!appointment) return null;

  return (
    <Card className="mb-8 w-full max-w-md p-4 text-left text-sm">
      <p className="font-semibold text-foreground">{appointment.expert?.fullName ?? "Chuyên gia"}</p>
      <p className="mt-1 text-muted-foreground">{formatAppointmentTime(appointment.startTime, appointment.endTime)}</p>
      {appointment.specializationName && <p className="text-muted-foreground">{appointment.specializationName}</p>}
      {appointment.patient?.fullName && (
        <p className="text-muted-foreground">Người khám: {appointment.patient.fullName}</p>
      )}
      {appointment.price ? (
        <p className="mt-2 font-semibold text-foreground">{appointment.price.toLocaleString("vi-VN")} đ</p>
      ) : null}
    </Card>
  );
}

/** Auto-creates the VNPay order on mount and redirects; offers retry/pay-later on failure. */
export function PaymentStep({ appointmentId, onPay, isPending, isError }: PaymentStepProps) {
  useEffect(() => {
    onPay();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex h-full flex-col items-center justify-center text-center">
      {appointmentId && <BookingSummary appointmentId={appointmentId} />}
      {isError ? (
        <>
          <AlertTriangle className="mb-4 h-12 w-12 text-danger" />
          <h1 className="mb-2 text-2xl font-bold text-foreground">Couldn&apos;t start payment</h1>
          <p className="mb-8 max-w-md text-sm text-muted-foreground">
            Your booking was created and is on hold as pending payment. You can retry now, or complete
            payment later from My Bookings.
          </p>
          <div className="flex gap-3">
            <Link href={ROUTES.PATIENT.MY_BOOKINGS} className={buttonClasses("outline")}>
              Pay later
            </Link>
            <Button onClick={onPay} disabled={isPending}>
              {isPending ? "Retrying..." : "Retry payment"}
            </Button>
          </div>
        </>
      ) : (
        <>
          <Loader2 className="mb-4 h-12 w-12 animate-spin text-primary" />
          <h1 className="mb-2 text-2xl font-bold text-foreground">Redirecting to secure payment…</h1>
          <p className="max-w-md text-sm text-muted-foreground">
            Your booking is on hold. Please wait while we take you to VNPay to complete payment.
          </p>
        </>
      )}
    </div>
  );
}
