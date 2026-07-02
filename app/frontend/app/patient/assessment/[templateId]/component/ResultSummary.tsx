import Link from "next/link";
import { Sparkles } from "lucide-react";
import { Card, Button } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { AssessmentSubmitResponse } from "@/types";

export function ResultSummary({ result }: { result: AssessmentSubmitResponse }) {
  return (
    <div className="space-y-6">
      <Card className="p-6 text-center">
        <p className="mb-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">Total score</p>
        <h2 className="text-4xl font-extrabold text-foreground">{result.totalScore}</h2>
      </Card>

      {Object.keys(result.dimensionScores).length > 0 && (
        <Card className="p-6">
          <h3 className="mb-3 text-sm font-bold text-foreground">Scores by dimension</h3>
          <div className="space-y-2">
            {Object.entries(result.dimensionScores).map(([dimension, score]) => (
              <div key={dimension} className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">{dimension}</span>
                <span className="font-bold text-foreground">{score}</span>
              </div>
            ))}
          </div>
        </Card>
      )}

      {result.aiEvaluation && (
        <Card className="p-6">
          <h3 className="mb-2 flex items-center gap-1.5 text-sm font-bold text-foreground">
            <Sparkles className="h-4 w-4 text-primary" /> AI evaluation
          </h3>
          <p className="text-sm leading-relaxed text-muted-foreground">{result.aiEvaluation}</p>
        </Card>
      )}

      <Link href={ROUTES.PATIENT.ASSESSMENT}>
        <Button variant="outline">Back to assessments</Button>
      </Link>
    </div>
  );
}
