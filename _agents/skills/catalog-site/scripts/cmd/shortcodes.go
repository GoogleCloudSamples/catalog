package cmd

import (
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/shortcodes"
	"github.com/spf13/cobra"
)

func newShortcodesCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "shortcodes",
		Short: "Validate Hugo guideline_examples shortcodes and code file mappings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runChecks(cmd.OutOrStdout(), opts, shortcodes.Check)
		},
	}
}
