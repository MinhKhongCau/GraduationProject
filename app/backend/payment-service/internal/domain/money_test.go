package domain

import "testing"

func TestMoneyOperations(t *testing.T) {
	m1 := Money(1000)
	m2 := Money(500)

	if m1.Add(m2) != Money(1500) {
		t.Errorf("Expected 1000 + 500 = 1500, got %d", m1.Add(m2))
	}

	if m1.Sub(m2) != Money(500) {
		t.Errorf("Expected 1000 - 500 = 500, got %d", m1.Sub(m2))
	}

	if m1.MulFloat(0.15) != Money(150) {
		t.Errorf("Expected 1000 * 0.15 = 150, got %d", m1.MulFloat(0.15))
	}

	if !Money(0).IsZero() {
		t.Error("Expected Money(0).IsZero() to be true")
	}

	if !Money(10).IsPositive() {
		t.Error("Expected Money(10).IsPositive() to be true")
	}

	if !Money(-5).IsNegative() {
		t.Error("Expected Money(-5).IsNegative() to be true")
	}
}
