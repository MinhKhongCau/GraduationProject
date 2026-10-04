"use client";

import { useState } from "react";
import { ExpertFilterBar } from "./component/ExpertFilterBar";
import { ExpertCard } from "./component/ExpertCard";
import { PageHeader, Spinner } from "@/components/ui";
import { useApiQuery, useDebounce } from "@/hooks";
import { expertApi, specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";

export default function FindExpertsPage() {
  const [query, setQuery] = useState("");
  const [specializationId, setSpecializationId] = useState("");
  const debouncedQuery = useDebounce(query);

  const { data: specializations = [] } = useApiQuery({
    queryKey: QUERY_KEYS.specializations(),
    queryFn: () => specializationApi.getAllSpecializations(),
  });

  const filters = { query: debouncedQuery, specializationId };
  const { data: experts = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.experts(filters),
    queryFn: () => expertApi.searchExperts(filters),
  });

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader title="Find an expert" />

      <ExpertFilterBar
        query={query}
        onQueryChange={setQuery}
        specializationId={specializationId}
        onSpecializationChange={setSpecializationId}
        specializations={specializations}
      />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : experts.length === 0 ? (
        <p className="text-sm text-muted-foreground">No experts found. Try a different search.</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {experts.map((expert) => (
            <ExpertCard key={expert.expertId} expert={expert} />
          ))}
        </div>
      )}
    </div>
  );
}
