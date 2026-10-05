"use client";

import { use } from "react";
import Link from "next/link";
import { CalendarPlus } from "lucide-react";
import { ExpertProfileHeader } from "./component/ExpertProfileHeader";
import { buttonClasses, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { expertApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";

export default function ExpertDetailPage({ params }: { params: Promise<{ expertId: string }> }) {
  const { expertId } = use(params);

  const { data: expert, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.expertProfile(expertId),
    queryFn: () => expertApi.getExpertProfile(expertId),
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (!expert) {
    return <p className="text-sm text-muted-foreground">Expert not found.</p>;
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <ExpertProfileHeader expert={expert} />
      <Link
        href={`${ROUTES.PATIENT.BOOK_APPOINTMENT}?expertId=${expertId}`}
        className={buttonClasses("primary", "lg", "w-full sm:w-auto")}
      >
        <CalendarPlus className="h-4 w-4" />
        Book with {expert.fullName}
      </Link>
    </div>
  );
}
