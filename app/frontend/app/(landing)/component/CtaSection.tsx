import Link from "next/link";
import { Button } from "@/components/ui";
import { ROUTES } from "@/constants";

export function CtaSection() {
  return (
    <section className="px-6 py-20">
      <div className="mx-auto flex max-w-4xl flex-col items-center rounded-3xl bg-primary px-8 py-14 text-center text-white">
        <h2 className="mb-4 text-3xl font-bold">Ready to take the first step?</h2>
        <p className="mb-8 max-w-xl text-primary-soft">
          Create your free account and get matched with an expert who understands you.
        </p>
        <Link href={ROUTES.AUTH.REGISTER}>
          <Button size="lg" variant="soft">
            Create your account
          </Button>
        </Link>
      </div>
    </section>
  );
}
