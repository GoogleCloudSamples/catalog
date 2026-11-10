func createWorkflow(w io.Writer, projectID, locationID, workflowID string) error {
	ctx := context.Background()
	client, err := workflows.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("workflows.NewClient: %w", err)
	}
	defer client.Close()

	// CreateWorkflow starts a long-running operation. The workflow isn't
	// ready until the operation completes.
	op, err := client.CreateWorkflow(ctx, &workflowspb.CreateWorkflowRequest{
		Parent:     fmt.Sprintf("projects/%s/locations/%s", projectID, locationID),
		Workflow:   &workflowspb.Workflow{SourceContents: "- done:\n    return: ok\n"},
		WorkflowId: workflowID,
	})
	if status.Code(err) == codes.AlreadyExists {
		fmt.Fprintf(w, "error: workflow %q already exists in %q\n", workflowID, locationID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("CreateWorkflow: %w", err)
	}
	meta, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("reading operation metadata: %w", err)
	}
	fmt.Fprintf(w, "Started operation %q at %s\n", op.Name(), meta.GetCreateTime().AsTime().Format(time.RFC3339))

	// Poll for up to 90 seconds; the operation may take longer to complete.
	waitCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	workflow, err := op.Wait(waitCtx)
	if err != nil && errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
		// Leaving the wait does not cancel the operation; it continues
		// running on the server.
		fmt.Fprintf(w, "Operation %q is still running; use this operation name to check progress later\n", op.Name())
		return nil
	}
	if status.Code(err) == codes.AlreadyExists {
		fmt.Fprintf(w, "error: operation %q failed: workflow %q already exists\n", op.Name(), workflowID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("operation %q failed: %w", op.Name(), err)
	}
	fmt.Fprintf(w, "Created workflow %s (revision %s)\n", workflow.GetName(), workflow.GetRevisionId())
	return nil
}
