func exampleSnippet(projectID string) error {
	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", projectID),
		TargetLanguageCode: "es",
		Contents:           []string{"Hello, world!"},
	}

	if _, err := client.TranslateText(ctx, req); err != nil {
		return fmt.Errorf("TranslateText: %w", err)
	}
	return nil
}
