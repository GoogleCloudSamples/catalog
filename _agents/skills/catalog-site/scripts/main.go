// The validate_catalog command validates Hugo shortcodes, Markdown links and
// anchors, and multi-language code snippets in the Cloud Samples Catalog
// repository.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		if !errors.Is(err, cmd.ErrValidationFailed) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}
