package cmd

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// repeatStr
// ---------------------------------------------------------------------------

func TestRepeatStr_BasicRepeat(t *testing.T) {
	got := repeatStr("█", 3)
	if got != "███" {
		t.Errorf("expected %q, got %q", "███", got)
	}
}

func TestRepeatStr_ZeroCount(t *testing.T) {
	got := repeatStr("█", 0)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestRepeatStr_MultiCharString(t *testing.T) {
	got := repeatStr("ab", 3)
	if got != "ababab" {
		t.Errorf("expected %q, got %q", "ababab", got)
	}
}

// ---------------------------------------------------------------------------
// progressBar
// ---------------------------------------------------------------------------

func TestProgressBar_ZeroTotal_ReturnsDash(t *testing.T) {
	got := progressBar(0, 0)
	// The rendered output wraps the dash in ANSI codes; check it contains "—".
	if !strings.Contains(got, "—") {
		t.Errorf("expected output to contain %q when total is 0, got %q", "—", got)
	}
}

func TestProgressBar_ZeroPercent(t *testing.T) {
	got := progressBar(0, 100)
	if !strings.Contains(got, "0%") {
		t.Errorf("expected output to contain %q, got %q", "0%", got)
	}
}

func TestProgressBar_FiftyPercent(t *testing.T) {
	got := progressBar(50, 100)
	if !strings.Contains(got, "50%") {
		t.Errorf("expected output to contain %q, got %q", "50%", got)
	}
}

func TestProgressBar_HundredPercent(t *testing.T) {
	got := progressBar(100, 100)
	if !strings.Contains(got, "100%") {
		t.Errorf("expected output to contain %q, got %q", "100%", got)
	}
}

func TestProgressBar_CapsAtHundred(t *testing.T) {
	// current > total should not produce a percentage above 100.
	got := progressBar(500, 100)
	if !strings.Contains(got, "100%") {
		t.Errorf("expected output to be capped at 100%%, got %q", got)
	}
}

func TestProgressBar_PartialProgress(t *testing.T) {
	// 180 / 380 ≈ 47%
	got := progressBar(180, 380)
	if !strings.Contains(got, "47%") {
		t.Errorf("expected output to contain %q, got %q", "47%", got)
	}
}

// ---------------------------------------------------------------------------
// Command registration
// ---------------------------------------------------------------------------

func TestCommandsRegistered(t *testing.T) {
	want := []string{"add", "list", "del", "version"}

	registered := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		registered[c.Name()] = true
	}

	for _, name := range want {
		if !registered[name] {
			t.Errorf("expected command %q to be registered on rootCmd", name)
		}
	}
}
