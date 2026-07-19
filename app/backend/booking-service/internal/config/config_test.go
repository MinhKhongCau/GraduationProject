package config

import "testing"

func TestRollingSlotDaysDefaultAndValidation(t *testing.T) {
	if got, err := ParseRollingSlotDays("30"); err != nil || got != 30 {
		t.Fatalf("default rolling window: got %d, err %v", got, err)
	}
	for _, value := range []string{"", "0", "-1", "366", "abc"} {
		if _, err := ParseRollingSlotDays(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}
