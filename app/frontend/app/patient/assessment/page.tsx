"use client";

import { AssessmentCard } from "./component/AssessmentCard";
import { AssessmentHistoryList } from "./component/AssessmentHistoryList";
import { Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";

export default function AssessmentListPage() {
  const { data: templates = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentTemplates(),
    queryFn: () => assessmentApi.getTemplates(),
  });

  const { data: history = [], isLoading: isHistoryLoading } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentHistory(),
    queryFn: () => assessmentApi.getMyAssessments(),
  });

  return (
    <div className="mx-auto max-w-5xl space-y-10">
      <div>
        <h1 className="mb-2 text-2xl font-bold text-foreground">Psychological Assessments</h1>
        <p className="mb-6 text-sm text-muted-foreground">
          Take a short assessment to help your expert understand you better.
        </p>

        {isLoading ? (
          <Spinner className="h-6 w-6" />
        ) : templates.length === 0 ? (
          <p className="text-sm text-muted-foreground">No assessments are available right now.</p>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {templates.map((template) => (
              <AssessmentCard key={template.slug} template={template} />
            ))}
          </div>
        )}
      </div>

      <div>
        <h2 className="mb-4 text-lg font-bold text-foreground">My Assessment History</h2>
        {isHistoryLoading ? <Spinner className="h-6 w-6" /> : <AssessmentHistoryList history={history} />}
      </div>
    </div>
  );
}
