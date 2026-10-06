package appointment

import (
	"context"
	"errors"
)

var (
	ErrBookingProfileNotFound         = errors.New("expert or patient record not found")
	ErrBookingProfileInvalid          = errors.New("invalid expert, patient record or specialization id")
	ErrBookingSpecializationMismatch  = errors.New("expert does not offer the selected specialization")
	ErrProfileServiceUnavailable      = errors.New("profile service unavailable")
	ErrBookingSlotNotHeld             = errors.New("slot is not held by patient or the hold expired")
	ErrBookingConfirmationUnavailable = errors.New("booking confirmation unavailable")
)

// ProfileGateway truy vấn profile-service (triển khai bằng gRPC ở infrastructure/client).
type ProfileGateway interface {
	GetBookingInfo(ctx context.Context, query BookingInfoQuery) (*BookingInfo, error)
}

type BookingInfoQuery struct {
	ExpertID         string
	PatientRecordID  string
	OwnerID          string
	SpecializationID string // tuỳ chọn
}

type BookingInfo struct {
	Expert         ExpertInfo          `json:"expert"`
	PatientRecord  PatientRecordInfo   `json:"patient_record"`
	Specialization *SpecializationInfo `json:"specialization"`
}

type ExpertInfo struct {
	ExpertID           string               `json:"expert_id"`
	FullName           string               `json:"full_name"`
	AvatarURL          string               `json:"avatar_url"`
	Email              string               `json:"email"`
	VerificationStatus string               `json:"verification_status"`
	Specializations    []SpecializationInfo `json:"specializations"`
}

type SpecializationInfo struct {
	SpecID string `json:"spec_id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
}

type PatientRecordInfo struct {
	RecordID     string `json:"record_id"`
	FullName     string `json:"full_name"`
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Address      string `json:"address"`
	Relationship string `json:"relationship"`
}
