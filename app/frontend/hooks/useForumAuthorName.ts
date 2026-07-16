"use client";

import { useApiQuery } from "./useApiQuery";
import { expertApi } from "@/api";

/**
 * forum-service has no author join — resolves a display name client-side.
 * GAP: only resolves EXPERT authors (profile-service's expert lookup is
 * public); patient authors render as "Community member" since there's no
 * public endpoint to look up a patient's profile by id.
 */
export function useForumAuthorName(authorId: string) {
  const query = useApiQuery({
    queryKey: ["forum", "author-name", authorId],
    queryFn: () => expertApi.getExpertProfile(authorId).catch(() => null),
    enabled: !!authorId,
  });

  return query.data?.fullName ?? "Community member";
}
