import Link from "next/link";
import { buttonClasses } from "@/components/ui";
import { ROUTES } from "@/constants";

export function CtaSection() {
  return (
    <section className="px-4 py-16 sm:px-6 sm:py-20">
      <div className="mx-auto flex max-w-4xl flex-col items-center rounded-2xl bg-primary px-6 py-12 text-center text-white shadow-elevated sm:px-10 sm:py-14">
        <h2 className="mb-4 text-3xl font-bold tracking-tight">Ready to take the first step?</h2>
        <p className="mb-8 max-w-xl text-primary-soft">
          Create your free account and get matched with an expert who understands you.
        </p>
        <Link href={ROUTES.AUTH.REGISTER} className={buttonClasses("outline", "lg", "border-white bg-white text-primary hover:bg-primary-soft")}>
          Create your account
        </Link>
      </div>
    </section>
  );
}
