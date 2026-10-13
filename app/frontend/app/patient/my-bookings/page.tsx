"use client";

import { useState } from "react";
import clsx from "clsx";
import { CalendarDays, Download, Ticket, X } from "lucide-react";
import { BookingTableRow, type OptionalColumn } from "./component/BookingTableRow";
import { CancelBookingDialog } from "./component/CancelBookingDialog";
import { ToolbarDropdown, ToolbarDropdownCheckbox, ToolbarDropdownLabel } from "./component/ToolbarDropdown";
import { downloadBookingsCsv, downloadTicketsIcs } from "./component/downloads";
import { Button, Pagination, Spinner } from "@/components/ui";
import { APPOINTMENT_STATUS_TEXT } from "@/components/appointment";
import { useMyBookingsPage, useMyBookingCounts, useCancelBooking, usePaymentRedirect } from "@/hooks";
import { useApiMutation } from "@/hooks";
import { paymentApi } from "@/api";
import { CHANNELING_FEE } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AppointmentWithExpert } from "@/hooks";
import type { AppointmentListParams, AppointmentStatus } from "@/types";

type StatusFilter = "ALL" | AppointmentStatus;
type BulkAction = "" | "cancel" | "download";

const STATUSES = [
  "PENDING_PAYMENT",
  "CONFIRMED",
  "COMPLETED",
  "CANCELLED",
] as const satisfies readonly AppointmentStatus[];
const PAGE_SIZES = [10, 20, 50] as const;
/** booking-service rejects date ranges longer than this. */
const MAX_RANGE_DAYS = 366;

const SORT_OPTIONS = [
  { id: "session-desc", label: "Session: newest first", orderBy: "start_time", sort: "desc" },
  { id: "session-asc", label: "Session: oldest first", orderBy: "start_time", sort: "asc" },
  { id: "booked-desc", label: "Booked: newest first", orderBy: "created_at", sort: "desc" },
] as const satisfies readonly {
  id: string;
  label: string;
  orderBy: AppointmentListParams["orderBy"];
  sort: AppointmentListParams["sort"];
}[];
type SortOption = (typeof SORT_OPTIONS)[number]["id"];

/** Shift a YYYY-MM-DD date by whole days (UTC math, so no DST drift). */
function shiftDate(date: string, days: number): string {
  const value = new Date(`${date}T00:00:00Z`);
  value.setUTCDate(value.getUTCDate() + days);
  return value.toISOString().slice(0, 10);
}

const STATUS_FILTERS: { id: StatusFilter; label: string }[] = [
  { id: "ALL", label: "All" },
  { id: "PENDING_PAYMENT", label: APPOINTMENT_STATUS_TEXT.PENDING_PAYMENT },
  { id: "CONFIRMED", label: APPOINTMENT_STATUS_TEXT.CONFIRMED },
  { id: "COMPLETED", label: APPOINTMENT_STATUS_TEXT.COMPLETED },
  { id: "CANCELLED", label: APPOINTMENT_STATUS_TEXT.CANCELLED },
];

const OPTIONAL_COLUMNS: {
  id: OptionalColumn;
  label: string;
  align?: "right";
}[] = [
  { id: "specialization", label: "Specialization" },
  { id: "session", label: "Session time" },
  { id: "bookedAt", label: "Booked at" },
  { id: "price", label: "Amount", align: "right" },
  { id: "status", label: "Status" },
];

const isCancelable = (a: AppointmentWithExpert) => a.statusLabel === "CONFIRMED" || a.statusLabel === "PENDING_PAYMENT";

export default function MyBookingsPage() {
  const [cancelingAppointments, setCancelingAppointments] = useState<AppointmentWithExpert[]>([]);
  const [payingAppointmentId, setPayingAppointmentId] = useState<string | null>(null);
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
  const [isBulkCanceling, setIsBulkCanceling] = useState(false);
  const [columns, setColumns] = useState<Record<OptionalColumn, boolean>>({
    specialization: true,
    session: true,
    bookedAt: true,
    price: true,
    status: true,
  });
  const { showSuccess, showError } = useErrorContext();

  // Only send the dates the user picked: an explicit `undefined` would drop the API's default window.
  const dateRange: Pick<AppointmentListParams, "from" | "to"> = {};
  if (fromDate) dateRange.from = fromDate;
  if (toDate) dateRange.to = toDate;
  const sort = SORT_OPTIONS.find((option) => option.id === sortOption) ?? SORT_OPTIONS[0];

  const {
    data: bookingPage,
    isLoading,
    isFetching,
    refetch,
  } = useMyBookingsPage({
    ...dateRange,
    status: statusFilter === "ALL" ? undefined : statusFilter,
    orderBy: sort.orderBy,
    sort: sort.sort,
    page,
    size: pageSize,
  });
  const filterCounts = useMyBookingCounts(STATUSES, dateRange);
  const cancelMutation = useCancelBooking();
  const { isPaymentBrowserOpen, openPayment } = usePaymentRedirect({
    onBrowserFinished: () => {
      setPayingAppointmentId(null);
      void refetch();
      showError({
        message:
          "Payment browser closed. Your booking may still be pending payment; please check its status before trying again.",
      });
    },
  });

  const payMutation = useApiMutation({
    mutationFn: (appointment: AppointmentWithExpert) =>
      paymentApi.createOrder({
        expertId: appointment.expertId,
        amount: appointment.price ?? CHANNELING_FEE,
        gateway: "VNPAY",
        appointmentId: appointment.appointmentId,
      }),
    onSuccess: async (order) => {
      const opened = await openPayment(order.paymentUrl);
      if (!opened) {
        setPayingAppointmentId(null);
        showError({
          message: "Unable to open the payment page. Please try again.",
        });
      }
    },
  });

  const visibleAppointments = bookingPage?.items ?? [];
  const totalItems = bookingPage?.totalItems ?? 0;
  const totalPages = bookingPage?.totalPages ?? 0;

  // Cancelling the last rows of the last page (under a status filter) can leave `page` past the end.
  if (bookingPage && !isFetching && page > 0 && page >= totalPages) {
    setPage(Math.max(totalPages - 1, 0));
  }

  const selectedAppointments = visibleAppointments.filter((a) => selectedIds.has(a.appointmentId));
  const allVisibleSelected =
    visibleAppointments.length > 0 && selectedAppointments.length === visibleAppointments.length;
  const firstRow = totalItems === 0 ? 0 : page * pageSize + 1;
  const lastRow = page * pageSize + visibleAppointments.length;
  // checkbox + Booking ID + Expert + Actions, plus the optional columns shown.
  const visibleColumnCount = 4 + OPTIONAL_COLUMNS.filter((c) => columns[c.id]).length;

  function goToPage(next: number) {
    setPage(next);
    setSelectedIds(new Set());
  }

  function handleFilterChange(filter: StatusFilter) {
    setStatusFilter(filter);
    goToPage(0);
  }

  function handleDateChange(which: "from" | "to", value: string) {
    if (which === "from") setFromDate(value);
    else setToDate(value);
    goToPage(0);
  }

  function clearDates() {
    setFromDate("");
    setToDate("");
    goToPage(0);
  }

  function handleSortChange(option: SortOption) {
    setSortOption(option);
    goToPage(0);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    goToPage(0);
  }

  function toggleSelected(appointmentId: string, selected: boolean) {
    setSelectedIds((current) => {
      const next = new Set(current);
      if (selected) next.add(appointmentId);
      else next.delete(appointmentId);
      return next;
    });
  }

  function toggleAllVisible(selected: boolean) {
    setSelectedIds(selected ? new Set(visibleAppointments.map((a) => a.appointmentId)) : new Set());
  }

  function handlePayNow(appointment: AppointmentWithExpert) {
    setPayingAppointmentId(appointment.appointmentId);
    payMutation.mutate(appointment);
  }

  function handleApplyBulkAction() {
    if (selectedAppointments.length === 0) {
      showError({ message: "Select at least one booking first." });
      return;
    }
    if (bulkAction === "download") {
      downloadBookingsCsv(selectedAppointments);
    } else if (bulkAction === "cancel") {
      const cancelable = selectedAppointments.filter(isCancelable);
      if (cancelable.length === 0) {
        showError({
          message: "None of the selected bookings can be canceled.",
        });
        return;
      }
      setCancelingAppointments(cancelable);
    }
  }

  function handleDownloadTickets() {
    const source = selectedAppointments.length > 0 ? selectedAppointments : visibleAppointments;
    const count = downloadTicketsIcs(source);
    if (count === 0) showError({ message: "No confirmed sessions to download tickets for." });
  }

  async function handleConfirmCancel(reason: string) {
    if (cancelingAppointments.length === 0) return;
    setIsBulkCanceling(true);
    const results = await Promise.allSettled(
      cancelingAppointments.map((a) => cancelMutation.mutateAsync({ appointmentId: a.appointmentId, reason }))
    );
    setIsBulkCanceling(false);

    const succeeded = results.filter((r) => r.status === "fulfilled").length;
    if (succeeded > 0) showSuccess(succeeded > 1 ? `${succeeded} bookings canceled.` : "Booking canceled.");
    if (succeeded < results.length)
      showError({
        message: `${results.length - succeeded} booking(s) could not be canceled.`,
      });

    setCancelingAppointments([]);
    setSelectedIds(new Set());
    setBulkAction("");
  }

  return (
    <>
      <div>
        {/* Page header */}
        <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
          <h1 className="flex items-center gap-2 text-2xl font-bold text-foreground">
            <CalendarDays className="h-5 w-5 text-primary" aria-hidden="true" />
            Bookings
          </h1>
          <div className="flex items-center gap-1.5">
            <ToolbarDropdown label="Screen Options">
              <ToolbarDropdownLabel>Columns</ToolbarDropdownLabel>
              {OPTIONAL_COLUMNS.map((column) => (
                <ToolbarDropdownCheckbox
                  key={column.id}
                  checked={columns[column.id]}
                  onCheckedChange={(checked) =>
                    setColumns((current) => ({
                      ...current,
                      [column.id]: checked,
                    }))
                  }
                >
                  {column.label}
                </ToolbarDropdownCheckbox>
              ))}
              <ToolbarDropdownLabel>Bookings per page</ToolbarDropdownLabel>
              {PAGE_SIZES.map((size) => (
                <ToolbarDropdownCheckbox
                  key={size}
                  checked={pageSize === size}
                  onCheckedChange={() => handlePageSizeChange(size)}
                >
                  {size}
                </ToolbarDropdownCheckbox>
              ))}
            </ToolbarDropdown>
            <ToolbarDropdown label="Help">
              <div className="space-y-2 px-3 py-2 text-xs leading-relaxed text-muted-foreground">
                <p>
                  The Actions button on each row opens a menu: booking details, the expert&apos;s profile, joining the
                  consultation room, and paying or cancelling when that applies.
                </p>
                <p>Select rows and use Bulk Actions to cancel or download several bookings at once.</p>
                <p>
                  &quot;Download bookings&quot; and &quot;Download tickets&quot; use the bookings on the current page;
                  tickets are a calendar file (.ics) of confirmed sessions.
                </p>
              </div>
            </ToolbarDropdown>
          </div>
        </div>

        {/* Status filters */}
        <ul className="mb-3 flex flex-wrap items-center gap-y-1 text-sm">
          {STATUS_FILTERS.map((filter, index) => {
            const active = statusFilter === filter.id;
            return (
              <li key={filter.id} className="flex items-center">
                {index > 0 && (
                  <span className="px-2 text-border-strong" aria-hidden="true">
                    |
                  </span>
                )}
                <button
                  type="button"
                  onClick={() => handleFilterChange(filter.id)}
                  aria-pressed={active}
                  className={clsx(active ? "font-semibold text-foreground" : "text-primary hover:underline")}
                >
                  {filter.label}{" "}
                  {filterCounts[filter.id] !== undefined && (
                    <span className="text-muted-foreground">({filterCounts[filter.id]})</span>
                  )}
                </button>
              </li>
            );
          })}
        </ul>

        {/* Date range + sort */}
        <div className="mb-3 flex flex-wrap items-end gap-3 text-xs">
          <label className="flex flex-col gap-1 text-muted-foreground">
            From
            <input
              type="date"
              value={fromDate}
              max={toDate || undefined}
              min={toDate ? shiftDate(toDate, -(MAX_RANGE_DAYS - 1)) : undefined}
              onChange={(event) => handleDateChange("from", event.target.value)}
              className="h-8 rounded-lg border border-border-strong bg-background px-2 text-xs text-foreground"
            />
          </label>
          <label className="flex flex-col gap-1 text-muted-foreground">
            To
            <input
              type="date"
              value={toDate}
              min={fromDate || undefined}
              max={fromDate ? shiftDate(fromDate, MAX_RANGE_DAYS - 1) : undefined}
              onChange={(event) => handleDateChange("to", event.target.value)}
              className="h-8 rounded-lg border border-border-strong bg-background px-2 text-xs text-foreground"
            />
          </label>
          {(fromDate || toDate) && (
            <Button size="sm" variant="ghost" onClick={clearDates}>
              <X className="h-3.5 w-3.5" aria-hidden="true" />
              Clear dates
            </Button>
          )}
          <label className="ml-auto flex flex-col gap-1 text-muted-foreground">
            Sort by
            <select
              value={sortOption}
              onChange={(event) => handleSortChange(event.target.value as SortOption)}
              className="h-8 rounded-lg border border-border-strong bg-background px-2 text-xs text-foreground"
            >
              {SORT_OPTIONS.map((option) => (
                <option key={option.id} value={option.id}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
        </div>

        {/* Action toolbar */}
        <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <select
              value={bulkAction}
              onChange={(event) => setBulkAction(event.target.value as BulkAction)}
              aria-label="Bulk actions"
              className="h-8 rounded-lg border border-border-strong bg-background px-2 text-xs text-foreground"
            >
              <option value="">Bulk Actions</option>
              <option value="cancel">Cancel bookings</option>
              <option value="download">Download selected</option>
            </select>
            <Button size="sm" variant="outline" onClick={handleApplyBulkAction} disabled={!bulkAction}>
              Apply
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => downloadBookingsCsv(visibleAppointments)}
              disabled={visibleAppointments.length === 0}
            >
              <Download className="h-3.5 w-3.5" aria-hidden="true" />
              Download bookings
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={handleDownloadTickets}
              disabled={visibleAppointments.length === 0}
            >
              <Ticket className="h-3.5 w-3.5" aria-hidden="true" />
              Download tickets
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Showing{" "}
            <span className="font-semibold text-foreground">
              {firstRow}–{lastRow}
            </span>{" "}
            of <span className="font-semibold text-foreground">{totalItems}</span>{" "}
            {totalItems === 1 ? "booking" : "bookings"}
            {selectedAppointments.length > 0 && (
              <>
                <span className="px-1.5" aria-hidden="true">
                  |
                </span>
                {selectedAppointments.length} selected
              </>
            )}
          </p>
        </div>

        {/* Data table */}
        <div className="overflow-hidden rounded-xl border border-border bg-background">
          <div
            className={clsx("overflow-x-auto transition-opacity", isFetching && !isLoading && "opacity-60")}
            aria-busy={isFetching}
          >
            <table className="w-full border-collapse text-left">
              <thead className="border-b border-border bg-surface">
                <tr className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  <th scope="col" className="w-10 px-3 py-2.5">
                    <input
                      type="checkbox"
                      checked={allVisibleSelected}
                      onChange={(event) => toggleAllVisible(event.target.checked)}
                      disabled={visibleAppointments.length === 0}
                      aria-label="Select all bookings"
                      className="h-4 w-4 accent-primary"
                    />
                  </th>
                  <th scope="col" className="px-3 py-2.5">
                    Booking ID
                  </th>
                  <th scope="col" className="px-3 py-2.5">
                    Expert
                  </th>
                  {OPTIONAL_COLUMNS.filter((c) => columns[c.id]).map((column) => (
                    <th
                      key={column.id}
                      scope="col"
                      className={clsx("px-3 py-2.5", column.align === "right" && "text-right")}
                    >
                      {column.label}
                    </th>
                  ))}
                  <th scope="col" className="px-3 py-2.5 text-right">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {isLoading ? (
                  <tr>
                    <td colSpan={visibleColumnCount} className="px-3 py-10">
                      <Spinner className="mx-auto h-6 w-6" />
                    </td>
                  </tr>
                ) : visibleAppointments.length === 0 ? (
                  <tr>
                    <td colSpan={visibleColumnCount} className="px-3 py-10 text-center text-sm text-muted-foreground">
                      {statusFilter === "ALL" && !fromDate && !toDate
                        ? "You don't have any bookings yet."
                        : "No bookings match these filters."}
                    </td>
                  </tr>
                ) : (
                  visibleAppointments.map((appointment) => (
                    <BookingTableRow
                      key={appointment.appointmentId}
                      appointment={appointment}
                      columns={columns}
                      selected={selectedIds.has(appointment.appointmentId)}
                      onSelectedChange={(selected) => toggleSelected(appointment.appointmentId, selected)}
                      onCancel={(a) => setCancelingAppointments([a])}
                      onPayNow={handlePayNow}
                      isPaying={
                        (payMutation.isPending || isPaymentBrowserOpen) &&
                        payingAppointmentId === appointment.appointmentId
                      }
                    />
                  ))
                )}
              </tbody>
            </table>
          </div>
          <Pagination page={page + 1} totalPages={totalPages} onPageChange={(next) => goToPage(next - 1)} />
        </div>
      </div>

      <CancelBookingDialog
        appointments={cancelingAppointments}
        onOpenChange={(open) => !open && setCancelingAppointments([])}
        onConfirm={handleConfirmCancel}
        isSubmitting={isBulkCanceling}
      />
    </>
  );
}
