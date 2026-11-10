package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFormatFlagDescriptions(t *testing.T) {
	rootCmd := NewRootCmd()

	rootFormat := rootCmd.PersistentFlags().Lookup("format")
	if rootFormat == nil {
		t.Fatal("rootCmd PersistentFlags().Lookup(\"format\") = nil, want flag")
	}
	if got, want := rootFormat.Usage, "Output format (text or json)"; got != want {
		t.Errorf("rootCmd --format Usage = %q, want %q", got, want)
	}

	metricsCmd, _, err := rootCmd.Find([]string{"metrics"})
	if err != nil {
		t.Fatalf("rootCmd.Find([\"metrics\"]) unexpected error: %v", err)
	}
	metricsFormat := metricsCmd.Flags().Lookup("format")
	if metricsFormat == nil {
		t.Fatal("metricsCmd Flags().Lookup(\"format\") = nil, want local flag")
	}
	if got, want := metricsFormat.Usage, "Output format (text, markdown, or json)"; got != want {
		t.Errorf("metricsCmd --format Usage = %q, want %q", got, want)
	}
}

func TestValidationRejectsMarkdownFormat(t *testing.T) {
	var out bytes.Buffer
	err := runChecks(&out, &options{repoRoot: t.TempDir(), format: "markdown"})
	if err == nil {
		t.Fatal("runChecks() with format=\"markdown\" succeeded, want error")
	}
	if !strings.Contains(err.Error(), `unsupported output format "markdown"`) {
		t.Errorf("runChecks() error = %q, want substring %q", err.Error(), `unsupported output format "markdown"`)
	}
}

func TestResolveTimeWindow(t *testing.T) {
	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		month     string
		since     string
		until     string
		wantLabel string
		wantErr   string
	}{
		{
			name:      "defaults_to_previous_month",
			wantLabel: "2026-10",
		},
		{
			name:      "explicit_month",
			month:     "2026-09",
			wantLabel: "2026-09",
		},
		{
			name:      "since_and_until",
			since:     "2026-10-01",
			until:     "2026-10-15",
			wantLabel: "2026-10-01 to 2026-10-15",
		},
		{
			name:    "until_without_since",
			until:   "2026-10-15",
			wantErr: "--until requires --since",
		},
		{
			name:    "since_after_until",
			since:   "2026-10-20",
			until:   "2026-10-15",
			wantErr: "cannot be after --until",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, label, err := resolveTimeWindow(tc.month, tc.since, tc.until, now)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("resolveTimeWindow() error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveTimeWindow() unexpected error: %v", err)
			}
			if label != tc.wantLabel {
				t.Errorf("resolveTimeWindow() label = %q, want %q", label, tc.wantLabel)
			}
		})
	}
}
