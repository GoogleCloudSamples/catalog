package cmd

import (
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/snippets"
	"github.com/spf13/cobra"
)

func newSnippetsCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "snippets",
		Short: "Validate multi-language code snippets for region tags, author notes, and syntax",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runChecks(cmd.OutOrStdout(), opts, snippets.Check)
		},
	}
}
