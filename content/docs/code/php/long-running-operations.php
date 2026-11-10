function create_workflow(string $projectId, string $locationId, string $workflowId): void
{
    $client = new WorkflowsClient();
    $workflow = (new Workflow())->setSourceContents("- done:\n    return: ok\n");
    $request = (new CreateWorkflowRequest())
        ->setParent(WorkflowsClient::locationName($projectId, $locationId))
        ->setWorkflow($workflow)
        ->setWorkflowId($workflowId);
    try {
        // createWorkflow starts a long-running operation. The workflow isn't
        // ready until the operation completes.
        $operation = $client->createWorkflow($request);
    } catch (ApiException $e) {
        if ($e->getStatus() === ApiStatus::ALREADY_EXISTS) {
            fwrite(STDERR, sprintf(
                "error: workflow '%s' already exists in '%s'" . PHP_EOL,
                $workflowId,
                $locationId
            ));
            return;
        }
        throw $e;
    }
    printf(
        "Started operation '%s' at %s" . PHP_EOL,
        $operation->getName(),
        $operation->getMetadata()->getCreateTime()->toDateTime()->format(DATE_RFC3339)
    );

    // Poll for up to 90 seconds (90,000 ms); the operation may take longer
    // to complete.
    $operation->pollUntilComplete(['totalPollTimeoutMillis' => 90000]);
    if (!$operation->isDone()) {
        // Leaving the wait does not cancel the operation; it continues
        // running on the server.
        printf(
            "Operation '%s' is still running; use this operation name to check progress later" . PHP_EOL,
            $operation->getName()
        );
        return;
    }
    if ($operation->operationFailed()) {
        $error = $operation->getError();
        if ($error->getCode() === Code::ALREADY_EXISTS) {
            fwrite(STDERR, sprintf(
                "error: workflow '%s' already exists in '%s'" . PHP_EOL,
                $workflowId,
                $locationId
            ));
            return;
        }
        fwrite(STDERR, sprintf(
            "error: operation '%s' failed: %s" . PHP_EOL,
            $operation->getName(),
            $error->getMessage()
        ));
        return;
    }
    $result = $operation->getResult();
    printf('Created workflow %s (revision %s)' . PHP_EOL, $result->getName(), $result->getRevisionId());
}
