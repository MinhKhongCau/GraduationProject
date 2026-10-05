import Image from "next/image";
import { BadgeCheck, Languages, GraduationCap } from "lucide-react";
import { Card } from "@/components/ui";
import type { ExpertProfile } from "@/types";

export function ExpertProfileHeader({ expert }: { expert: ExpertProfile }) {
  return (
    <Card className="p-6">
      <div className="flex flex-col items-center gap-4 text-center sm:flex-row sm:items-start sm:text-left">
        <Image
          src={expert.avatarUrl || "/images/auth-bg.png"}
          alt={expert.fullName}
          width={96}
          height={96}
          unoptimized
          className="h-24 w-24 rounded-full border border-border object-cover"
        />
        <div className="flex-1">
          <div className="flex items-center justify-center gap-1.5 sm:justify-start">
            <h1 className="text-xl font-bold tracking-tight text-foreground">{expert.fullName}</h1>
            {expert.verificationStatus === "VERIFIED" && (
              <BadgeCheck className="h-5 w-5 text-primary" aria-label="Verified" />
            )}
          </div>
          <p className="mb-3 text-sm text-muted-foreground">
            {expert.specializations.map((spec) => spec.name).join(", ") || "General counseling"}
          </p>
          {expert.bio && <p className="text-sm text-foreground">{expert.bio}</p>}
          <div className="mt-4 flex flex-wrap justify-center gap-2 text-xs text-muted-foreground sm:justify-start">
            <span className="flex items-center gap-1 rounded-lg border border-border bg-surface px-2.5 py-1">
              <GraduationCap className="h-3.5 w-3.5" /> {expert.specializations.length} specializations
            </span>
            <span className="flex items-center gap-1 rounded-lg border border-border bg-surface px-2.5 py-1">
              <Languages className="h-3.5 w-3.5" /> Vietnamese, English
            </span>
          </div>
        </div>
      </div>
    </Card>
  );
}
