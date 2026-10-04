"use client";

import { useQuery } from "@tanstack/react-query";
import { CalendarCheck, Users, Wallet } from "lucide-react";
import { StatCard } from "./component/StatCard";
import { useExpertAppointments } from "@/hooks";
import { paymentApi } from "@/api";
import { EXPERT_DASHBOARD_FEATURES, QUERY_KEYS } from "@/constants";
import { FeatureGrid } from "@/components/dashboard";
import { useAuthContext } from "@/context/AuthContext";
import { PageHeader } from "@/components/ui";

export default function ExpertDashboardPage() {
  const { user } = useAuthContext();

  const { data: appointments = [] } = useExpertAppointments();

  const uniquePatients = new Set(appointments.map((appointment) => appointment.patientId)).size;

  const { data: wallet } = useQuery({
    queryKey: user ? QUERY_KEYS.wallet(user.id) : ["wallet", "anonymous"],
    queryFn: () => paymentApi.getWallet(user!.id),
    enabled: !!user,
    retry: 0,
  });

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader title={`Welcome back${user ? `, ${user.fullName}` : ""}`} />

      <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Upcoming appointments" value={appointments.length} icon={CalendarCheck} />
        <StatCard label="Patients seen" value={uniquePatients} icon={Users} />
        <StatCard label="Wallet balance" value={`${(wallet?.balance ?? 0).toLocaleString("vi-VN")} đ`} icon={Wallet} />
      </div>

      <FeatureGrid groups={EXPERT_DASHBOARD_FEATURES} />
    </div>
  );
}
