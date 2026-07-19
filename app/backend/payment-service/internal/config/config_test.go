package config

import (
	"testing"
	"time"
)

func TestParsePaymentLifecycleConfig(t *testing.T) {
	ttl, minimum, err := parsePaymentLifecycleConfig("15", "60")
	if err != nil || ttl != 15*time.Minute || minimum != time.Minute {
		t.Fatalf("unexpected config: ttl=%s minimum=%s err=%v", ttl, minimum, err)
	}
	for _, values := range [][2]string{{"0", "60"}, {"61", "60"}, {"bad", "60"}, {"15", "0"}, {"1", "61"}} {
		if _, _, err := parsePaymentLifecycleConfig(values[0], values[1]); err == nil {
			t.Fatalf("expected invalid config for ttl=%q minimum=%q", values[0], values[1])
		}
	}
}
