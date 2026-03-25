package decoder

import (
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		Name      string    `mapstructure:"name"`
		StartedAt time.Time `mapstructure:"started_at"`
	}

	input := map[string]any{
		"name":       "alpha",
		"started_at": "2026-03-25T08:30:00Z",
	}

	var cfg testConfig
	err := Decode(input, &cfg)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if cfg.Name != "alpha" {
		t.Fatalf("Decode() name = %q, want %q", cfg.Name, "alpha")
	}

	expected := time.Date(2026, 3, 25, 8, 30, 0, 0, time.UTC)
	if !cfg.StartedAt.Equal(expected) {
		t.Fatalf("Decode() started_at = %v, want %v", cfg.StartedAt, expected)
	}
}

func TestDecodeInvalidTime(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		StartedAt time.Time `mapstructure:"started_at"`
	}

	input := map[string]any{
		"started_at": "not-a-time",
	}

	var cfg testConfig
	err := Decode(input, &cfg)
	if err == nil {
		t.Fatal("Decode() error = nil, want non-nil")
	}
}
