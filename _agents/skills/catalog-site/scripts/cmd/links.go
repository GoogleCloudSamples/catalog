package cmd

import (
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/links"
	"github.com/spf13/cobra"
)

func newLinksCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "links",
		Short: "Validate internal Markdown relative links, heading anchors, and content hygiene",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runChecks(cmd.OutOrStdout(), opts, links.Check)
		},
	}
}
