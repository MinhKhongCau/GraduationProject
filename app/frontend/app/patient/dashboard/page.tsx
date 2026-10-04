"use client";

import { UpcomingAppointmentCard } from "./component/UpcomingAppointmentCard";
import { WalletSnapshotCard } from "./component/WalletSnapshotCard";
import { PageHeader, Spinner } from "@/components/ui";
import { useApiQuery, useMyBookings } from "@/hooks";
import { paymentApi } from "@/api";
import { PATIENT_DASHBOARD_FEATURES, QUERY_KEYS } from "@/constants";
import { FeatureGrid } from "@/components/dashboard";
import { useAuthContext } from "@/context/AuthContext";

export default function PatientDashboardPage() {
  const { user } = useAuthContext();

  // const { data: appointments, isLoading: isLoadingAppointments } = useMyBookings();

  const { data: wallet } = useApiQuery({
    queryKey: user ? QUERY_KEYS.wallet(user.id) : ["wallet", "anonymous"],
    queryFn: () => paymentApi.getWallet(user!.id),
    enabled: !!user,
  });

  // const nextAppointment = appointments?.find((appointment) => appointment.statusLabel === "CONFIRMED") ?? null;

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader
        title={`Welcome back${user ? `, ${user.fullName}` : ""}`}
        description="Everything beyond the main menu lives here."
      />

      {/* {isLoadingAppointments ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="mb-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
          <UpcomingAppointmentCard appointment={nextAppointment} />
          <WalletSnapshotCard balance={wallet?.balance} />
        </div>
      )} */}

      <FeatureGrid groups={PATIENT_DASHBOARD_FEATURES} />
    </div>
  );
}
