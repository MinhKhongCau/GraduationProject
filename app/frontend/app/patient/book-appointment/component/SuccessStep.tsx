import Link from "next/link";
import { CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui";
import { ROUTES } from "@/constants";

export function SuccessStep() {
  return (
    <div className="flex h-full flex-col items-center justify-center text-center">
      <CheckCircle2 className="mb-4 h-16 w-16 text-success" />
      <h1 className="mb-2 text-2xl font-bold text-foreground">Booking successful!</h1>
      <p className="mb-8 max-w-md text-sm text-muted-foreground">
        Your appointment is pending payment — complete it within 15 minutes to keep your slot. You can
        manage this booking from My Bookings.
      </p>
      <Link href={ROUTES.PATIENT.MY_BOOKINGS}>
        <Button>Go to My Bookings</Button>
      </Link>
    </div>
  );
}
