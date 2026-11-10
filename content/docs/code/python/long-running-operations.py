def create_workflow(project_id: str, location_id: str, workflow_id: str) -> None:
    with workflows_v1.WorkflowsClient() as client:
        try:
            # create_workflow starts a long-running operation. The workflow
            # isn't ready until the operation completes.
            operation = client.create_workflow(
                request={
                    "parent": f"projects/{project_id}/locations/{location_id}",
                    "workflow": {"source_contents": "- done:\n    return: ok\n"},
                    "workflow_id": workflow_id,
                }
            )
        except exceptions.AlreadyExists:
            print(
                f"error: workflow '{workflow_id}' already exists in '{location_id}'",
                file=sys.stderr,
            )
            return
        name = operation.operation.name
        print(f"Started operation '{name}' at {operation.metadata.create_time.isoformat()}")

        try:
            # Poll for up to 90 seconds; the operation may take longer to complete.
            workflow = operation.result(timeout=90)
        except concurrent.futures.TimeoutError:
            # Leaving the wait does not cancel the operation; it continues
            # running on the server.
            print(
                f"Operation '{name}' is still running; "
                "use this operation name to check progress later"
            )
            return
        except exceptions.AlreadyExists:
            print(
                f"error: workflow '{workflow_id}' already exists in '{location_id}'",
                file=sys.stderr,
            )
            return
        except exceptions.GoogleAPICallError as err:
            print(f"error: operation '{name}' failed: {err}", file=sys.stderr)
            return
        print(f"Created workflow {workflow.name} (revision {workflow.revision_id})")
