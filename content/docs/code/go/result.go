func exampleSnippet(w io.Writer, projectID, filePath string) error {
	// Snippet content ...
	fmt.Fprintf(w, "Response values: %s, %s\n", resp.GetExampleName(), resp.GetExampleValue())
	return nil
}
