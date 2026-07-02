import Link from "next/link";
import { ClipboardCheck, ArrowRight } from "lucide-react";
import { Card } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { AssessmentTemplate } from "@/types";

export function AssessmentCard({ template }: { template: AssessmentTemplate }) {
  return (
    <Link href={ROUTES.PATIENT.ASSESSMENT_DETAIL(template.templateId)}>
      <Card className="flex h-full flex-col p-6 transition-shadow hover:shadow-elevated">
        <ClipboardCheck className="mb-4 h-8 w-8 text-primary" />
        <h3 className="mb-2 font-bold text-foreground">{template.title}</h3>
        {template.description && (
          <p className="mb-4 flex-1 text-sm text-muted-foreground">{template.description}</p>
        )}
        <span className="inline-flex items-center gap-1 text-sm font-semibold text-primary">
          Start assessment <ArrowRight className="h-3.5 w-3.5" />
        </span>
      </Card>
    </Link>
  );
}
