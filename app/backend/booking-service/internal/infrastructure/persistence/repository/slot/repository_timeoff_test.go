package repository

import (
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProcessedTimeOffIsExcludedByAvailableSlotQuery(t *testing.T) {
	db := dryRunPostgres(t)
	query := withoutActiveTimeOff(db.Table(`"Booking_Expert_Slots"`)).Find(&[]struct{}{})
	assertActiveTimeOffSQL(t, query.Statement.SQL.String())
}

func dryRunPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=localhost user=test dbname=test sslmode=disable"), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func assertActiveTimeOffSQL(t *testing.T, sql string) {
	t.Helper()
	for _, required := range []string{"Booking_Expert_Time_Off", "expert_id", "start_datetime", "end_datetime", "start_time", "end_time", "NOT EXISTS"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("coverage SQL missing %q: %s", required, sql)
		}
	}
	if strings.Contains(strings.ToLower(sql), "processed_at") {
		t.Fatalf("processed time-off would not remain active: %s", sql)
	}
}
