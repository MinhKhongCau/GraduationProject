export const HERO_CONTENT = {
  eyebrowKey: "landing.hero.eyebrow",
  eyebrow: "Online psychological counseling",
  titleKey: "landing.hero.title",
  title: "Take care of your mind, with people who understand you",
  subtitleKey: "landing.hero.subtitle",
  subtitle:
    "MindCare connects you with licensed psychological experts for safe, convenient, and effective online counseling.",
  primaryCtaKey: "landing.hero.primaryCta",
  primaryCta: "Book an appointment",
  secondaryCtaKey: "landing.hero.secondaryCta",
  secondaryCta: "Find an expert",
};

export interface HowItWorksStep {
  id: string;
  titleKey: string;
  title: string;
  descriptionKey: string;
  description: string;
}

export const HOW_IT_WORKS_STEPS: HowItWorksStep[] = [
  {
    id: "choose-topic",
    titleKey: "landing.howItWorks.chooseTopic.title",
    title: "Tell us what's on your mind",
    descriptionKey: "landing.howItWorks.chooseTopic.description",
    description: "Pick the topics you'd like to talk about so we can match you with the right expert.",
  },
  {
    id: "pick-expert",
    titleKey: "landing.howItWorks.pickExpert.title",
    title: "Choose your expert",
    descriptionKey: "landing.howItWorks.pickExpert.description",
    description: "Browse experts by specialization, experience, and availability.",
  },
  {
    id: "book-slot",
    titleKey: "landing.howItWorks.bookSlot.title",
    title: "Book a time that works for you",
    descriptionKey: "landing.howItWorks.bookSlot.description",
    description: "Select an open slot on your expert's calendar and confirm your session.",
  },
  {
    id: "get-support",
    titleKey: "landing.howItWorks.getSupport.title",
    title: "Get ongoing support",
    descriptionKey: "landing.howItWorks.getSupport.description",
    description: "Track your appointments, assessments, and progress all in one place.",
  },
];

export const TESTIMONIALS = [
  {
    id: "t1",
    nameKey: "landing.testimonials.t1.name",
    name: "Minh Anh",
    quoteKey: "landing.testimonials.t1.quote",
    quote: "MindCare helped me find the right expert within a day. Booking was effortless.",
  },
  {
    id: "t2",
    nameKey: "landing.testimonials.t2.name",
    name: "David Tran",
    quoteKey: "landing.testimonials.t2.quote",
    quote: "The assessment tools gave me real insight before my first session.",
  },
  {
    id: "t3",
    nameKey: "landing.testimonials.t3.name",
    name: "Lan Pham",
    quoteKey: "landing.testimonials.t3.quote",
    quote: "Managing my bookings and wallet in one dashboard is a huge time saver.",
  },
];
