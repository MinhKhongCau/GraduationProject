export interface MedicalHistory {
  historyId: string;
  conditionName: string;
  description?: string;
  diagnosedAt: string;
  isChronic: boolean;
  isActive: boolean;
}

export interface CreateMedicalHistoryRequest {
  conditionName: string;
  description?: string;
  diagnosedAt: string;
  isChronic: boolean;
}
