"use client";

import { useEffect } from "react";
import { useQuery, type UseQueryOptions, type UseQueryResult } from "@tanstack/react-query";
import { normalizeError } from "@/api/http/errorNormalizer";
import { useErrorContext } from "@/context/ErrorContext";

export interface UseApiQueryOptions<TData> extends Omit<UseQueryOptions<TData, unknown>, "queryFn"> {
  queryFn: () => Promise<TData>;
  /** Called once per successful fetch — react-query v5 dropped onSuccess. */
  onSuccess?: (data: TData) => void;
}

/**
 * Thin wrapper over react-query's useQuery: callers only ever supply
 * onSuccess — errors are normalized and routed to ErrorContext
 * automatically. See DESIGN.md "React Query + error auto-handling".
 */
export function useApiQuery<TData>({
  onSuccess,
  ...options
}: UseApiQueryOptions<TData>): UseQueryResult<TData, unknown> {
  const { showError } = useErrorContext();
  const query = useQuery(options);

  useEffect(() => {
    if (query.isError) {
      showError(normalizeError(query.error));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query.isError, query.error]);

  useEffect(() => {
    if (query.isSuccess && query.data !== undefined) {
      onSuccess?.(query.data);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query.isSuccess, query.dataUpdatedAt]);

  return query;
}
