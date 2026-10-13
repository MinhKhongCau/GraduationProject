"use client";

import type { ReactNode } from "react";
import clsx from "clsx";
import { CalendarDays, Download, Ticket, X } from "lucide-react";
import { Button } from "@/components/ui";
import { APPOINTMENT_STATUS_TEXT } from "../status";
import { ToolbarDropdown, ToolbarDropdownCheckbox, ToolbarDropdownLabel } from "./ToolbarDropdown";
import {
  MAX_RANGE_DAYS,
  OPTIONAL_COLUMNS,
  PAGE_SIZES,
  SORT_OPTIONS,
  type AppointmentListState,
  type BulkAction,
  type SortOption,
  type StatusFilter,
} from "./useAppointmentListState";

const STATUS_FILTERS: { id: StatusFilter; label: string }[] = [
  { id: "ALL", label: "All" },
  { id: "PENDING_PAYMENT", label: APPOINTMENT_STATUS_TEXT.PENDING_PAYMENT },
  { id: "CONFIRMED", label: APPOINTMENT_STATUS_TEXT.CONFIRMED },
  { id: "COMPLETED", label: APPOINTMENT_STATUS_TEXT.COMPLETED },
  { id: "CANCELLED", label: APPOINTMENT_STATUS_TEXT.CANCELLED },
];

const SELECT_CLASS = "h-8 rounded-lg border border-border-strong bg-background px-2 text-xs text-foreground";

/** Shift a YYYY-MM-DD date by whole days (UTC math, so no DST drift). */
function shiftDate(date: string, days: number): string {
  const value = new Date(`${date}T00:00:00Z`);
  value.setUTCDate(value.getUTCDate() + days);
  return value.toISOString().slice(0, 10);
}

/** Title with "Screen Options" (columns, page size) and "Help" dropdowns. */
export function AppointmentListHeader({
  title,
  state,
  help,
}: {
  title: string;
  state: AppointmentListState;
  help: ReactNode;
}) {
  return (
    <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
      <h1 className="flex items-center gap-2 text-2xl font-bold text-foreground">
        <CalendarDays className="h-5 w-5 text-primary" aria-hidden="true" />
        {title}
      </h1>
      <div className="flex items-center gap-1.5">
        <ToolbarDropdown label="Screen Options">
          <ToolbarDropdownLabel>Columns</ToolbarDropdownLabel>
          {OPTIONAL_COLUMNS.map((column) => (
            <ToolbarDropdownCheckbox
              key={column.id}
              checked={state.columns[column.id]}
              onCheckedChange={(checked) => state.setColumn(column.id, checked)}
            >
              {column.label}
            </ToolbarDropdownCheckbox>
          ))}
          <ToolbarDropdownLabel>Rows per page</ToolbarDropdownLabel>
          {PAGE_SIZES.map((size) => (
            <ToolbarDropdownCheckbox
              key={size}
              checked={state.pageSize === size}
              onCheckedChange={() => state.setPageSize(size)}
            >
              {size}
            </ToolbarDropdownCheckbox>
          ))}
        </ToolbarDropdown>
        <ToolbarDropdown label="Help">
          <div className="space-y-2 px-3 py-2 text-xs leading-relaxed text-muted-foreground">{help}</div>
        </ToolbarDropdown>
      </div>
    </div>
  );
}

/** Inline "All (n) | Pending (n) | …" status links. */
export function StatusFilterLinks({
  state,
  counts,
}: {
  state: AppointmentListState;
  counts: Partial<Record<StatusFilter, number>>;
}) {
  return (
    <ul className="mb-3 flex flex-wrap items-center gap-y-1 text-sm">
      {STATUS_FILTERS.map((filter, index) => {
        const active = state.statusFilter === filter.id;
        return (
          <li key={filter.id} className="flex items-center">
            {index > 0 && (
              <span className="px-2 text-border-strong" aria-hidden="true">
                |
              </span>
            )}
            <button
              type="button"
              onClick={() => state.setStatusFilter(filter.id)}
              aria-pressed={active}
              className={clsx(active ? "font-semibold text-foreground" : "text-primary hover:underline")}
            >
              {filter.label}{" "}
              {counts[filter.id] !== undefined && <span className="text-muted-foreground">({counts[filter.id]})</span>}
            </button>
          </li>
        );
      })}
    </ul>
  );
}

/** Session date range (kept within the server's 366-day limit) and sort order. */
export function DateSortFilters({ state }: { state: AppointmentListState }) {
  const { fromDate, toDate } = state;
  return (
    <div className="mb-3 flex flex-wrap items-end gap-3 text-xs">
      <label className="flex flex-col gap-1 text-muted-foreground">
        From
        <input
          type="date"
          value={fromDate}
          max={toDate || undefined}
          min={toDate ? shiftDate(toDate, -(MAX_RANGE_DAYS - 1)) : undefined}
          onChange={(event) => state.setDate("from", event.target.value)}
          className={SELECT_CLASS}
        />
      </label>
      <label className="flex flex-col gap-1 text-muted-foreground">
        To
        <input
          type="date"
          value={toDate}
          min={fromDate || undefined}
          max={fromDate ? shiftDate(fromDate, MAX_RANGE_DAYS - 1) : undefined}
          onChange={(event) => state.setDate("to", event.target.value)}
          className={SELECT_CLASS}
        />
      </label>
      {(fromDate || toDate) && (
        <Button size="sm" variant="ghost" onClick={state.clearDates}>
          <X className="h-3.5 w-3.5" aria-hidden="true" />
          Clear dates
        </Button>
      )}
      <label className="ml-auto flex flex-col gap-1 text-muted-foreground">
        Sort by
        <select
          value={state.sortOption}
          onChange={(event) => state.setSortOption(event.target.value as SortOption)}
          className={SELECT_CLASS}
        >
          {SORT_OPTIONS.map((option) => (
            <option key={option.id} value={option.id}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
    </div>
  );
}

/** Bulk actions + downloads on the left, "Showing a–b of N" on the right. */
export function AppointmentListToolbar({
  state,
  visibleCount,
  selectedCount,
  totalItems,
  onApplyBulkAction,
  onDownloadBookings,
  onDownloadTickets,
}: {
  state: AppointmentListState;
  visibleCount: number;
  selectedCount: number;
  totalItems: number;
  onApplyBulkAction: () => void;
  onDownloadBookings: () => void;
  onDownloadTickets: () => void;
}) {
  const firstRow = totalItems === 0 ? 0 : state.page * state.pageSize + 1;
  const lastRow = state.page * state.pageSize + visibleCount;
  return (
    <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <select
          value={state.bulkAction}
          onChange={(event) => state.setBulkAction(event.target.value as BulkAction)}
          aria-label="Bulk actions"
          className={SELECT_CLASS}
        >
          <option value="">Bulk Actions</option>
          <option value="cancel">Cancel bookings</option>
          <option value="download">Download selected</option>
        </select>
        <Button size="sm" variant="outline" onClick={onApplyBulkAction} disabled={!state.bulkAction}>
          Apply
        </Button>
        <Button size="sm" variant="outline" onClick={onDownloadBookings} disabled={visibleCount === 0}>
          <Download className="h-3.5 w-3.5" aria-hidden="true" />
          Download bookings
        </Button>
        <Button size="sm" variant="outline" onClick={onDownloadTickets} disabled={visibleCount === 0}>
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
        {selectedCount > 0 && (
          <>
            <span className="px-1.5" aria-hidden="true">
              |
            </span>
            {selectedCount} selected
          </>
        )}
      </p>
    </div>
  );
}
