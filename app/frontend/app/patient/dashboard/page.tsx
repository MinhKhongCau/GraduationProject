"use client";

import { UpcomingAppointmentCard } from "./component/UpcomingAppointmentCard";
import { WalletSnapshotCard } from "./component/WalletSnapshotCard";
import { QuickLinksGrid } from "./component/QuickLinksGrid";
import { Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi, paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";

export default function PatientDashboardPage() {
  const { user } = useAuthContext();

  const { data: appointments, isLoading: isLoadingAppointments } = useApiQuery({
    queryKey: QUERY_KEYS.bookingHistory(),
    queryFn: () => bookingApi.getBookingHistory(),
  });

  const { data: wallet } = useApiQuery({
    queryKey: user ? QUERY_KEYS.wallet(user.id) : ["wallet", "anonymous"],
    queryFn: () => paymentApi.getWallet(user!.id),
    enabled: !!user,
  });

  const nextAppointment = appointments?.find((appointment) => appointment.status === "CONFIRMED") ?? null;

  return (
    <div className="mx-auto max-w-6xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">
        Welcome back{user ? `, ${user.fullName}` : ""}
      </h1>

      {isLoadingAppointments ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="mb-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
          <UpcomingAppointmentCard appointment={nextAppointment} />
          <WalletSnapshotCard balance={wallet?.balance} />
        </div>
      )}

      <QuickLinksGrid />
    </div>
  );
}
