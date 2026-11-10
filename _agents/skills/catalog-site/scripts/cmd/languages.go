package cmd

import (
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/languages"
	"github.com/spf13/cobra"
)

func newLanguagesCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "languages",
		Short: "Validate that every language in languages.html is fully set up",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runChecks(cmd.OutOrStdout(), opts, languages.Check)
		},
	}
}
