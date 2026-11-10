package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/metrics"
	"github.com/spf13/cobra"
)

type metricsFlags struct {
	month  string
	since  string
	until  string
	repos  []string
	format string
}

func newMetricsCmd() *cobra.Command {
	flags := &metricsFlags{}

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Collect monthly created and updated sample metrics across sample repositories",
		Long:  "Queries the GitHub REST API across the Google Cloud sample repositories to identify samples created and updated in a target month (excluding Dependabot and automated dependency bumps).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			since, until, label, err := resolveTimeWindow(flags.month, flags.since, flags.until, time.Now().UTC())
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			token := metrics.ResolveGitHubToken(ctx)

			rep, err := metrics.Collect(ctx, metrics.Options{
				Since:      since,
				Until:      until,
				MonthLabel: label,
				Repos:      flags.repos,
				Token:      token,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			switch strings.ToLower(strings.TrimSpace(flags.format)) {
			case "text":
				return metrics.WriteText(out, rep)
			case "markdown", "md":
				return metrics.WriteMarkdown(out, rep)
			case "json":
				return metrics.WriteJSON(out, rep)
			default:
				return fmt.Errorf("unsupported output format %q (want \"text\", \"markdown\", or \"json\")", flags.format)
			}
		},
	}

	cmd.Flags().StringVar(&flags.month, "month", "", "Target calendar month in YYYY-MM format (defaults to previous month)")
	cmd.Flags().StringVar(&flags.since, "since", "", "Start date in YYYY-MM-DD format (overrides --month)")
	cmd.Flags().StringVar(&flags.until, "until", "", "End date in YYYY-MM-DD format (inclusive, used with --since)")
	cmd.Flags().StringSliceVar(&flags.repos, "repos", nil, "Comma-separated list of owner/repo names (defaults to the 8 Google Cloud sample repos)")
	cmd.Flags().StringVar(&flags.format, "format", "text", "Output format (text, markdown, or json)")

	return cmd
}

func resolveTimeWindow(month, sinceStr, untilStr string, now time.Time) (time.Time, time.Time, string, error) {
	if sinceStr == "" && untilStr != "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--until requires --since")
	}

	if sinceStr != "" {
		sinceDate, err := time.Parse("2006-01-02", strings.TrimSpace(sinceStr))
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("parsing --since %q (want YYYY-MM-DD): %w", sinceStr, err)
		}
		since := time.Date(sinceDate.Year(), sinceDate.Month(), sinceDate.Day(), 0, 0, 0, 0, time.UTC)

		until := now
		if untilStr != "" {
			untilDate, err := time.Parse("2006-01-02", strings.TrimSpace(untilStr))
			if err != nil {
				return time.Time{}, time.Time{}, "", fmt.Errorf("parsing --until %q (want YYYY-MM-DD): %w", untilStr, err)
			}
			until = time.Date(untilDate.Year(), untilDate.Month(), untilDate.Day(), 23, 59, 59, 0, time.UTC)
		}
		if since.After(until) {
			return time.Time{}, time.Time{}, "", fmt.Errorf("--since date %s cannot be after --until date %s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		}
		label := fmt.Sprintf("%s to %s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		return since, until, label, nil
	}

	if month == "" {
		firstOfCurrent := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		prevMonth := firstOfCurrent.AddDate(0, -1, 0)
		month = prevMonth.Format("2006-01")
	}

	since, until, err := metrics.ParseMonthWindow(month)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	return since, until, month, nil
}
