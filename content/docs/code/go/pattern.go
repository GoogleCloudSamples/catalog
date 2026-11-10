import (
	"context"
	"fmt"
	"io"

	dlp "cloud.google.com/go/dlp/apiv2"
	"cloud.google.com/go/dlp/apiv2/dlppb"
)

// listInfoTypes prints the sensitive data detectors (infoTypes) supported by
// Cloud Data Loss Prevention (DLP) for risk analysis.
func listInfoTypes(w io.Writer) error {
	ctx := context.Background()
	client, err := dlp.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("dlp.NewClient: %w", err)
	}
	defer client.Close()

	req := &dlppb.ListInfoTypesRequest{
		// Return infoTypes that risk analysis jobs can detect.
		Filter: "supported_by=RISK_ANALYSIS",
	}

	resp, err := client.ListInfoTypes(ctx, req)
	if err != nil {
		return fmt.Errorf("client.ListInfoTypes: %w", err)
	}

	fmt.Fprintln(w, "Info types found:")
	for _, t := range resp.GetInfoTypes() {
		fmt.Fprintf(w, "Name: %s\n", t.GetName())
		fmt.Fprintf(w, "Display name: %s\n", t.GetDisplayName())
	}
	return nil
}
