"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { CheckCircle2, Clock, XCircle } from "lucide-react";
import { paymentApi } from "@/api";
import { buttonClasses, Card, Spinner } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { PaymentReturnResult } from "@/types";

const POLL_INTERVAL_MS = 2000;
const MAX_POLLS = 10;

type ViewState =
  | { kind: "loading" }
  | { kind: "result"; result: PaymentReturnResult; waitingForBooking: boolean }
  | { kind: "error"; message: string };

/** Booking confirmation is asynchronous (outbox → gRPC/RabbitMQ), so keep polling while it is pending. */
function isBookingPending(result: PaymentReturnResult): boolean {
  return (
    result.paymentStatus === "SUCCESS" &&
    !!result.appointmentId &&
    (result.bookingStatus === "PENDING_PAYMENT" || result.bookingStatus === "UNKNOWN")
  );
}

/**
 * VNPay's browser redirect target (VNP_RETURN_URL). The query string VNPay appended is
 * forwarded to payment-service, which verifies vnp_SecureHash, settles the order (idempotent,
 * same logic as the IPN) and reports the payment status plus the current booking status.
 */
export default function PaymentResultPage() {
  const [state, setState] = useState<ViewState>({ kind: "loading" });

  useEffect(() => {
    // Raw query string: VNPay's signature covers these exact parameters.
    const rawQuery = window.location.search;
    const hasPaymentParams = new URLSearchParams(rawQuery).has("vnp_TxnRef");
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const verify = async (attempt: number) => {
      if (!hasPaymentParams) {
        setState({ kind: "error", message: "No payment information was found in this link." });
        return;
      }
      try {
        const result = await paymentApi.verifyVNPayReturn(rawQuery);
        if (cancelled) return;
        const pending = isBookingPending(result);
        const waitingForBooking = pending && attempt < MAX_POLLS;
        setState({ kind: "result", result, waitingForBooking });
        if (waitingForBooking) {
          timer = setTimeout(() => void verify(attempt + 1), POLL_INTERVAL_MS);
        }
      } catch {
        if (cancelled) return;
        setState({
          kind: "error",
          message: "We could not verify this payment. Please check My Bookings for the latest status.",
        });
      }
    };

    void verify(1);
    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, []);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-surface p-6 text-center">
      <Card className="flex w-full max-w-md flex-col items-center gap-4 p-6 sm:p-8">
        <ResultContent state={state} />
        <Link href={ROUTES.PATIENT.MY_BOOKINGS} className={buttonClasses("primary", "md")}>
          Go to My Bookings
        </Link>
      </Card>
    </div>
  );
}

function ResultContent({ state }: { state: ViewState }) {
  if (state.kind === "loading") {
    return (
      <>
        <Spinner className="h-10 w-10" />
        <h1 className="text-2xl font-bold text-foreground">Verifying your payment…</h1>
      </>
    );
  }

  if (state.kind === "error") {
    return (
      <>
        <StatusIcon tone="danger" />
        <h1 className="text-2xl font-bold text-foreground">Payment could not be verified</h1>
        <p className="text-sm text-muted-foreground">{state.message}</p>
      </>
    );
  }

  const { result, waitingForBooking } = state;
  const amount = `${result.amountVnd.toLocaleString("vi-VN")} đ`;

  if (result.paymentStatus === "SUCCESS") {
    const confirmed = result.bookingStatus === "CONFIRMED" || result.bookingStatus === "COMPLETED";
    const bookingProblem = result.bookingStatus === "CANCELLED";
    return (
      <>
        <StatusIcon tone={bookingProblem ? "warning" : "success"} />
        <h1 className="text-2xl font-bold text-foreground">Payment successful</h1>
        <p className="text-sm text-muted-foreground">
          Your payment of {amount} was received.{" "}
          {!result.appointmentId
            ? ""
            : confirmed
              ? "Your booking is confirmed."
              : bookingProblem
                ? "Your booking could not be confirmed because it was no longer available. Our team will refund this payment."
                : waitingForBooking
                  ? "Confirming your booking…"
                  : "Your booking is still being confirmed — check My Bookings in a moment."}
        </p>
        {waitingForBooking && <Spinner className="h-6 w-6" />}
        <OrderReference result={result} />
      </>
    );
  }

  if (result.paymentStatus === "PENDING") {
    return (
      <>
        <StatusIcon tone="warning" />
        <h1 className="text-2xl font-bold text-foreground">Payment is being processed</h1>
        <p className="text-sm text-muted-foreground">
          We have not received the final result from VNPay yet. Check My Bookings in a few minutes.
        </p>
        <OrderReference result={result} />
      </>
    );
  }

  return (
    <>
      <StatusIcon tone="danger" />
      <h1 className="text-2xl font-bold text-foreground">Payment not completed</h1>
      <p className="text-sm text-muted-foreground">
        {result.paymentStatus === "EXPIRED"
          ? "This payment link expired before the payment was completed."
          : result.responseCode === "24"
            ? "You cancelled the payment."
            : "The payment was not successful."}{" "}
        {result.appointmentId && result.bookingStatus === "CANCELLED"
          ? "The time slot has been released — please book again."
          : "No money was charged for this order."}
      </p>
      <OrderReference result={result} />
    </>
  );
}

function StatusIcon({ tone }: { tone: "success" | "warning" | "danger" }) {
  if (tone === "success") {
    return (
      <span className="flex h-16 w-16 items-center justify-center rounded-full bg-success-soft">
        <CheckCircle2 className="h-9 w-9 text-success" aria-hidden="true" />
      </span>
    );
  }
  if (tone === "warning") {
    return (
      <span className="flex h-16 w-16 items-center justify-center rounded-full bg-warning-soft">
        <Clock className="h-9 w-9 text-warning" aria-hidden="true" />
      </span>
    );
  }
  return (
    <span className="flex h-16 w-16 items-center justify-center rounded-full bg-danger-soft">
      <XCircle className="h-9 w-9 text-danger" aria-hidden="true" />
    </span>
  );
}

function OrderReference({ result }: { result: PaymentReturnResult }) {
  return (
    <p className="text-xs text-muted-foreground">
      Order {result.orderId}
      {result.gatewayTxnRef ? ` · VNPay ref ${result.gatewayTxnRef}` : ""}
    </p>
  );
}
