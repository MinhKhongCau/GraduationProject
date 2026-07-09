"use client";

import { useQuery } from "@tanstack/react-query";
import { CalendarCheck, Users, Wallet } from "lucide-react";
import { StatCard } from "./component/StatCard";
import { useExpertAppointments } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";

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
      <h1 className="mb-6 text-2xl font-bold text-foreground">
        Welcome back{user ? `, ${user.fullName}` : ""}
      </h1>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Upcoming appointments" value={appointments.length} icon={CalendarCheck} />
        <StatCard label="Patients seen" value={uniquePatients} icon={Users} />
        <StatCard label="Wallet balance" value={`${(wallet?.balance ?? 0).toLocaleString("vi-VN")} đ`} icon={Wallet} />
      </div>
    </div>
  );
}
