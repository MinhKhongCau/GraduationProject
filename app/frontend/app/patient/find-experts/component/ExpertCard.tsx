import Link from "next/link";
import Image from "next/image";
import { BadgeCheck } from "lucide-react";
import { Card, Button } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { ExpertProfile } from "@/types";

export function ExpertCard({ expert }: { expert: ExpertProfile }) {
  return (
    <Card className="flex flex-col p-6">
      <div className="mb-4 flex items-center gap-3">
        <Image
          src={expert.avatarUrl || "/images/auth-bg.png"}
          alt={expert.fullName}
          width={48}
          height={48}
          unoptimized
          className="h-12 w-12 rounded-full object-cover"
        />
        <div>
          <div className="flex items-center gap-1">
            <h3 className="font-bold text-foreground">{expert.fullName}</h3>
            {expert.verificationStatus === "VERIFIED" && (
              <BadgeCheck className="h-4 w-4 text-primary" aria-label="Verified" />
            )}
          </div>
          <p className="text-xs text-muted-foreground">
            {expert.specializations.map((spec) => spec.name).join(", ") || "General counseling"}
          </p>
        </div>
      </div>
      {expert.bio && <p className="mb-4 line-clamp-3 flex-1 text-sm text-muted-foreground">{expert.bio}</p>}
      <Link href={ROUTES.PATIENT.EXPERT_DETAIL(expert.accountId)}>
        <Button size="sm" className="w-full">
          View profile
        </Button>
      </Link>
    </Card>
  );
}
