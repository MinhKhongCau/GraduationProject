package appointment

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAttachProfilesSetsExpertAndBookingAccount(t *testing.T) {
	a := Appointment{ExpertID: "e1", PatientID: "p1"}
	a.AttachProfiles(map[string]ParticipantProfile{
		"e1": {AuthID: "e1", Role: "EXPERT", FullName: "Dr. A"},
		"p1": {AuthID: "p1", Role: "PATIENT", FullName: "Booker B"},
	})
	if a.Expert == nil || a.Expert.FullName != "Dr. A" || a.PatientAccount == nil || a.PatientAccount.FullName != "Booker B" {
		t.Fatalf("profiles not attached: %+v / %+v", a.Expert, a.PatientAccount)
	}
}

func TestAppointmentJSONOmitsMissingProfiles(t *testing.T) {
	raw, err := json.Marshal(Appointment{AppointmentID: "a1"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"expert"`) || strings.Contains(string(raw), `"patient_account"`) {
		t.Fatalf("absent profiles must be omitted: %s", raw)
	}
}
