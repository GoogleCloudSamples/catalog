public void CreateWorkflow(string projectId, string locationId, string workflowId)
{
    var client = WorkflowsClient.Create();
    var request = new CreateWorkflowRequest
    {
        ParentAsLocationName = LocationName.FromProjectLocation(projectId, locationId),
        Workflow = new Workflow { SourceContents = "- done:\n    return: ok\n" },
        WorkflowId = workflowId,
    };
    Operation<Workflow, OperationMetadata> operation;
    try
    {
        // CreateWorkflow starts a long-running operation. The workflow isn't
        // ready until the operation completes.
        operation = client.CreateWorkflow(request);
    }
    catch (RpcException e) when (e.StatusCode == StatusCode.AlreadyExists)
    {
        Console.Error.WriteLine($"error: workflow '{workflowId}' already exists in '{locationId}'");
        return;
    }
    Console.WriteLine($"Started operation '{operation.Name}' at {operation.Metadata.CreateTime.ToDateTime():O}");

    Operation<Workflow, OperationMetadata> completed;
    try
    {
        // Poll for up to 90 seconds; the operation may take longer to complete.
        completed = operation.PollUntilCompleted(
            new PollSettings(Expiration.FromTimeout(TimeSpan.FromSeconds(90)), TimeSpan.FromSeconds(10)));
    }
    catch (TimeoutException)
    {
        // Leaving the wait does not cancel the operation; it continues
        // running on the server.
        Console.WriteLine(
            $"Operation '{operation.Name}' is still running; use this operation name to check progress later");
        return;
    }
    if (completed.IsFaulted)
    {
        if (completed.RpcMessage?.Error?.Code == (int)StatusCode.AlreadyExists)
        {
            Console.Error.WriteLine($"error: workflow '{workflowId}' already exists in '{locationId}'");
            return;
        }
        Console.Error.WriteLine($"error: operation '{operation.Name}' failed: {completed.Exception.Message}");
        return;
    }
    Workflow workflow = completed.Result;
    Console.WriteLine($"Created workflow {workflow.Name} (revision {workflow.RevisionId})");
}
