package appointmentpostgres

import (
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProcessedTimeOffBlocksDirectSlotLockQuery(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=test dbname=test sslmode=disable"), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	query := withoutActiveTimeOffForSlot(db.Table(`"Booking_Expert_Slots"`)).Find(&[]struct{}{})
	sql := query.Statement.SQL.String()
	for _, required := range []string{"Booking_Expert_Time_Off", "expert_id", "start_datetime", "end_datetime", "start_time", "end_time", "NOT EXISTS"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("lock coverage SQL missing %q: %s", required, sql)
		}
	}
	if strings.Contains(strings.ToLower(sql), "processed_at") {
		t.Fatalf("processed time-off would not block lock: %s", sql)
	}
}
