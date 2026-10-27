package cmd

import "testing"

func TestAddTokenDurationUpdate(t *testing.T) {
	update := map[string]any{}
	addTokenDurationUpdate(update, "max_age", "30 00:00:00")
	if update["max_age"] != "30 00:00:00" {
		t.Fatalf("max_age = %#v", update["max_age"])
	}

	addTokenDurationUpdate(update, "max_unused_period", "")
	if value, ok := update["max_unused_period"]; !ok || value != nil {
		t.Fatalf("max_unused_period = %#v, want explicit nil", value)
	}

	// Duration syntax is validated by the API, not the CLI.
	addTokenDurationUpdate(update, "max_age", "30d")
	if update["max_age"] != "30d" {
		t.Fatalf("max_age = %#v, want passthrough value", update["max_age"])
	}
}
