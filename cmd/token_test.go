package cmd

import "testing"

func TestValidateTokenDuration(t *testing.T) {
	valid := []string{
		"",
		"3600",
		"3600.000000",
		"00:00:00",
		"01:00:00",
		"30 00:00:00",
		"7 00:00:00",
		"1 02:03:04.123456",
	}
	for _, value := range valid {
		if err := validateTokenDuration("--max-age", value); err != nil {
			t.Errorf("validateTokenDuration(%q) error = %v", value, err)
		}
	}

	invalid := []string{"30d", "7d", "1h", "-1", "1 2 3", "01:60:00x"}
	for _, value := range invalid {
		if err := validateTokenDuration("--max-age", value); err == nil {
			t.Errorf("validateTokenDuration(%q) = nil, want error", value)
		}
	}
}

func TestAddTokenDurationUpdate(t *testing.T) {
	update := map[string]any{}
	if err := addTokenDurationUpdate(update, "max_age", "30 00:00:00"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if update["max_age"] != "30 00:00:00" {
		t.Fatalf("max_age = %#v", update["max_age"])
	}

	if err := addTokenDurationUpdate(update, "max_unused_period", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value, ok := update["max_unused_period"]; !ok || value != nil {
		t.Fatalf("max_unused_period = %#v, want explicit nil", value)
	}

	if err := addTokenDurationUpdate(update, "max_age", "30d"); err == nil {
		t.Fatal("expected invalid duration to fail")
	}
}
