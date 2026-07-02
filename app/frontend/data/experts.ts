import type { ExpertProfile } from "@/types";

/**
 * Curated sample experts for the public landing page's "Featured Experts"
 * section. The authenticated find-experts page calls the real
 * GET /profiles/experts/ endpoint instead — see api/expert.ts.
 */
export const FEATURED_EXPERTS_MOCK: ExpertProfile[] = [
  {
    expertId: "featured-1",
    accountId: "featured-1",
    fullName: "MSc. Nguyen An Binh",
    email: "an.binh@mindcare.example",
    avatarUrl: "https://i.pravatar.cc/150?u=binh",
    bio: "10 years of experience in behavioral therapy and stress management.",
    verificationStatus: "VERIFIED",
    specializations: [
      { specId: "s1", name: "Clinical Psychology", isActive: true },
      { specId: "s2", name: "CBT", isActive: true },
    ],
  },
  {
    expertId: "featured-2",
    accountId: "featured-2",
    fullName: "Dr. Tran Minh Tam",
    email: "minh.tam@mindcare.example",
    avatarUrl: "https://i.pravatar.cc/150?u=tam",
    bio: "Specialist in anxiety and panic disorders, family therapy.",
    verificationStatus: "VERIFIED",
    specializations: [{ specId: "s3", name: "Family Therapy", isActive: true }],
  },
  {
    expertId: "featured-3",
    accountId: "featured-3",
    fullName: "MSc. Le Thi Hanh",
    email: "thi.hanh@mindcare.example",
    avatarUrl: "https://i.pravatar.cc/150?u=hanh",
    bio: "Grief counseling and trauma-informed care.",
    verificationStatus: "VERIFIED",
    specializations: [{ specId: "s4", name: "Trauma & PTSD", isActive: true }],
  },
];
