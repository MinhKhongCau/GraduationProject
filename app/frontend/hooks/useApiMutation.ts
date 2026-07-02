"use client";

import { useMutation, type UseMutationOptions, type UseMutationResult } from "@tanstack/react-query";
import { normalizeError } from "@/api/http/errorNormalizer";
import { useErrorContext } from "@/context/ErrorContext";

/**
 * Thin wrapper over react-query's useMutation: errors are normalized and
 * routed to ErrorContext automatically before any caller-supplied onError
 * runs — callers only ever need to supply onSuccess. See DESIGN.md
 * "React Query + error auto-handling".
 */
export function useApiMutation<TData, TVariables = void>(
  options: UseMutationOptions<TData, unknown, TVariables>
): UseMutationResult<TData, unknown, TVariables> {
  const { showError } = useErrorContext();

  return useMutation({
    ...options,
    onError: (error, variables, onMutateResult, context) => {
      showError(normalizeError(error));
      options.onError?.(error, variables, onMutateResult, context);
    },
  });
}
