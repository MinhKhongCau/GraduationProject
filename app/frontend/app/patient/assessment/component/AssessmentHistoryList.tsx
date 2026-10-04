import { Badge, Card } from "@/components/ui";
import type { AssessmentHistoryItem } from "@/types";

export interface AssessmentHistoryListProps {
  history: AssessmentHistoryItem[];
}

export function AssessmentHistoryList({ history }: AssessmentHistoryListProps) {
  if (history.length === 0) {
    return <p className="text-sm text-muted-foreground">You haven&apos;t submitted any assessments yet.</p>;
  }

  return (
    <div className="space-y-3">
      {history.map((item) => (
        <Card key={item.resultId} className="p-4">
          <div className="mb-1 flex items-center justify-between gap-3">
            <div>
              <Badge tone="primary" className="uppercase">
                {item.templateCode}
              </Badge>
              <p className="mt-1.5 text-sm font-semibold text-foreground">{item.templateTitle}</p>
            </div>
            <div className="text-right">
              <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Score</p>
              <p className="text-lg font-extrabold text-foreground">{item.totalScore}</p>
            </div>
          </div>

          {Object.keys(item.dimensionScores).length > 0 && (
            <div className="mt-2 flex flex-wrap gap-2">
              {Object.entries(item.dimensionScores).map(([dimension, score]) => (
                <span
                  key={dimension}
                  className="rounded-lg border border-border bg-surface px-2 py-1 text-xs text-muted-foreground"
                >
                  {dimension}: <span className="font-semibold text-foreground">{score}</span>
                </span>
              ))}
            </div>
          )}

          <p className="mt-2 text-xs text-muted-foreground">
            Submitted {new Date(item.createdAt).toLocaleDateString()}
          </p>
        </Card>
      ))}
    </div>
  );
}
