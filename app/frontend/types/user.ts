import type { UserRole } from "./auth";
import type { Specialization } from "./specialization";
import type { MedicalHistory } from "./medical-history";

/** Session-shaped user, derived from LoginResponse + rehydrated profile. */
export interface User {
  id: string;
  fullName: string;
  email: string;
  role: UserRole;
  avatarUrl?: string;
}

export interface PatientProfile {
  patientId: string;
  accountId: string;
  fullName: string;
  phoneNumber?: string;
  email: string;
  avatarUrl?: string;
  dateOfBirth?: string;
  gender?: string;
  address?: string;
  medicalHistories?: MedicalHistory[];
}

export interface UpdatePatientProfileRequest {
  fullName: string;
  phoneNumber?: string;
  email?: string;
  avatarUrl?: string;
  dateOfBirth: string;
  gender?: string;
  address?: string;
}

export type ExpertVerificationStatus = "UNVERIFIED" | "PENDING" | "VERIFIED" | "REJECTED";

export interface ExpertProfile {
  expertId: string;
  accountId: string;
  fullName: string;
  phoneNumber?: string;
  email: string;
  avatarUrl?: string;
  introductionVideoUrl?: string;
  bio?: string;
  verificationStatus: ExpertVerificationStatus;
  specializations: Specialization[];
}

export interface CreateExpertProfileRequest {
  fullName: string;
  phoneNumber?: string;
  email: string;
  avatarUrl?: string;
  introductionVideoUrl?: string;
  bio?: string;
  specializationIds: string[];
}
