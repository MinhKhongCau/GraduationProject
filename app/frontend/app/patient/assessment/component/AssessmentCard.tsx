import Link from "next/link";
import { ClipboardCheck, ArrowRight } from "lucide-react";
import { Card } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { AssessmentTemplate } from "@/types";

export function AssessmentCard({ template }: { template: AssessmentTemplate }) {
  return (
    <Link href={ROUTES.PATIENT.ASSESSMENT_DETAIL(template.slug)} className="group block rounded-xl">
      <Card className="flex h-full flex-col p-6 transition-all duration-200 group-hover:border-primary/40 group-hover:shadow-elevated">
        <span className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-primary-soft text-primary">
          <ClipboardCheck className="h-6 w-6" />
        </span>
        <h3 className="mb-2 font-semibold text-foreground">{template.title}</h3>
        {template.description && (
          <p className="mb-4 flex-1 text-sm text-muted-foreground">{template.description}</p>
        )}
        <span className="inline-flex items-center gap-1 text-sm font-semibold text-primary">
          Start assessment <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
        </span>
      </Card>
    </Link>
  );
}
