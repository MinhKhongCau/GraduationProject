import Image from "next/image";
import { FEATURED_EXPERTS_MOCK } from "@/data";
import { Card } from "@/components/ui";

export function FeaturedExperts() {
  return (
    <section className="bg-surface px-6 py-20">
      <div className="mx-auto max-w-6xl">
        <h2 className="mb-12 text-center text-3xl font-bold text-foreground">Meet our experts</h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {FEATURED_EXPERTS_MOCK.map((expert) => (
            <Card key={expert.expertId} className="p-6 text-center">
              <Image
                src={expert.avatarUrl ?? "/images/auth-bg.png"}
                alt={expert.fullName}
                width={64}
                height={64}
                className="mx-auto mb-4 h-16 w-16 rounded-full object-cover"
                unoptimized
              />
              <h3 className="font-bold text-foreground">{expert.fullName}</h3>
              <p className="mb-3 text-xs text-muted-foreground">
                {expert.specializations.map((spec) => spec.name).join(", ")}
              </p>
              <p className="text-sm text-muted-foreground">{expert.bio}</p>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
