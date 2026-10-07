export type PatientRecordRelationship =
  | "SELF"
  | "PARENT"
  | "CHILD"
  | "SPOUSE"
  | "SIBLING"
  | "RELATIVE"
  | "OTHER";

export type PatientRecordGender = "MALE" | "FEMALE" | "OTHER";

/** A person the logged-in patient can book for. The SELF record is auto-created server-side. */
export interface PatientRecord {
  recordId: string;
  ownerAuthId: string;
  fullName: string;
  dateOfBirth: string | null;
  gender: PatientRecordGender | "";
  phoneNumber: string;
  email: string;
  address: string;
  relationship: PatientRecordRelationship;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePatientRecordRequest {
  fullName: string;
  /** YYYY-MM-DD */
  dateOfBirth: string;
  gender: PatientRecordGender;
  phoneNumber: string;
  email?: string;
  address?: string;
  /** SELF is reserved for the auto-created record. */
  relationship: Exclude<PatientRecordRelationship, "SELF">;
}
