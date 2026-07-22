"use client";

import { useEffect } from "react";
import Link from "next/link";
import { Loader2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui";
import { ROUTES } from "@/constants";

export interface PaymentStepProps {
  onPay: () => void;
  isPending: boolean;
  isError: boolean;
}

/** Auto-creates the VNPay order on mount and redirects; offers retry/pay-later on failure. */
export function PaymentStep({ onPay, isPending, isError }: PaymentStepProps) {
  useEffect(() => {
    onPay();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex h-full flex-col items-center justify-center text-center">
      {isError ? (
        <>
          <AlertTriangle className="mb-4 h-12 w-12 text-danger" />
          <h1 className="mb-2 text-2xl font-bold text-foreground">Couldn&apos;t start payment</h1>
          <p className="mb-8 max-w-md text-sm text-muted-foreground">
            Your booking was created and is on hold as pending payment. You can retry now, or complete
            payment later from My Bookings.
          </p>
          <div className="flex gap-3">
            <Link href={ROUTES.PATIENT.MY_BOOKINGS}>
              <Button variant="outline">Pay later</Button>
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
