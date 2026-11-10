public static void createWorkflow(String projectId, String locationId, String workflowId)
    throws IOException, InterruptedException, ExecutionException {
  try (WorkflowsClient client = WorkflowsClient.create()) {
    CreateWorkflowRequest request =
        CreateWorkflowRequest.newBuilder()
            .setParent(LocationName.of(projectId, locationId).toString())
            .setWorkflow(Workflow.newBuilder().setSourceContents("- done:\n    return: ok\n"))
            .setWorkflowId(workflowId)
            .build();
    OperationFuture<Workflow, OperationMetadata> operation;
    String name;
    OperationMetadata metadata;
    try {
      // createWorkflowAsync starts a long-running operation. The workflow isn't
      // ready until the operation completes.
      operation = client.createWorkflowAsync(request);
      name = operation.getName();
      metadata = operation.peekMetadata().get();
    } catch (ExecutionException e) {
      if (e.getCause() instanceof AlreadyExistsException) {
        System.err.printf("error: workflow '%s' already exists in '%s'%n", workflowId, locationId);
        return;
      }
      throw e;
    }
    System.out.printf(
        "Started operation '%s' at %s%n", name, Timestamps.toString(metadata.getCreateTime()));

    try {
      // Poll for up to 90 seconds; the operation may take longer to complete.
      Workflow workflow = operation.get(90, TimeUnit.SECONDS);
      System.out.printf(
          "Created workflow %s (revision %s)%n", workflow.getName(), workflow.getRevisionId());
    } catch (TimeoutException e) {
      // Leaving the wait does not cancel the operation; it continues running
      // on the server.
      System.out.printf(
          "Operation '%s' is still running; use this operation name to check progress later%n",
          name);
    } catch (ExecutionException e) {
      if (e.getCause() instanceof AlreadyExistsException) {
        System.err.printf("error: workflow '%s' already exists in '%s'%n", workflowId, locationId);
        return;
      }
      System.err.printf("error: operation '%s' failed: %s%n", name, e.getCause().getMessage());
    }
  }
}
