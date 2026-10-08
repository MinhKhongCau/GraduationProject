"use client";

import { useState } from "react";
import Link from "next/link";
import { Eye } from "lucide-react";
import { Badge, Button, Card, Input, Label, PageHeader, Pagination, Select, Spinner } from "@/components/ui";
import {
  APPOINTMENT_STATUS_TEXT,
  APPOINTMENT_STATUS_TONES,
  AppointmentDetailModal,
  formatAppointmentTime,
} from "@/components/appointment";
import { useApiQuery } from "@/hooks";
import { bookingApi, expertApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";
import type { AppointmentListParams, AppointmentStatus } from "@/types";

const PAGE_SIZE = 10;
/** profile-service caps page_size at 100. */
const MANAGED_EXPERTS_PARAMS = { page: 1, pageSize: 100 };

function toDateInput(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

/** Last 30 days through the next 30 days, on the slot start time. */
function defaultRange(): Pick<AppointmentListParams, "from" | "to"> {
  const from = new Date();
  from.setDate(from.getDate() - 30);
  const to = new Date();
  to.setDate(to.getDate() + 30);
  return { from: toDateInput(from), to: toDateInput(to) };
}

/**
 * Admin appointments: only experts this admin approved (and therefore manages). booking-service
 * enforces the scope via profile-service and returns 403 for any other expert.
 */
export default function AdminAppointmentsPage() {
  const [filters, setFilters] = useState<AppointmentListParams>(defaultRange);
  const [page, setPage] = useState(0);
  const [detailId, setDetailId] = useState<string | null>(null);

  const updateFilters = (next: AppointmentListParams) => {
    setFilters(next);
    setPage(0);
  };

  const { data: managed, isLoading: loadingExperts } = useApiQuery({
    queryKey: QUERY_KEYS.managedExperts(MANAGED_EXPERTS_PARAMS),
    queryFn: () => expertApi.listManagedExperts(MANAGED_EXPERTS_PARAMS),
  });
  const experts = managed?.items ?? [];

  const params = { ...filters, page, size: PAGE_SIZE };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.adminAppointments(params),
    queryFn: () => bookingApi.listAdminAppointments(params),
    enabled: experts.length > 0,
  });

  return (
    <div className="mx-auto max-w-7xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý lịch hẹn"
        description="Lịch hẹn của các chuyên gia bạn đã duyệt."
      />

      {loadingExperts ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : experts.length === 0 ? (
        <Card className="p-10 text-center">
          <p className="text-sm text-muted-foreground">
            Bạn chưa quản lý chuyên gia nào. Hãy duyệt hồ sơ chuyên gia để quản lý lịch hẹn của họ.
          </p>
          <Link href={ROUTES.ADMIN.EXPERTS} className="mt-3 inline-block text-sm font-semibold text-primary hover:underline">
            Đi tới Quản lý chuyên gia
          </Link>
        </Card>
      ) : (
        <>
          <Card className="p-4">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <div>
                <Label htmlFor="appointments-status">Trạng thái</Label>
                <Select
                  id="appointments-status"
                  value={filters.status ?? ""}
                  onChange={(e) =>
                    updateFilters({ ...filters, status: (e.target.value || undefined) as AppointmentStatus | undefined })
                  }
                >
                  <option value="">Tất cả</option>
                  {Object.entries(APPOINTMENT_STATUS_TEXT).map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <Label htmlFor="appointments-expert">Chuyên gia</Label>
                <Select
                  id="appointments-expert"
                  value={filters.expertId ?? ""}
                  onChange={(e) => updateFilters({ ...filters, expertId: e.target.value || undefined })}
                >
                  <option value="">Tất cả chuyên gia tôi quản lý</option>
                  {experts.map((expert) => (
                    <option key={expert.accountId} value={expert.accountId}>
                      {expert.fullName || expert.email}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <Label htmlFor="appointments-from">Từ ngày</Label>
                <Input
                  id="appointments-from"
                  type="date"
                  value={filters.from ?? ""}
                  max={filters.to}
                  onChange={(e) => updateFilters({ ...filters, from: e.target.value || undefined })}
                />
              </div>
              <div>
                <Label htmlFor="appointments-to">Đến ngày</Label>
                <Input
                  id="appointments-to"
                  type="date"
                  value={filters.to ?? ""}
                  min={filters.from}
                  onChange={(e) => updateFilters({ ...filters, to: e.target.value || undefined })}
                />
              </div>
            </div>
          </Card>

          {isLoading ? (
            <div className="flex justify-center py-20">
              <Spinner className="h-8 w-8" />
            </div>
          ) : (
            <Card className="overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full border-collapse text-left text-sm">
                  <thead className="border-b border-border bg-surface text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                    <tr>
                      <th className="px-4 py-3">Thời gian khám</th>
                      <th className="px-4 py-3">Chuyên gia</th>
                      <th className="px-4 py-3">Người khám</th>
                      <th className="px-4 py-3">Chuyên khoa</th>
                      <th className="px-4 py-3 text-right">Phí</th>
                      <th className="px-4 py-3">Trạng thái</th>
                      <th className="px-4 py-3 text-right">Chi tiết</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border bg-background">
                    {data?.items.map((appointment) => (
                      <tr key={appointment.appointmentId} className="transition-colors hover:bg-surface/60">
                        <td className="whitespace-nowrap px-4 py-3 text-foreground">
                          {formatAppointmentTime(appointment.startTime, appointment.endTime)}
                        </td>
                        <td className="px-4 py-3 font-medium text-foreground">
                          {appointment.expert?.fullName ?? `#${appointment.expertId.slice(0, 8)}`}
                        </td>
                        <td className="px-4 py-3 text-muted-foreground">
                          <div className="text-foreground">
                            {appointment.patient?.fullName || appointment.patientAccount?.fullName || "—"}
                          </div>
                          {appointment.patientAccount?.fullName &&
                            appointment.patient?.fullName &&
                            appointment.patientAccount.fullName !== appointment.patient.fullName && (
                              <div className="text-xs">Đặt bởi {appointment.patientAccount.fullName}</div>
                            )}
                        </td>
                        <td className="px-4 py-3 text-muted-foreground">{appointment.specializationName || "—"}</td>
                        <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums">
                          {appointment.price ? `${appointment.price.toLocaleString("vi-VN")} đ` : "—"}
                        </td>
                        <td className="px-4 py-3">
                          <Badge tone={APPOINTMENT_STATUS_TONES[appointment.statusLabel] ?? "neutral"}>
                            {APPOINTMENT_STATUS_TEXT[appointment.statusLabel] ?? appointment.statusLabel}
                          </Badge>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setDetailId(appointment.appointmentId)}
                            className="text-primary hover:bg-primary-soft"
                          >
                            <Eye className="h-3.5 w-3.5" /> Xem
                          </Button>
                        </td>
                      </tr>
                    ))}
                    {data?.items.length === 0 && (
                      <tr>
                        <td colSpan={7} className="px-6 py-10 text-center text-muted-foreground">
                          Không có lịch hẹn nào trong khoảng thời gian này.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
              {data && (
                <Pagination page={data.page + 1} totalPages={data.totalPages} onPageChange={(next) => setPage(next - 1)} />
              )}
            </Card>
          )}
        </>
      )}

      <AppointmentDetailModal appointmentId={detailId} onClose={() => setDetailId(null)} />
    </div>
  );
}
