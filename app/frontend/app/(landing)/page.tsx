import { Hero } from "./component/Hero";
import { HowItWorks } from "./component/HowItWorks";
import { FeaturedExperts } from "./component/FeaturedExperts";
import { Testimonials } from "./component/Testimonials";
import { CtaSection } from "./component/CtaSection";

export default function LandingPage() {
  return (
    <>
      <Hero />
      <HowItWorks />
      <FeaturedExperts />
      <Testimonials />
      <CtaSection />
    </>
  );
}
