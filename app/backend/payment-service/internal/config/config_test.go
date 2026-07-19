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

func TestParseOutboxConfig(t *testing.T) {
	cfg, err := parseOutboxConfig("5", "10", "5", "300", "50")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PollInterval != 5*time.Second || cfg.MaxAttempts != 10 || cfg.BaseBackoff != 5*time.Second || cfg.MaxBackoff != 5*time.Minute || cfg.BatchSize != 50 {
		t.Fatalf("unexpected outbox config: %+v", cfg)
	}

	invalid := [][5]string{
		{"0", "10", "5", "300", "50"},
		{"5", "0", "5", "300", "50"},
		{"5", "10", "0", "300", "50"},
		{"5", "10", "301", "300", "50"},
		{"5", "10", "5", "0", "50"},
		{"5", "10", "5", "300", "0"},
		{"3601", "10", "5", "300", "50"},
		{"5", "101", "5", "300", "50"},
		{"5", "10", "5", "86401", "50"},
		{"5", "10", "5", "300", "1001"},
		{"bad", "10", "5", "300", "50"},
	}
	for _, values := range invalid {
		if _, err := parseOutboxConfig(values[0], values[1], values[2], values[3], values[4]); err == nil {
			t.Fatalf("expected invalid outbox config for %v", values)
		}
	}
}
