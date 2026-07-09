"use client";

import { Users, GraduationCap, CalendarCheck, UserCog } from "lucide-react";
import { StatCard } from "./component/StatCard";
import { Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { expertApi, specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";

/**
 * No admin aggregate endpoints exist yet (no GET /admin/stats-style route
 * in any service). Experts/specializations counts are real; booking-service
 * only exposes a patient's own or an expert's own appointments, not an
 * admin-wide list, so bookings/patients stay labeled placeholders.
 */
export default function AdminDashboardPage() {
  const { data: experts = [], isLoading: isLoadingExperts } = useApiQuery({
    queryKey: QUERY_KEYS.experts(),
    queryFn: () => expertApi.getAllExperts(),
  });

  const { data: specializations = [] } = useApiQuery({
    queryKey: QUERY_KEYS.specializations(),
    queryFn: () => specializationApi.getAllSpecializations(),
  });

  return (
    <div className="mx-auto max-w-6xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">Admin Dashboard</h1>

      {isLoadingExperts ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard label="Total Experts" value={experts.length} icon={GraduationCap} />
          <StatCard label="Specializations" value={specializations.length} icon={UserCog} />
          <StatCard label="Recent Bookings (no data source)" value="—" icon={CalendarCheck} />
          <StatCard label="Total Patients (mock)" value="—" icon={Users} />
        </div>
      )}
    </div>
  );
}
