def create_workflow project_id:, location_id:, workflow_id:
  client = Google::Cloud::Workflows::V1::Workflows::Client.new
  begin
    # create_workflow starts a long-running operation. The workflow isn't
    # ready until the operation completes.
    operation = client.create_workflow(
      parent:      client.location_path(project: project_id, location: location_id),
      workflow:    { source_contents: "- done:\n    return: ok\n" },
      workflow_id: workflow_id
    )
  rescue Google::Cloud::AlreadyExistsError
    warn "error: workflow '#{workflow_id}' already exists in '#{location_id}'"
    return
  end
  puts "Started operation '#{operation.name}' at #{operation.metadata.create_time.to_time.utc}"

  # Poll for up to 90 seconds; the operation may take longer to complete.
  operation.wait_until_done! retry_policy: { timeout: 90 }
  unless operation.done?
    # Leaving the wait does not cancel the operation; it continues running
    # on the server.
    puts "Operation '#{operation.name}' is still running; use this operation name to check progress later"
    return
  end
  if operation.error?
    if operation.error.code == GRPC::Core::StatusCodes::ALREADY_EXISTS
      warn "error: workflow '#{workflow_id}' already exists in '#{location_id}'"
      return
    end
    warn "error: operation '#{operation.name}' failed: #{operation.error.message}"
    return
  end
  workflow = operation.response
  puts "Created workflow #{workflow.name} (revision #{workflow.revision_id})"
end
