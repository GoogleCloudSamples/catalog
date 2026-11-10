package cmd

import (
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/templates"
	"github.com/spf13/cobra"
)

func newTemplatesCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "Statically validate the layouts that render the agent/LLM text outputs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runChecks(cmd.OutOrStdout(), opts, templates.Check)
		},
	}
}
