"use client";

import { useState } from "react";
import type { Appointment, AppointmentListParams, AppointmentStatus } from "@/types";

export type StatusFilter = "ALL" | AppointmentStatus;
export type OptionalColumn = "specialization" | "session" | "bookedAt" | "price" | "status";
export type BulkAction = "" | "cancel" | "download";

export const STATUSES = [
  "PENDING_PAYMENT",
  "CONFIRMED",
  "COMPLETED",
  "CANCELLED",
] as const satisfies readonly AppointmentStatus[];
export const PAGE_SIZES = [10, 20, 50] as const;
/** booking-service rejects date ranges longer than this. */
export const MAX_RANGE_DAYS = 366;

export const SORT_OPTIONS = [
  { id: "session-desc", label: "Session: newest first", orderBy: "start_time", sort: "desc" },
  { id: "session-asc", label: "Session: oldest first", orderBy: "start_time", sort: "asc" },
  { id: "booked-desc", label: "Booked: newest first", orderBy: "created_at", sort: "desc" },
] as const satisfies readonly {
  id: string;
  label: string;
  orderBy: AppointmentListParams["orderBy"];
  sort: AppointmentListParams["sort"];
}[];
export type SortOption = (typeof SORT_OPTIONS)[number]["id"];

export const OPTIONAL_COLUMNS: { id: OptionalColumn; label: string; align?: "right" }[] = [
  { id: "specialization", label: "Specialization" },
  { id: "session", label: "Session time" },
  { id: "bookedAt", label: "Booked at" },
  { id: "price", label: "Amount", align: "right" },
  { id: "status", label: "Status" },
];

export const isCancelable = (a: Appointment) => a.statusLabel === "CONFIRMED" || a.statusLabel === "PENDING_PAYMENT";

/**
 * Filter, sort, paging, selection and column state shared by the patient Bookings and expert
 * Appointments tables. `params` is what the paged list query takes; `dateRange` feeds the counts.
 */
export function useAppointmentListState() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("ALL");
  /** Zero-based, as booking-service expects. */
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState<number>(PAGE_SIZES[0]);
  /** YYYY-MM-DD on the session date; empty = the API's default window. */
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [sortOption, setSortOption] = useState<SortOption>("session-desc");
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkAction, setBulkAction] = useState<BulkAction>("");
  const [columns, setColumns] = useState<Record<OptionalColumn, boolean>>({
    specialization: true,
    session: true,
    bookedAt: true,
    price: true,
    status: true,
  });

  // Only send the dates the user picked: an explicit `undefined` would drop the API's default window.
  const dateRange: Pick<AppointmentListParams, "from" | "to"> = {};
  if (fromDate) dateRange.from = fromDate;
  if (toDate) dateRange.to = toDate;
  const sort = SORT_OPTIONS.find((option) => option.id === sortOption) ?? SORT_OPTIONS[0];

  const params: AppointmentListParams = {
    ...dateRange,
    status: statusFilter === "ALL" ? undefined : statusFilter,
    orderBy: sort.orderBy,
    sort: sort.sort,
    page,
    size: pageSize,
  };

  function goToPage(next: number) {
    setPage(next);
    setSelectedIds(new Set());
  }

  return {
    params,
    dateRange,
    statusFilter,
    page,
    pageSize,
    fromDate,
    toDate,
    sortOption,
    selectedIds,
    bulkAction,
    columns,
    hasFilters: statusFilter !== "ALL" || !!fromDate || !!toDate,
    setBulkAction,
    setPage,
    goToPage,
    clearSelection: () => setSelectedIds(new Set()),
    setStatusFilter(filter: StatusFilter) {
      setStatusFilter(filter);
      goToPage(0);
    },
    setDate(which: "from" | "to", value: string) {
      if (which === "from") setFromDate(value);
      else setToDate(value);
      goToPage(0);
    },
    clearDates() {
      setFromDate("");
      setToDate("");
      goToPage(0);
    },
    setSortOption(option: SortOption) {
      setSortOption(option);
      goToPage(0);
    },
    setPageSize(size: number) {
      setPageSize(size);
      goToPage(0);
    },
    setColumn(column: OptionalColumn, visible: boolean) {
      setColumns((current) => ({ ...current, [column]: visible }));
    },
    toggleSelected(appointmentId: string, selected: boolean) {
      setSelectedIds((current) => {
        const next = new Set(current);
        if (selected) next.add(appointmentId);
        else next.delete(appointmentId);
        return next;
      });
    },
    selectAll(appointmentIds: string[], selected: boolean) {
      setSelectedIds(selected ? new Set(appointmentIds) : new Set());
    },
  };
}

export type AppointmentListState = ReturnType<typeof useAppointmentListState>;
