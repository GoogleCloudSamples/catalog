async function createWorkflow(projectId, locationId, workflowId) {
  const client = new v1.WorkflowsClient();
  let operation;
  try {
    let operation;
    try {
      // createWorkflow starts a long-running operation. The workflow isn't
      // ready until the operation completes.
      [operation] = await client.createWorkflow({
        parent: client.locationPath(projectId, locationId),
        workflow: {sourceContents: '- done:\n    return: ok\n'},
        workflowId,
      });
    } catch (err) {
      if (err.code === status.ALREADY_EXISTS) {
        console.error(`error: workflow '${workflowId}' already exists in '${locationId}'`);
        return;
      }
      throw err;
    }
    const createTime = new Date(Number(operation.metadata.createTime.seconds) * 1000);
    console.log(`Started operation '${operation.name}' at ${createTime.toISOString()}`);

    try {
      // Poll for up to 90 seconds (90,000 ms); the operation may take longer
      // to complete.
      operation.backoffSettings.totalTimeoutMillis = 90000;
      const [workflow] = await operation.promise();
      console.log(`Created workflow ${workflow.name} (revision ${workflow.revisionId})`);
    } catch (err) {
      if (err.code === status.DEADLINE_EXCEEDED) {
        // Leaving the wait does not cancel the operation; it continues
        // running on the server.
        console.log(
          `Operation '${operation.name}' is still running; use this operation name to check progress later`,
        );
        return;
      }
      if (err.code === status.ALREADY_EXISTS) {
        console.error(`error: workflow '${workflowId}' already exists in '${locationId}'`);
        return;
      }
      console.error(`error: operation '${operation.name}' failed: ${err.message}`);
    }
  } finally {
    await client.close();
  }
}
