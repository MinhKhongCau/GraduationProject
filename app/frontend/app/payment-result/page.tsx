"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { CheckCircle2, XCircle } from "lucide-react";
import { buttonClasses, Card } from "@/components/ui";
import { ROUTES } from "@/constants";

/**
 * VNPay's browser redirect target (VNP_RETURN_URL) after the patient finishes
 * paying — the actual booking confirmation is done server-to-server via the
 * IPN webhook, this page is purely for immediate UX feedback.
 */
export default function PaymentResultPage() {
  const searchParams = useSearchParams();
  const responseCode = searchParams.get("vnp_ResponseCode");
  const isSuccess = responseCode === "00";
  const amount = searchParams.get("vnp_Amount");
  const displayAmount = amount ? Number(amount) / 100 : null; // VNPay sends amount * 100

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-surface p-6 text-center">
      <Card className="flex w-full max-w-md flex-col items-center gap-4 p-6 sm:p-8">
        {isSuccess ? (
          <>
            <span className="flex h-16 w-16 items-center justify-center rounded-full bg-success-soft"><CheckCircle2 className="h-9 w-9 text-success" aria-hidden="true" /></span>
            <h1 className="text-2xl font-bold text-foreground">Payment successful</h1>
            <p className="text-sm text-muted-foreground">
              {displayAmount
                ? `Your payment of ${displayAmount.toLocaleString("vi-VN")} đ was received. `
                : ""}
              Your booking is being confirmed — this only takes a moment.
            </p>
          </>
        ) : (
          <>
            <span className="flex h-16 w-16 items-center justify-center rounded-full bg-danger-soft"><XCircle className="h-9 w-9 text-danger" aria-hidden="true" /></span>
            <h1 className="text-2xl font-bold text-foreground">Payment not completed</h1>
            <p className="text-sm text-muted-foreground">
              Your booking is still on hold as pending payment. You can retry payment anytime from My
              Bookings.
            </p>
          </>
        )}
        <Link href={ROUTES.PATIENT.MY_BOOKINGS} className={buttonClasses("primary", "md")}>
          Go to My Bookings
        </Link>
      </Card>
    </div>
  );
}
