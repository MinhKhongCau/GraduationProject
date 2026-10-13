"use client";

import type { ReactNode } from "react";
import clsx from "clsx";
import { Pagination, Spinner } from "@/components/ui";
import { OPTIONAL_COLUMNS, type AppointmentListState } from "./useAppointmentListState";

export interface AppointmentTableProps {
  state: AppointmentListState;
  /** Header of the other-party column ("Expert" / "Patient"). */
  counterpartLabel: string;
  visibleIds: string[];
  totalPages: number;
  isLoading: boolean;
  isFetching: boolean;
  emptyText: string;
  /** One row per appointment. */
  children: ReactNode;
}

/** Bordered table frame: select-all header, loading/empty rows, rows, then pagination. */
export function AppointmentTable({
  state,
  counterpartLabel,
  visibleIds,
  totalPages,
  isLoading,
  isFetching,
  emptyText,
  children,
}: AppointmentTableProps) {
  const shownColumns = OPTIONAL_COLUMNS.filter((column) => state.columns[column.id]);
  // checkbox + Booking ID + counterpart + Actions, plus the optional columns shown.
  const columnCount = 4 + shownColumns.length;
  const allSelected = visibleIds.length > 0 && visibleIds.every((id) => state.selectedIds.has(id));

  return (
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
                  checked={allSelected}
                  onChange={(event) => state.selectAll(visibleIds, event.target.checked)}
                  disabled={visibleIds.length === 0}
                  aria-label="Select all bookings"
                  className="h-4 w-4 accent-primary"
                />
              </th>
              <th scope="col" className="px-3 py-2.5">
                Booking ID
              </th>
              <th scope="col" className="px-3 py-2.5">
                {counterpartLabel}
              </th>
              {shownColumns.map((column) => (
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
                <td colSpan={columnCount} className="px-3 py-10">
                  <Spinner className="mx-auto h-6 w-6" />
                </td>
              </tr>
            ) : visibleIds.length === 0 ? (
              <tr>
                <td colSpan={columnCount} className="px-3 py-10 text-center text-sm text-muted-foreground">
                  {emptyText}
                </td>
              </tr>
            ) : (
              children
            )}
          </tbody>
        </table>
      </div>
      <Pagination page={state.page + 1} totalPages={totalPages} onPageChange={(next) => state.goToPage(next - 1)} />
    </div>
  );
}
